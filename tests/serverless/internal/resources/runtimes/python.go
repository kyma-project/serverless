package runtimes

import (
	"fmt"

	v1 "k8s.io/api/core/v1"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
)

func BasicPythonFunction(msg string, runtime serverlessv1alpha2.Runtime) serverlessv1alpha2.FunctionSpec {
	src := fmt.Sprintf(`import arrow
import sdk
def main():
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
import sdk
from flask import request

import requests
from opentelemetry.instrumentation.requests import RequestsInstrumentor


def main():
    print("event headers: ", dict(request.headers))
    print("event data: ", request.get_data(as_text=True))
    print("event method: ", request.method)
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
def main():
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
import sdk
from flask import request

event_data = {}


def main():
    print("event headers: ", dict(request.headers))
    print("event data: ", request.get_data(as_text=True))
    print("event method: ", request.method)
    global event_data

    if request.method == 'GET':
        event_type = request.args.get('type')
        if event_type is None:
            print("type is not specified, returning all event data: ", json.dumps(event_data))
            return json.dumps(event_data)
        source = request.args.get('source')
        runtime_events = event_data.get(source, {})
        saved_event = runtime_events.get(event_type, "")
        print("getting saved event from memory for type:", event_type, ", for source: ", source, ", returning: ", json.dumps(saved_event))
        return json.dumps(saved_event)

    elif request.method == 'POST':
        ce = sdk.get_cloud_event()
        ce_time = ce.get_time()
        stored = {
            'ce-type': ce.get_type(),
            'ce-source': str(ce.get_source()),
            'ce-specversion': ce.get_specversion(),
            'ce-id': ce.get_id(),
            'ce-time': ce_time.isoformat() if ce_time else '',
            'ce-datacontenttype': ce.get_datacontenttype() or '',
            'data': ce.get_data(),
        }
        event_data[str(ce.get_source())] = {
            ce.get_type(): stored,
        }
        print("saving CE headers in-memory, source: ", ce.get_source(), ", headers: ", event_data[str(ce.get_source())], ", returning: 201")
        print("current event_data: ", event_data)
        return "", 201

    print("Unexpected call, returning: 405")
    return "", 405
`

	return serverlessv1alpha2.FunctionSpec{
		Runtime: serverlessv1alpha2.Python314,
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
import sdk
from flask import request

import requests

event_data = {}

send_check_event_type = "send-check"

runtime = os.getenv("CE_SOURCE")


def main():
    print("event headers: ", dict(request.headers))
    print("event data: ", request.get_data(as_text=True))
    print("event method: ", request.method)
    global event_data

    if request.method == 'GET':
        event_type = request.args.get('type')
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

    ce = sdk.get_cloud_event()
    if ce is None:
        sdk.emit_cloud_event(send_check_event_type, runtime, request.get_json())
        print("publishing CE, type: ", send_check_event_type, ", source: ", runtime, ", data: ", request.get_json())
        return ""

    ce_time = ce.get_time()
    stored = {
        'ce-type': ce.get_type(),
        'ce-source': str(ce.get_source()),
        'ce-specversion': ce.get_specversion(),
        'ce-id': ce.get_id(),
        'ce-time': ce_time.isoformat() if ce_time else '',
        'ce-datacontenttype': ce.get_datacontenttype() or '',
        'data': ce.get_data(),
    }
    event_data[ce.get_type()] = stored
    print("saving received cloud event, type: ", ce.get_type(), " headers: ", event_data[ce.get_type()])
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
