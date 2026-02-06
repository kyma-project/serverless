package runtimes

import (
	"fmt"

	v1 "k8s.io/api/core/v1"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
)

func BasicPythonFunction(msg string, runtime serverlessv1alpha2.Runtime) serverlessv1alpha2.FunctionSpec {
	src := fmt.Sprintf(`import arrow 
def main(event, context):
	return "%s"`, msg)

	dpd := `requests==2.31.0
arrow==0.15.8`

	return serverlessv1alpha2.FunctionSpec{
		Runtime: runtime,
		Source: serverlessv1alpha2.Source{
			Inline: &serverlessv1alpha2.InlineSource{
				Source:       src,
				Dependencies: dpd,
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

func BasicTracingPythonFunction(runtime serverlessv1alpha2.Runtime, externalURL string) serverlessv1alpha2.FunctionSpec {

	// TODO: New Buildless Serverless cannot use deprecated lib with new (0.50b0) opentelemetry libs - https://github.com/kyma-project/serverless/issues/1211#issuecomment-2636352928
	//	dpd := `opentelemetry-instrumentation==0.43b0
	//opentelemetry-instrumentation-requests==0.43b0
	//requests>=2.31.0`

	src := fmt.Sprintf(`import json

import requests
from opentelemetry.instrumentation.requests import RequestsInstrumentor


def main(event, context):
    print("event headers: ", vars(event['extensions']['request'].headers))
    print("event data: ", vars(event['extensions']['request'].body))
    print("event method: ", event['extensions']['request'].method)
    RequestsInstrumentor().instrument()
    response = requests.get('%s', timeout=1)
    headers = response.request.headers
    tracingHeaders = {}
    for key, value in headers.items():
        if key.startswith("x-b3") or key.startswith("traceparent"):
            tracingHeaders[key] = value
    print("response: ", json.dumps(tracingHeaders))
    return json.dumps(tracingHeaders)`, externalURL)

	return serverlessv1alpha2.FunctionSpec{
		Runtime: runtime,
		Source: serverlessv1alpha2.Source{
			Inline: &serverlessv1alpha2.InlineSource{
				//Dependencies: dpd,
				Source: src,
			},
		},
	}
}

func BasicPythonFunctionWithCustomDependency(msg string, runtime serverlessv1alpha2.Runtime) serverlessv1alpha2.FunctionSpec {
	src := fmt.Sprintf(
		`import arrow
def main(event, context):
	return "%s"`, msg)

	dpd := `requests==2.31.0
arrow==0.15.8
kyma-pypi-test==1.0.0`

	return serverlessv1alpha2.FunctionSpec{
		Runtime: runtime,
		Source: serverlessv1alpha2.Source{
			Inline: &serverlessv1alpha2.InlineSource{
				Source:       src,
				Dependencies: dpd,
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

func PythonPublisherProxyMock() serverlessv1alpha2.FunctionSpec {
	dpd := ``

	src := `import json

import bottle

event_data = {}


def main(event, context):
    print("event headers: ", vars(event['extensions']['request'].headers))
    print("event data: ", vars(event['extensions']['request'].body))
    print("event method: ", event['extensions']['request'].method)
    global event_data
    req = event.ceHeaders['extensions']['request']

    if req.method == 'GET':
        event_type = req.query.get(key='type')
        if event_type is None:
            print("type is not specified, returning all event data: ", json.dumps(event_data))
            return json.dumps(event_data)
        source = req.query.get(key='source')
        runtime_events = event_data.get(source, {})
        saved_event = runtime_events.get(event_type, "")
        print("getting saved event from memory for type:", event_type, ", for source: ", source, ", returning: ", json.dumps(saved_event))
        return json.dumps(saved_event)

    elif req.method == 'POST':
        event_ce_headers = event.ceHeaders
        event_ce_headers.pop('extensions')
        event_data[str(event_ce_headers['ce-source'])] = {
            event_ce_headers['ce-type']: event_ce_headers
        }
        print("saving CE headers in-memory, source: ", event_ce_headers['ce-source'], ", headers: ", event_data[str(event_ce_headers['ce-source'])], ", returning: 201")
        print("current event_data: ", event_data)
        return bottle.HTTPResponse(status=201)

    print("Unexpected call, returning: 405")
    return bottle.HTTPResponse(status=405)
`

	return serverlessv1alpha2.FunctionSpec{
		Runtime: serverlessv1alpha2.Python312,
		Source: serverlessv1alpha2.Source{
			Inline: &serverlessv1alpha2.InlineSource{
				Source:       src,
				Dependencies: dpd,
			},
		},
		Env: []v1.EnvVar{},
		Labels: map[string]string{
			"app.kubernetes.io/name": "eventing-publisher-proxy",
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

func PythonCloudEvent(runtime serverlessv1alpha2.Runtime) serverlessv1alpha2.FunctionSpec {
	dpd := ``

	src := `import json
import os

import requests

event_data = {}

send_check_event_type = "send-check"

runtime = os.getenv("CE_SOURCE")


def main(event, context):
    print("event headers: ", vars(event['extensions']['request'].headers))
    print("event data: ", vars(event['extensions']['request'].body))
    print("event method: ", event['extensions']['request'].method)
    global event_data
    req = event.ceHeaders['extensions']['request']
    
    if req.method == 'GET':
        event_type = req.query.get(key='type')
        if event_type == send_check_event_type:
            publisher_proxy = os.getenv("PUBLISHER_PROXY_ADDRESS")
            resp = requests.get(publisher_proxy, params={
                "type": event_type,
				"source": runtime
            })
            print("getting saved events from publisher proxy, type: ", send_check_event_type, ", source: ", runtime, ", returning: ", resp.json())
            return resp.json()
        
        saved_event = event_data.get(event_type, {})
        print("getting saved event from memory for type: ", event_type, ", returning: ", json.dumps(saved_event))
        return json.dumps(saved_event)
    
    if 'ce-type' not in event.ceHeaders:
        event.emitCloudEvent(send_check_event_type, runtime, req.json, {'eventtypeversion': 'v1alpha2'})
        print("publishing CE, type: ", send_check_event_type, ", source: ", runtime, ", data: ", req.json, ", attr: ", {'eventtypeversion': 'v1alpha2'})
        return ""

    event_ce_headers = event.ceHeaders
    event_ce_headers.pop('extensions')
    
    event_data[event_ce_headers['ce-type']] = event_ce_headers
    print("saving received cloud event, type: ", event_ce_headers['ce-type'], " headers: ", event_data[event_ce_headers['ce-type']])
    return ""
`

	return serverlessv1alpha2.FunctionSpec{
		Runtime: runtime,
		Source: serverlessv1alpha2.Source{
			Inline: &serverlessv1alpha2.InlineSource{
				Source:       src,
				Dependencies: dpd,
			},
		},
		Env: []v1.EnvVar{
			{
				Name:  "CE_SOURCE",
				Value: string(runtime),
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
