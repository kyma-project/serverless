package runtimes

import (
	"fmt"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/serverless/pkg/apis/serverless/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
)

func BasicNodeJSFunction(msg string, rtm serverlessv1alpha2.Runtime) serverlessv1alpha2.FunctionSpec {
	return serverlessv1alpha2.FunctionSpec{
		Runtime: rtm,
		Source: serverlessv1alpha2.Source{
			Inline: &serverlessv1alpha2.InlineSource{
				Source:       fmt.Sprintf(`module.exports = { main: function(event, context) { return "%s" } }`, msg),
				Dependencies: `{ "name": "hellobasic", "version": "0.0.1", "dependencies": {} }`,
			},
		},
		ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
			Function: &serverlessv1alpha2.ResourceRequirements{
				Profile: "M",
			},
			Build: &serverlessv1alpha2.ResourceRequirements{
				Profile: "fast",
			},
		},
	}
}

func BasicTracingNodeFunction(rtm serverlessv1alpha2.Runtime, externalSvcURL string) serverlessv1alpha2.FunctionSpec {
	dpd := `{
  "name": "sanitise-fn",
  "version": "0.0.1",
  "dependencies": {
    "axios":"0.26.1"
  }
}`
	src := fmt.Sprintf(`const axios = require("axios")


module.exports = {
    main: async function (event, context) {
        console.log("event: ", event)
        let resp = await axios("%s",{timeout: 1000});
        let interceptedHeaders = resp.request._header
        let tracingHeaders = getTracingHeaders(interceptedHeaders)
        console.log("return: ", JSON.stringify(tracingHeaders, null, 4))
        return tracingHeaders
    }
}

function getTracingHeaders(textHeaders) {
    tracingHeaders = textHeaders.split('\n')
        .filter(val => {
            let out = val.split(":")
            return out.length === 2;
        })
        .map(item => {
            let stringHeader = item.split(":")
            return {
                key: stringHeader[0],
                value: stringHeader[1]
            }
        })
        .filter(item => {
            return item.key.startsWith("x-b3") || item.key.startsWith("traceparent");
        })
        .map(val => {
            return {
                [val.key]: val.value
            }
        })
        .reduce((prev, current) => {
            return Object.assign(prev, current)
        })
    return tracingHeaders
}`, externalSvcURL)
	return serverlessv1alpha2.FunctionSpec{
		Runtime: rtm,
		Source: serverlessv1alpha2.Source{
			Inline: &serverlessv1alpha2.InlineSource{
				Source:       src,
				Dependencies: dpd,
			},
		},
		ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
			Function: &serverlessv1alpha2.ResourceRequirements{
				Profile: "L",
			},
			Build: &serverlessv1alpha2.ResourceRequirements{
				Profile: "fast",
			},
		},
	}
}

func BasicNodeJSFunctionWithCustomDependency(msg string, rtm serverlessv1alpha2.Runtime) serverlessv1alpha2.FunctionSpec {
	return serverlessv1alpha2.FunctionSpec{
		Runtime: rtm,
		Source: serverlessv1alpha2.Source{
			Inline: &serverlessv1alpha2.InlineSource{
				Source:       fmt.Sprintf(`module.exports = { main: function(event, context) { return "%s" } }`, msg),
				Dependencies: `{ "name": "hellobasic", "version": "0.0.1", "dependencies": { "@kyma/kyma-npm-test": "1.0.0" } }`,
			},
		},
		ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
			Function: &serverlessv1alpha2.ResourceRequirements{
				Profile: "M",
			},
			Build: &serverlessv1alpha2.ResourceRequirements{
				Profile: "fast",
			},
		},
	}
}

func NodeJSFunctionWithEnvFromConfigMapAndSecret(configMapName, cmEnvKey, secretName, secretEnvKey string, rtm serverlessv1alpha2.Runtime) serverlessv1alpha2.FunctionSpec {
	mappedCmEnvKey := "CM_KEY"
	mappedSecretEnvKey := "SECRET_KEY"

	src := fmt.Sprintf(`module.exports = { main: function(event, context) { return process.env["%s"] + "-" + process.env["%s"]; } }`, mappedCmEnvKey, mappedSecretEnvKey)
	dpd := `{ "name": "hellowithconfigmapsecretenvs", "version": "0.0.1", "dependencies": { } }`

	return serverlessv1alpha2.FunctionSpec{
		Runtime: rtm,
		Source: serverlessv1alpha2.Source{
			Inline: &serverlessv1alpha2.InlineSource{
				Source:       src,
				Dependencies: dpd,
			},
		},
		Env: []corev1.EnvVar{
			{
				Name: mappedCmEnvKey,
				ValueFrom: &corev1.EnvVarSource{
					ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: configMapName,
						},
						Key: cmEnvKey,
					},
				},
			},
			{
				Name: mappedSecretEnvKey,
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: secretName,
						},
						Key: secretEnvKey,
					},
				},
			},
		},
		ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
			Function: &serverlessv1alpha2.ResourceRequirements{
				Profile: "M",
			},
			Build: &serverlessv1alpha2.ResourceRequirements{
				Profile: "fast",
			},
		},
	}
}

func NodeJSFunctionWithCloudEvent(rtm serverlessv1alpha2.Runtime) serverlessv1alpha2.FunctionSpec {
	src := `const process = require("process");
const axios = require('axios');

let cloudevent = {}

send_check_event_type = "send-check"

runtime = process.env.CE_SOURCE

module.exports = {
    main: async function (event, context) {
        console.log("event: ", event)
        switch (event.extensions.request.method) {
            case "POST":
                res = handlePost(event)
                return res
            case "GET":
                res = handleGet(event.extensions.request)
                return res
            default:
                event.extensions.response.statusCode = 405
                console.log("Unexpected call, return: 405")
                return ""
        }
    }
}

function handlePost(event) {
    if (!Object.keys(event).includes("ce-type")) {
        event.emitCloudEvent(send_check_event_type, runtime, event.data, {'eventtypeversion': 'v1alpha2'})
        console.log("publishing CE, type: ", send_check_event_type, ", source: ", runtime, ", data: ", event.data,  ", attr: {eventtypeversion: v1alpha2}")
        return ""
    }
    Object.keys(event).filter((val) => {
        return val.startsWith("ce-")
    }).forEach((item) => {
        cloudevent[item] = event[item]
    })
    cloudevent.data = event.data
    console.log("saving received cloud event, type: ", event["ce-type"], "data: ", cloudevent.data)
    return ""
}

async function handleGet(req) {
    if (req.query.type === 'send-check') {
        let data = {}
        let publisherProxy = process.env.PUBLISHER_PROXY_ADDRESS
        await axios.get(publisherProxy, {
            params: {
                type: req.query.type,
				source: runtime
            }
        }).then((res) => {
            data = res.data
        }).catch((error) => {
            data = error
        })
        console.log("getting saved events from publisher proxy, type: ", req.query.type, ", source: ", runtime, ", returning: ", JSON.stringify(data, null, 4))
        return data
    }

    console.log("getting saved event from memory for type: ", req.query.type, ", returning: ", JSON.stringify(cloudevent, null, 4))
    return cloudevent
}
`
	return serverlessv1alpha2.FunctionSpec{
		Runtime: rtm,
		Source: serverlessv1alpha2.Source{
			Inline: &serverlessv1alpha2.InlineSource{
				Source:       src,
				Dependencies: `{ "name": "cloudevent", "version": "0.0.1", "dependencies": {} }`,
			},
		},
		Env: []v1.EnvVar{
			{
				Name:  "CE_SOURCE",
				Value: string(rtm),
			},
		},
		ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
			Function: &serverlessv1alpha2.ResourceRequirements{
				Profile: "L",
			},
			Build: &serverlessv1alpha2.ResourceRequirements{
				Profile: "fast",
			},
		},
	}
}

func NodeJSFunctionUsingHanaClient(rtm serverlessv1alpha2.Runtime) serverlessv1alpha2.FunctionSpec {
	src := `var hana = require('@sap/hana-client');

	module.exports = {
		main: async function (event, context) {
			//this is fake
			var conn_params = {
				serverNode: "62e223c1-7de9-4c8a-bab6-411f70fdf925.hana.canary-eu10.hanacloud.ondemand.com:443",
				uid: "DBADMIN",
				pwd: "foo",
			  };
			   
			  var conn = hana.createConnection();
	
			  try {
				await conn.connect(conn_params)
				let result = await conn.exec('SELECT 1 AS "One" FROM DUMMY')
				return result;
			  } catch(err) {
				// it is expected to leave here. The purpose is to check if hana client returns a known error instead of crashing the whole container with SIGSEGV
				// HY000 means general error - https://stackoverflow.com/questions/7472884/what-is-sql-error-5-sqlstate-hy000-and-what-can-cause-this-error
				if(err.sqlState && err.sqlState=="HY000"){
                  return "OK";
                } 
				return "NOK";
			  }
		}
	}
`
	return serverlessv1alpha2.FunctionSpec{
		Runtime: rtm,
		Source: serverlessv1alpha2.Source{
			Inline: &serverlessv1alpha2.InlineSource{
				Source:       src,
				Dependencies: `{"name": "hana-client","version": "0.0.1","dependencies": { "@sap/hana-client": "latest"} }`,
			},
		},
		ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
			Function: &serverlessv1alpha2.ResourceRequirements{
				Profile: "M",
			},
			Build: &serverlessv1alpha2.ResourceRequirements{
				Profile: "fast",
			},
		},
	}
}
