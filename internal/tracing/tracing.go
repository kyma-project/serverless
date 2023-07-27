package tracing

import (
	"context"
	telemetryv1alpha1 "github.com/kyma-project/telemetry-manager/apis/telemetry/v1alpha1"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetTracePipeline(ctx context.Context, client client.Client) (*telemetryv1alpha1.TracePipelineList, error) {
	tracePipelines := &telemetryv1alpha1.TracePipelineList{}
	err := client.List(ctx, tracePipelines)
	if err != nil {
		return nil, errors.Wrap(err, "while listing tracing pipelines")
	}
	return tracePipelines, nil
}
