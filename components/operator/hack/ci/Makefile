OPERATOR_ROOT ?= ../..
PROJECT_ROOT ?= $(OPERATOR_ROOT)/../..
PROJECT_COMMON ?= $(OPERATOR_ROOT)/hack/common
IMG_DIRECTORY ?= "prod" # image directory ( prod / dev )
IMG_VERSION ?= "main" # image version ( release tag / commit hash / PR-<number> )
MODULE_VERSION ?= 0.0.0

ifndef MODULE_SHA
    MODULE_SHA = ${shell git rev-parse --short HEAD}
endif

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: module-build
module-build: ## Build the Module, push it to a registry and print it based on the commit hash.
	@make -C ${OPERATOR_ROOT} module-build \
		IMG=${IMG} \
		MODULE_REGISTRY=${MODULE_REGISTRY} \
		MODULE_VERSION=${MODULE_VERSION}-${MODULE_SHA}

	@echo "\n~~~~~~~~~~~~BEGINING OF MODULE TEMPLATE~~~~~~~~~~~~~~"
	@cat ${PROJECT_ROOT}/moduletemplate.yaml
	@echo "\n~~~~~~~~~~~~~~~END OF MODULE TEMPLATE~~~~~~~~~~~~~~~~"

##@ Tests

.PHONY: integration-test
integration-test: ## Run integration tests on self-prepared k3d cluster.
integration-test:
	cd ${PROJECT_ROOT}/tests/serverless && make serverless-integration serverless-contract-tests

.PHONY: gardener-integration-test
gardener-integration-test: ## Provision gardener cluster and run integration test on it.
	@make provision-gardener \
		run-without-lm-on-cluster \
		integration-test \
		deprovision-gardener || \
		(make deprovision-gardener && false)

.PHONY: k3d-lm-integration-test
k3d-lm-integration-test: ## Run integration tests on self-prepared k3d cluster with lifecycle-manager.
k3d-lm-integration-test: run-with-lifecycle-manager integration-test

.PHONY: k3d-integration-test
k3d-integration-test: ## Run integration tests on self-prepared k3d cluster.
k3d-integration-test: run-without-lifecycle-manager integration-test remove-serverless

.PHONY: integration-test-on-cluster
integration-test-on-cluster: ## Run integration tests on existing k3d cluster.
integration-test-on-cluster: run-without-lm-on-cluster integration-test remove-serverless

.PHONY: upgrade-test
upgrade-test: ## Installs Serverless from latest and upgrades to local
upgrade-test: \
	install-latest-serverless \
	run-without-lm-on-cluster \
	integration-test

.PHONY: k3d-upgrade-test
k3d-upgrade-test: ## Deploy K3D cluster and installs Serverless from latest and upgrades to local.
k3d-upgrade-test: \
	create-k3d \
	install-latest-serverless \
	run-without-lm-on-k3d-cluster \
	integration-test

.PHONY: k3d-k8s-compatibility-test
k3d-k8s-compatibility-test: ## K8s compatibility tests not implemented yet.
	@echo "k8s compatibility tests not implemented yet"

.PHONY: hyperscalers-compatibility-test
hyperscalers-compatibility-test: ## Hyperscalers compatibility tests not implemented yet.
	@echo "hyperscalers compatibility tests not implemented yet"

.PHONY: remove-serverless
remove-serverless: ## Remove Serverless CR
	kubectl delete serverless -n kyma-system default --timeout 2m || (kubectl get serverless -n kyma-system -oyaml && false)

### Internal Dependencies

.PHONY: create-k3d
create-k3d:
	@make -C ${PROJECT_COMMON} kyma create-k3d

.PHONY: install-latest-serverless
install-latest-serverless:
	kubectl create namespace kyma-system || true
	kubectl apply -f https://github.com/kyma-project/serverless-manager/releases/latest/download/serverless-operator.yaml
	kubectl apply -f https://github.com/kyma-project/serverless-manager/releases/latest/download/default-serverless-cr.yaml -n kyma-system
	@make -C ${PROJECT_COMMON} verify-serverless

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
	@make -C ${OPERATOR_ROOT} render-manifest \
		MODULE_VERSION=${MODULE_VERSION}

.PHONY: replace-chart-images
replace-chart-images:
	@IMG_VERSION=${IMG_VERSION} IMG_DIRECTORY=${IMG_DIRECTORY} \
 		make -C ${PROJECT_COMMON} replace-chart-images

.PHONY: run-without-lm-on-k3d-cluster
run-without-lm-on-k3d-cluster:
	@make -C ${PROJECT_COMMON} run-without-lm-on-k3d-cluster
