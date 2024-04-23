ifndef PROJECT_ROOT
$(error PROJECT_ROOT is undefined)
endif
include ${PROJECT_ROOT}/hack/tools.mk

##@ Gardener

GARDENER_INFRASTRUCTURE = az
HIBERNATION_HOUR=$(shell echo $$(( ( $(shell date +%H | sed s/^0//g) + 5 ) % 24 )))
GIT_COMMIT_SHA=$(shell git rev-parse --short=8 HEAD)
ifneq (,$(GARDENER_SA_PATH))
GARDENER_K8S_VERSION=$(shell kubectl --kubeconfig=${GARDENER_SA_PATH} get cloudprofiles.core.gardener.cloud ${GARDENER_INFRASTRUCTURE} -o=jsonpath='{.spec.kubernetes.versions[0].version}')
else
GARDENER_K8S_VERSION=1.29.3
endif
#Overwrite default kyma cli gardenlinux version because it's not supported.
GARDENER_LINUX_VERSION=1312.3.0

.PHONY: provision-gardener
provision-gardener: kyma ## Provision gardener cluster with latest k8s version
	${KYMA} provision gardener ${GARDENER_INFRASTRUCTURE} -c ${GARDENER_SA_PATH} -n test-${GIT_COMMIT_SHA} -p ${GARDENER_PROJECT} -s ${GARDENER_SECRET_NAME} -k ${GARDENER_K8S_VERSION}\
		--gardenlinux-version=$(GARDENER_LINUX_VERSION) --hibernation-start="00 ${HIBERNATION_HOUR} * * ?"

.PHONY: deprovision-gardener
deprovision-gardener: kyma ## Deprovision gardener cluster
	kubectl --kubeconfig=${GARDENER_SA_PATH} annotate shoot test-${GIT_COMMIT_SHA} confirmation.gardener.cloud/deletion=true
	kubectl --kubeconfig=${GARDENER_SA_PATH} delete shoot test-${GIT_COMMIT_SHA} --wait=false
