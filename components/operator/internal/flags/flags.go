package flags

import (
	"fmt"

	"github.com/kyma-project/manager-toolkit/installation/chart"
)

type ImageReplace func(string) *Builder

type Builder struct {
	chart.FlagsBuilder
}

func NewBuilder() *Builder {
	return &Builder{
		FlagsBuilder: chart.NewFlagsBuilder(),
	}
}

func (b *Builder) WithControllerConfiguration(requeueDuration, healthzLivenessTimeout string) *Builder {
	optionalFlags := []struct {
		key   string
		value string
	}{
		{"functionRequeueDuration", requeueDuration},
		{"healthzLivenessTimeout", healthzLivenessTimeout},
	}

	for _, flag := range optionalFlags {
		if flag.value != "" {
			fullPath := fmt.Sprintf("containers.manager.configuration.data.%s", flag.key)
			b.With(fullPath, flag.value)
		}
	}

	return b
}

func (b *Builder) WithOptionalDependencies(publisherURL, traceCollectorURL string) *Builder {
	b.With("containers.manager.configuration.data.functionTraceCollectorEndpoint", traceCollectorURL)
	b.With("containers.manager.configuration.data.functionPublisherProxyAddress", publisherURL)
	return b
}

func (b *Builder) WithDefaultPresetFlags(defaultRuntimePodPreset string) *Builder {
	if defaultRuntimePodPreset != "" {
		b.With("containers.manager.configuration.data.resourcesConfiguration.function.resources.defaultPreset", defaultRuntimePodPreset)
	}

	return b
}

func (b *Builder) WithManagedByLabel(managedBy string) *Builder {
	b.With("global.commonLabels.app\\.kubernetes\\.io/managed-by", managedBy)
	return b
}

func (b *Builder) WithLogLevel(logLevel string) *Builder {
	if logLevel != "" {
		b.With("containers.manager.logConfiguration.data.logLevel", logLevel)
	}

	return b
}

func (b *Builder) WithLogFormat(logFormat string) *Builder {
	if logFormat != "" {
		b.With("containers.manager.logConfiguration.data.logFormat", logFormat)
	}

	return b
}

// WithLogFormatRestartAnnotation sets the restart annotation value which triggers a pod restart when logFormat changes
func (b *Builder) WithLogFormatRestartAnnotation(logFormat string) *Builder {
	if logFormat != "" {
		b.With("containers.manager.logConfiguration.restartAnnotationValue", logFormat)
	}

	return b
}

func (b *Builder) WithFipsModeEnabled(enabled bool) *Builder {
	b.With("containers.manager.fipsModeEnabled", enabled)
	return b
}

func (b *Builder) WithImageFunctionController(image string) *Builder {
	b.With("global.images.function_controller", image)
	return b
}

func (b *Builder) WithImageFunctionInit(image string) *Builder {
	b.With("global.images.function_init", image)
	return b
}

func (b *Builder) WithImageFunctionRuntimeNodejs20(image string) *Builder {
	b.With("global.images.function_runtime_nodejs20", image)
	return b
}

func (b *Builder) WithImageFunctionRuntimeNodejs22(image string) *Builder {
	b.With("global.images.function_runtime_nodejs22", image)
	return b
}

func (b *Builder) WithImageFunctionRuntimeNodejs24(image string) *Builder {
	b.With("global.images.function_runtime_nodejs24", image)
	return b
}

func (b *Builder) WithImageFunctionRuntimePython312(image string) *Builder {
	b.With("global.images.function_runtime_python312", image)
	return b
}
