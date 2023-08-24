PROJECT_ROOT ?= ../..
PROJECT_COMMON ?= ${PROJECT_ROOT}/hack/common

MODULECHARTTEST ?= ${PROJECT_ROOT}/module-chart-test

ifndef MODULE_VERSION
    include ${PROJECT_ROOT}/.version
endif

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
	@cat ${PROJECT_ROOT}/moduletemplate.yaml
	@echo "\n~~~~~~~~~~~~~~~END OF MODULE TEMPLATE~~~~~~~~~~~~~~~~"

##@ Tests

.PHONY: integration-test
integration-test: module-chart-test  ## Run integration tests on self-prepared k3d cluster.
	cd ${MODULECHARTTEST} && make serverless-integration

.PHONY: gardener-integration-test
gardener-integration-test: ## Provision gardener cluster and run integration test on it.
gardener-integration-test:
	@make provision-gardener \
		run-without-lm-on-cluster \
		integration-test \
		deprovision-gardener || \
		(make deprovision-gardener && false)

.PHONY: k3d-lm-integration-test ## Run integration tests on self-prepared k3d cluster with lifecycle-manager.
k3d-lm-integration-test: run-with-lifecycle-manager integration-test

.PHONY: k3d-integration-test ## Run integration tests on self-prepared k3d cluster.
k3d-integration-test: run-without-lifecycle-manager integration-test remove-serverless

.PHONY: k3d-upgrade-test
k3d-upgrade-test: ## Installs Serverless from latest and upgrades to local.
k3d-upgrade-test: \
	install-latest-serverless \
	reinstall-serverless \
	integration-test

.PHONY: k3d-k8s-compatibility-test
k3d-k8s-compatibility-test: ## K8s compatibility tests not implemented yet.
	@echo "k8s compatibility tests not implemented yet"

.PHONY: hyperscalers-compatibility-test
hyperscalers-compatibility-test: ## Hyperscalers compatibility tests not implemented yet.
	@echo "hyperscalers compatibility tests not implemented yet"

.PHONY: module-chart-test
module-chart-test: ## Generate module-chart-test dir.
	@${PROJECT_ROOT}/hack/clone_dir_from_github.sh kyma tests/function-controller ${MODULECHARTTEST}

.PHONY: remove-serverless
remove-serverless: ## Remove Serverless CR
	kubectl delete serverless -n kyma-system default --timeout 2m || (kubectl get serverless -n kyma-system -oyaml && false)

### Internal Dependencies

.PHONY: install-latest-serverless
install-latest-serverless:
	@make -C ${PROJECT_COMMON} kyma create-k3d
	curl -LJ -s https://github.com/kyma-project/serverless-manager/releases/latest/download/moduletemplate.yaml > ${PROJECT_ROOT}/moduletemplate-latest.yaml
	@cat ${PROJECT_ROOT}/moduletemplate-latest.yaml | sed -e 's/enableInternal: true/enableInternal: false/g' > ${PROJECT_ROOT}/moduletemplate-k3d.yaml
	@make -C ${PROJECT_COMMON} install-kyma-with-lm patch-mod-mgr-role install-module-template enable-module verify-kyma

.PHONY: reinstall-serverless
reinstall-serverless:
	@make -C ${PROJECT_COMMON} kyma \
	  module-image \
	  module-build \
	  fix-template \
	  install-module-template
	
	# wait some time to make sure lm starts the reconciliation first
	sleep 5

	@make -C ${PROJECT_COMMON} verify-kyma

.PHONY: provision-gardener
provision-gardener:
	@make -C ${PROJECT_COMMON} provision-gardener

.PHONY: deprovision-gardener
deprovision-gardener:
	@make -C ${PROJECT_COMMON} deprovision-gardener

.PHONY: run-with-lifecycle-manager
run-with-lifecycle-manager:
	@make -C ${PROJECT_COMMON} run-with-lifecycle-manager

.PHONY: run-without-lifecycle-manager
run-without-lifecycle-manager:
	@make -C ${PROJECT_COMMON} run-without-lifecycle-manager

.PHONY: run-without-lm-on-cluster
run-without-lm-on-cluster:
	@make -C ${PROJECT_COMMON} run-without-lm-on-cluster

.PHONY: render-manifest
render-manifest:
	@make -C ${PROJECT_ROOT} render-manifest