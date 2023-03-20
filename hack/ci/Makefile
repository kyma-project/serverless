PROJECT_ROOT ?= ../..
PROJECT_COMMON ?= ${PROJECT_ROOT}/hack/common

MODULECHARTTEST ?= ${PROJECT_ROOT}/module-chart-test

include ${PROJECT_ROOT}/.env

ifndef MODULE_SHA
    MODULE_SHA = ${shell git rev-parse --short HEAD}
endif

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: module-build
module-build: ## Build the Module, push it to a registry and print it based on the commit hash.
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

.PHONY: integration-test
integration-test: module-chart-test  ## Run integration tests on self-prepared k3d cluster.
	cd ${MODULECHARTTEST} && make serverless-integration

.PHONY: k3d-lm-integration-test ## Run integration tests on self-prepared k3d cluster with lifecycle-manager.
k3d-lm-integration-test: run-with-lifecycle-manager integration-test

.PHONY: k3d-integration-test ## Run integration tests on self-prepared k3d cluster.
k3d-integration-test: run-without-lifecycle-manager integration-test

.PHONY: k3d-upgrade-test
k3d-upgrade-test: ## Upgrade tests not implemented yet.
	@echo "upgrade tests not implemented yet"

.PHONY: k3d-k8s-compatibility-test
k3d-k8s-compatibility-test: ## K8s compatibility tests not implemented yet.
	@echo "k8s compatibility tests not implemented yet"

.PHONY: hyperscalers-compatibility-test
hyperscalers-compatibility-test: ## Hyperscalers compatibility tests not implemented yet.
	@echo "hyperscalers compatibility tests not implemented yet"

.PHONY: module-chart-test
module-chart-test: $(MODULECHARTTEST) ## Generate module-chart-test dir.

$(MODULECHARTTEST):
	mkdir -p ${MODULECHARTTEST}
		git clone \
		--depth 1  \
		--filter=blob:none  \
		--no-checkout \
		https://github.com/kyma-project/kyma ${MODULECHARTTEST}
	git -C ${MODULECHARTTEST} sparse-checkout set tests/function-controller
	git -C ${MODULECHARTTEST} checkout main
	mv ${MODULECHARTTEST}/tests/function-controller/* ${MODULECHARTTEST}
	rm -rf ${MODULECHARTTEST}/.git
	rm -rf ${MODULECHARTTEST}/tests
	echo module-chart-test generated

### Internal Dependencies

.PHONY: run-with-lifecycle-manager
run-with-lifecycle-manager:
	@make -C ${PROJECT_COMMON} run-with-lifecycle-manager

.PHONY: run-without-lifecycle-manager
run-without-lifecycle-manager:
	@make -C ${PROJECT_COMMON} run-without-lifecycle-manager
