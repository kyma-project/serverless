PROJECT_ROOT ?= ../..
TEMPLATES_DIR ?= ${PROJECT_ROOT}/config/moduletemplates

.PHONY: kyma
kyma:
	@make -C ${PROJECT_ROOT} kyma

.PHONY: module-build
module-build:
	@make -C ${PROJECT_ROOT} module-build \
		IMG=${IMG} \
		MODULE_REGISTRY=${MODULE_REGISTRY} \
		MODULE_VERSION=${MODULE_VERSION}

.PHONY: module-template-print
module-template-print:
	@cat ${TEMPLATES_DIR}/template.yaml

.PHONY: all
all: module-build module-template-print
