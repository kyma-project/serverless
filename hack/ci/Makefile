PROJECT_ROOT ?= ../..

include ${PROJECT_ROOT}/.env

ifndef MODULE_SHA
    MODULE_SHA = ${shell git rev-parse --short HEAD}
endif

.PHONY: kyma
kyma:
	@make -C ${PROJECT_ROOT} kyma

.PHONY: module-build
module-build:
	@make -C ${PROJECT_ROOT} module-build \
		IMG=${IMG} \
		MODULE_REGISTRY=${MODULE_REGISTRY} \
		MODULE_VERSION=${MODULE_VERSION}-${MODULE_SHA}

.PHONY: module-template-print
module-template-print:
	@cat ${PROJECT_ROOT}/template.yaml

.PHONY: all
all: module-build module-template-print

.PHONY: helm-install
helm-install:
	@make -C ${PROJECT_ROOT} helm-install

.PHONY: integration-test
integration-test: helm-install
	@helm install k3s-tests ${PROJECT_ROOT}/module-chart/charts/k3s-tests --values ${PROJECT_ROOT}/module-chart/values.yaml
	@kubectl wait --for=condition=ready pod -l job-name=k3s-serverless-test --timeout=30s
	kubectl logs -l job-name=k3s-serverless-test -f
	@kubectl wait --for=condition=complete job k3s-serverless-test --timeout=30s
