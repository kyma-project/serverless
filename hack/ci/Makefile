PROJECT_ROOT ?= ../..
PROJECT_COMMON ?= ${PROJECT_ROOT}/hack/common

include ${PROJECT_ROOT}/.env

ifndef MODULE_SHA
    MODULE_SHA = ${shell git rev-parse --short HEAD}
endif

# TODO: add help, comments and sections

.PHONY: kyma
kyma:
#	TODO: remove this target
	@make -C ${PROJECT_ROOT} kyma

.PHONY: module-build
module-build:
	@make -C ${PROJECT_ROOT} module-build \
		IMG=${IMG} \
		MODULE_REGISTRY=${MODULE_REGISTRY} \
		MODULE_VERSION=${MODULE_VERSION}-${MODULE_SHA}

.PHONY: module-template-print
module-template-print:
	@echo "\n~~~~~~~~~~~~BEGINING OF MODULE TEMPLATE~~~~~~~~~~~~~~"
	@cat ${PROJECT_ROOT}/template.yaml
	@echo "\n~~~~~~~~~~~~~~~END OF MODULE TEMPLATE~~~~~~~~~~~~~~~~"

.PHONY: all
all: module-build module-template-print

.PHONY: helm-install
helm-install:
	@make -C ${PROJECT_ROOT} helm-install

.PHONY: run
run:
#	TODO: rename it
	@make -C ${PROJECT_COMMON} run

.PHONY: integration-test
integration-test: helm-install run
	@helm install k3s-tests ${PROJECT_ROOT}/module-chart/charts/k3s-tests --values ${PROJECT_ROOT}/module-chart/values.yaml
	@kubectl wait --for=condition=ready pod -l job-name=k3s-serverless-test --timeout=30s
	kubectl logs -l job-name=k3s-serverless-test -f
	@kubectl wait --for=condition=complete job k3s-serverless-test --timeout=30s
