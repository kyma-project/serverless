# local variables
OPERATOR_ROOT = ../..
PROJECT_ROOT = $(OPERATOR_ROOT)/../..
PROJECT_COMMON = $(OPERATOR_ROOT)/hack/common

include ${PROJECT_ROOT}/hack/tools/help.Makefile
include ${PROJECT_ROOT}/hack/tools/gardener.Makefile

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

.PHONY: run
run:
	@make -C ${PROJECT_COMMON} run

.PHONY: run-on-cluster
run-on-cluster:
	@make -C ${PROJECT_COMMON} run-on-cluster

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
