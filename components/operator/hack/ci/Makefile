# local variables
OPERATOR_ROOT = ../..
PROJECT_ROOT = $(OPERATOR_ROOT)/../..
PROJECT_COMMON = $(OPERATOR_ROOT)/hack/common
include ${PROJECT_ROOT}/hack/tools/help.Makefile

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

##@ Gardener

HIBERNATION_HOUR=$(shell echo $$(( ( $(shell date +%H | sed s/^0//g) + 5 ) % 24 )))
GIT_COMMIT_SHA=$(shell git rev-parse --short=8 HEAD)
ifneq (,$(GARDENER_SA_PATH))
GARDENER_K8S_VERSION=$(shell kubectl --kubeconfig=${GARDENER_SA_PATH} get cloudprofiles.core.gardener.cloud ${GARDENER_INFRASTRUCTURE} -o=jsonpath='{.spec.kubernetes.versions[0].version}')
else
GARDENER_K8S_VERSION=1.27.4
endif

.PHONY: provision-gardener
provision-gardener: kyma ## Provision gardener cluster with latest k8s version
	${KYMA} provision gardener ${GARDENER_INFRASTRUCTURE} -c ${GARDENER_SA_PATH} -n test-${GIT_COMMIT_SHA} -p ${GARDENER_PROJECT} -s ${GARDENER_SECRET_NAME} -k ${GARDENER_K8S_VERSION}\
		--hibernation-start="00 ${HIBERNATION_HOUR} * * ?"

.PHONY: deprovision-gardener
deprovision-gardener: kyma ## Deprovision gardener cluster
	kubectl --kubeconfig=${GARDENER_SA_PATH} annotate shoot test-${GIT_COMMIT_SHA} confirmation.gardener.cloud/deletion=true
	kubectl --kubeconfig=${GARDENER_SA_PATH} delete shoot test-${GIT_COMMIT_SHA} --wait=false

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
