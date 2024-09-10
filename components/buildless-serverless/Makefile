# local variables
PROJECT_ROOT = ../..

include ${PROJECT_ROOT}/hack/tools.mk

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="$(PROJECT_ROOT)/hack/boilerplate.go.txt" paths="./..."
