ifndef PROJECT_ROOT
$(error PROJECT_ROOT is undefined)
endif
include ${PROJECT_ROOT}/hack/tools.mk

##@ Gardener

GARDENER_INFRASTRUCTURE = az
HIBERNATION_HOUR=$(shell echo $$(( ( $(shell date +%H | sed s/^0//g) + 5 ) % 24 )))
GIT_COMMIT_SHA=$(shell git rev-parse --short=8 HEAD)
SHOOT=test-${GIT_COMMIT_SHA}
ifneq (,$(GARDENER_SA_PATH))
GARDENER_K8S_VERSION=$(shell kubectl --kubeconfig=${GARDENER_SA_PATH} get cloudprofiles.core.gardener.cloud ${GARDENER_INFRASTRUCTURE} -o=jsonpath='{.spec.kubernetes.versions[0].version}')
else
GARDENER_K8S_VERSION=1.29.3
endif

.PHONY: provision-gardener
provision-gardener: ## Provision gardener cluster with latest k8s version
	PROJECT_ROOT=${PROJECT_ROOT} \
		GARDENER_SA_PATH=${GARDENER_SA_PATH} \
		SHOOT=${SHOOT} PROJECT=${GARDENER_PROJECT} \
		GARDENER_K8S_VERSION=${GARDENER_K8S_VERSION} \
		SECRET=${GARDENER_SECRET_NAME} \
		${PROJECT_ROOT}/hack/provision_gardener.sh

.PHONY: deprovision-gardener
deprovision-gardener: kyma ## Deprovision gardener cluster
	kubectl --kubeconfig=${GARDENER_SA_PATH} annotate shoot test-${GIT_COMMIT_SHA} confirmation.gardener.cloud/deletion=true
	kubectl --kubeconfig=${GARDENER_SA_PATH} delete shoot test-${GIT_COMMIT_SHA} --wait=false
