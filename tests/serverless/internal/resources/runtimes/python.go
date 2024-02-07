package runtimes

import (
	"fmt"

	v1 "k8s.io/api/core/v1"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/serverless/pkg/apis/serverless/v1alpha2"
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

	dpd := `opentelemetry-instrumentation==0.43b0
opentelemetry-instrumentation-requests==0.43b0
requests>=2.31.0`

	src := fmt.Sprintf(`import json

import requests
from opentelemetry.instrumentation.requests import RequestsInstrumentor


def main(event, context):
    print("event headers: ", vars(event['extensions']['request'].headers))
    print("event data: ", vars(event['extensions']['request'].body))
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
				Source:       src,
				Dependencies: dpd,
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
    global event_data
    req = event.ceHeaders['extensions']['request']

    if req.method == 'GET':
        event_type = req.query.get(key='type')
        if event_type is None:
            print("None type, return: ", json.dumps(event_data))
            return json.dumps(event_data)
        remote_addr = req.query.get(key='address', default=req.remote_addr)
        runtime_events = event_data.get(remote_addr, {})
        saved_event = runtime_events.get(event_type, "")
        print("GET, return: ", json.dumps(saved_event))
        return json.dumps(saved_event)
    elif req.method == 'POST':
        event_ce_headers = event.ceHeaders
        event_ce_headers.pop('extensions')
        event_data[str(req.remote_addr)] = {
            event_ce_headers['ce-type']: event_ce_headers
        }
        print("POST, return: 201")
        return bottle.HTTPResponse(status=201)

    print("Error, return: 405")
    return bottle.HTTPResponse(status=405)
`

	return serverlessv1alpha2.FunctionSpec{
		Runtime: serverlessv1alpha2.Python39,
		Source: serverlessv1alpha2.Source{
			Inline: &serverlessv1alpha2.InlineSource{
				Source:       src,
				Dependencies: dpd,
			},
		},
		Env: []v1.EnvVar{
			{
				Name:  "PUBLISHER_PROXY_ADDRESS",
				Value: "localhost:8080",
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

func PythonCloudEvent(runtime serverlessv1alpha2.Runtime) serverlessv1alpha2.FunctionSpec {
	dpd := ``

	src := `import json
import os

import requests

event_data = {}

send_check_event_type = "send-check"


def main(event, context):
    print("event headers: ", vars(event['extensions']['request'].headers))
    print("event data: ", vars(event['extensions']['request'].body))
    global event_data
    req = event.ceHeaders['extensions']['request']
    if req.method == 'GET':
        event_type = req.query.get(key='type')
        if event_type == send_check_event_type:
            publisher_proxy = os.getenv("PUBLISHER_PROXY_ADDRESS")
            resp = requests.get(publisher_proxy, params={
                "type": event_type
            })
            print("send-check, return: ", resp.json())
            return resp.json()

        saved_event = event_data.get(event_type, {})
        print("GET, return: ", json.dumps(saved_event))
        return json.dumps(saved_event)

    if 'ce-type' not in event.ceHeaders:
        event.emitCloudEvent(send_check_event_type, 'function', req.json, {'eventtypeversion': 'v1alpha2'})
        print("publish CE: type - ", send_check_event_type, ", source - function, data - ", req.json, ", attr - ", {'eventtypeversion': 'v1alpha2'})
        return ""
    event_ce_headers = event.ceHeaders
    event_ce_headers.pop('extensions')

    event_data[event_ce_headers['ce-type']] = event_ce_headers
    print("set ce headers and stop")
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
				Name:  "PUBLISHER_PROXY_ADDRESS",
				Value: "localhost:8080",
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
