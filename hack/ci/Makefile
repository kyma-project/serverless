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

.PHONY: integration-test
integration-test: helm-install run ## Run integration tests on self-prepared k3d cluster
	@helm install k3s-tests ${PROJECT_ROOT}/module-chart/charts/k3s-tests --values ${PROJECT_ROOT}/module-chart/values.yaml
	@kubectl wait --for=condition=ready pod -l job-name=k3s-serverless-test --timeout=30s
	kubectl logs -l job-name=k3s-serverless-test -f
	@kubectl wait --for=condition=complete job k3s-serverless-test --timeout=30s

.PHONY: all
# TODO: remove it after switching this target on the test-infra side
all: module-build

.PHONY: helm-install
helm-install:
	@make -C ${PROJECT_ROOT} helm-install

.PHONY: run
run:
#	TODO: rename it
	@make -C ${PROJECT_COMMON} run
