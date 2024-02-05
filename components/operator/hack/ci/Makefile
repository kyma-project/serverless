OPERATOR_ROOT ?= ../..
PROJECT_ROOT ?= $(OPERATOR_ROOT)/../..
PROJECT_COMMON ?= $(OPERATOR_ROOT)/hack/common
IMG_DIRECTORY ?= "prod" # image directory ( prod / dev )
IMG_VERSION ?= "main" # image version ( release tag / commit hash / PR-<number> )

ifndef MODULE_SHA
    MODULE_SHA = ${shell git rev-parse --short HEAD}
endif

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Tests

.PHONY: integration-test
integration-test: ## Run integration tests on self-prepared k3d cluster.
integration-test:
	cd ${PROJECT_ROOT}/tests/serverless && make serverless-integration serverless-contract-tests

.PHONY: gardener-integration-test
gardener-integration-test: ## Provision gardener cluster and run integration test on it.
	@make provision-gardener \
		run-on-cluster \
		integration-test \
		deprovision-gardener || \
		(make deprovision-gardener && false)

.PHONY: integration-test-on-cluster
integration-test-on-cluster: ## Run integration tests on existing k3d cluster.
integration-test-on-cluster: run-on-cluster integration-test remove-serverless

.PHONY: upgrade-test
upgrade-test: ## Installs Serverless from latest and upgrades to local
upgrade-test: \
	install-latest-serverless \
	run-on-cluster \
	integration-test

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

.PHONY: run
run:
	@make -C ${PROJECT_COMMON} run

.PHONY: run-on-cluster
run-on-cluster:
	@make -C ${PROJECT_COMMON} run-on-cluster

.PHONY: replace-only-main-chart-images
replace-only-main-chart-images:
	yq -i "(.global.images[] | select(.version == \"main\") | .directory) = \"${IMG_DIRECTORY}\"" ${PROJECT_ROOT}/config/serverless/values.yaml
	yq -i "(.global.images[] | select(.version == \"main\") | .version) = \"${IMG_VERSION}\"" ${PROJECT_ROOT}/config/serverless/values.yaml
	@echo "==== Local Changes ===="
	yq '.global.images' ${PROJECT_ROOT}/config/serverless/values.yaml
	@echo "==== End of Local Changes ===="

.PHONY: replace-function-chart-images
replace-function-chart-images:
	yq -i "(.global.images[] | select(key == \"function_*\") | .directory) = \"${IMG_DIRECTORY}\"" ${PROJECT_ROOT}/config/serverless/values.yaml
	yq -i "(.global.images[] | select(key == \"function_*\") | .version) = \"${IMG_VERSION}\"" ${PROJECT_ROOT}/config/serverless/values.yaml
	@echo "==== Local Changes ===="
	yq '.global.images' ${PROJECT_ROOT}/config/serverless/values.yaml
	@echo "==== End of Local Changes ===="
