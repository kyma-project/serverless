PROJECT_ROOT ?= ../..
PROJECT_COMMON ?= ${PROJECT_ROOT}/hack/common

include ${PROJECT_ROOT}/.env

ifndef MODULE_SHA
    MODULE_SHA = ${shell git rev-parse --short HEAD}
endif

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: module-build
module-build: ## Build the Module, push it to a registry and print it based on the commit hash
	@make -C ${PROJECT_ROOT} module-build \
		IMG=${IMG} \
		MODULE_REGISTRY=${MODULE_REGISTRY} \
		MODULE_VERSION=${MODULE_VERSION}-${MODULE_SHA}

	@echo "\n~~~~~~~~~~~~BEGINING OF MODULE TEMPLATE~~~~~~~~~~~~~~"
	@cat ${PROJECT_ROOT}/template.yaml
	@echo "\n~~~~~~~~~~~~~~~END OF MODULE TEMPLATE~~~~~~~~~~~~~~~~"

.PHONY: all
# TODO: remove it after switching this target on the test-infra side to the `module-build`
all: module-build

##@ Tests

HELM ?= ${PROJECT_ROOT}/bin/helm

.PHONY: integration-test
integration-test: helm ## Run integration tests on self-prepared k3d cluster
	@$(HELM) install k3s-tests ${PROJECT_ROOT}/module-chart/charts/k3s-tests --values ${PROJECT_ROOT}/module-chart/values.yaml
	@kubectl wait --for=condition=ready pod -l job-name=k3s-serverless-test --timeout=30s
	kubectl logs -l job-name=k3s-serverless-test -f
	@kubectl wait --for=condition=complete job k3s-serverless-test --timeout=30s

.PHONY: k3d-lm-integration-test ## Run integration tests on self-prepared k3d cluster with lifecycle-manager
k3d-lm-integration-test: run-with-lifecycle-manager integration-test

.PHONY: k3d-integration-test ## Run integration tests on self-prepared k3d cluster
k3d-integration-test: run-without-lifecycle-manager integration-test

.PHONY: helm
helm:
	@make -C ${PROJECT_ROOT} helm

.PHONY: run-with-lifecycle-manager
run-with-lifecycle-manager:
	@make -C ${PROJECT_COMMON} run-with-lifecycle-manager

.PHONY: run-without-lifecycle-manager
run-without-lifecycle-manager:
	@make -C ${PROJECT_COMMON} run-without-lifecycle-manager
