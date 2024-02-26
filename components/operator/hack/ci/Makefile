# local variables
OPERATOR_ROOT = ../..
PROJECT_ROOT = $(OPERATOR_ROOT)/../..
PROJECT_COMMON = $(OPERATOR_ROOT)/hack/common

include ${PROJECT_ROOT}/hack/tools/help.Makefile

.PHONY: replace-only-main-chart-images
replace-only-main-chart-images: check-img-envs
	yq -i "(.global.images[] | select(.version == \"main\") | .directory) = \"${IMG_DIRECTORY}\"" ${PROJECT_ROOT}/config/serverless/values.yaml
	yq -i "(.global.images[] | select(.version == \"main\") | .version) = \"${IMG_VERSION}\"" ${PROJECT_ROOT}/config/serverless/values.yaml
	@echo "==== Local Changes ===="
	yq '.global.images' ${PROJECT_ROOT}/config/serverless/values.yaml
	@echo "==== End of Local Changes ===="

.PHONY: replace-function-chart-images
replace-function-chart-images: check-img-envs
	yq -i "(.global.images[] | select(key == \"function_*\") | .directory) = \"${IMG_DIRECTORY}\"" ${PROJECT_ROOT}/config/serverless/values.yaml
	yq -i "(.global.images[] | select(key == \"function_*\") | .version) = \"${IMG_VERSION}\"" ${PROJECT_ROOT}/config/serverless/values.yaml
	@echo "==== Local Changes ===="
	yq '.global.images' ${PROJECT_ROOT}/config/serverless/values.yaml
	@echo "==== End of Local Changes ===="

.PHONY: check-img-envs
check-img-envs:
ifndef IMG_DIRECTORY
	$(error IMG_DIRECTORY is undefined)
endif
ifndef IMG_VERSION
	$(error IMG_VERSION is undefined)
endif
