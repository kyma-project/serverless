OPERATOR_ROOT ?= ../..
PROJECT_ROOT ?= $(OPERATOR_ROOT)/../..

MODULE_VERSION ?= 0.0.0

CLUSTER_NAME ?= kyma
KYMA_PROFILE ?= production
KYMA_SOURCE ?= main
REGISTRY_PORT ?= 5001
REGISTRY_NAME ?= ${CLUSTER_NAME}-registry

OPERATOR_IMAGE_NAME ?= serverless-operator-dev-local
OPERATOR_IMAGE_TAG ?= $(MODULE_VERSION)

GARDENER_INFRASTRUCTURE ?= az

KYMA ?= ${PROJECT_ROOT}/bin/kyma-unstable

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: run-with-lifecycle-manager
run-with-lifecycle-manager: ## Create k3d cluster and deploy module with the lifecycle-manager.
run-with-lifecycle-manager: kyma \
	create-k3d \
	module-image-dev \
	module-build \
	fix-template \
	install-kyma-with-lm \
	patch-mod-mgr-role \
	install-module-template \
	enable-module \
	verify-kyma

.PHONY: run-without-lifecycle-manager
run-without-lifecycle-manager: ## Create k3d cluster and deploy module without the lifecycle-manager.
run-without-lifecycle-manager: kyma \
	create-k3d \
	module-image-dev \
	deploy \
	apply-serverless \
	verify-serverless

.PHONY: run-without-lm-on-cluster
run-without-lm-on-cluster: ## Deploy module without the lifecycle-manager on existing cluster.
run-without-lm-on-cluster: \
	deploy-cluster \
	apply-serverless \
	verify-serverless

.PHONY: run-without-lm-on-k3d-cluster
run-without-lm-on-k3d-cluster: ## Build and deploy module without the lifecycle-manager on existing k3d cluster.
run-without-lm-on-k3d-cluster: \
	module-image-dev \
	deploy \
	apply-serverless \
	verify-serverless

##@ K3D

.PHONY: create-k3d
create-k3d: ## Create k3d with kyma CRDs.
	${KYMA} provision k3d --registry-port ${REGISTRY_PORT} --name ${CLUSTER_NAME} --ci -p 6080:8080@loadbalancer -p 6433:8433@loadbalancer
	kubectl create namespace kyma-system

.PHONY: delete-k3d
delete-k3d: delete-k3d-cluster delete-k3d-registry ## Delete k3d registry & cluster.

.PHONY: delete-k3d-registry
delete-k3d-registry: ## Delete k3d kyma registry.
	-k3d registry delete ${REGISTRY_NAME}

.PHONY: delete-k3d-cluster
delete-k3d-cluster: ## Delete k3d kyma cluster.
	-k3d cluster delete ${CLUSTER_NAME}

##@ Gardener

HIBERNATION_HOUR=$(shell echo $$(( ( $(shell date +%H | sed s/^0//g) + 5 ) % 24 )))
GIT_COMMIT_SHA=$(shell git rev-parse --short=8 HEAD)
ifneq (,$(GARDENER_SA_PATH))
GARDENER_K8S_VERSION?=$(shell kubectl --kubeconfig=${GARDENER_SA_PATH} get cloudprofiles.core.gardener.cloud ${GARDENER_INFRASTRUCTURE} -o=jsonpath='{.spec.kubernetes.versions[0].version}')
else
GARDENER_K8S_VERSION?=1.27.4
endif

.PHONY: provision-gardener
provision-gardener: kyma ## Provision gardener cluster with latest k8s version
	${KYMA} provision gardener ${GARDENER_INFRASTRUCTURE} -c ${GARDENER_SA_PATH} -n test-${GIT_COMMIT_SHA} -p ${GARDENER_PROJECT} -s ${GARDENER_SECRET_NAME} -k ${GARDENER_K8S_VERSION}\
		--hibernation-start="00 ${HIBERNATION_HOUR} * * ?"

.PHONY: deprovision-gardener
deprovision-gardener: kyma ## Deprovision gardener cluster
	kubectl --kubeconfig=${GARDENER_SA_PATH} annotate shoot test-${GIT_COMMIT_SHA} confirmation.gardener.cloud/deletion=true
	kubectl --kubeconfig=${GARDENER_SA_PATH} delete shoot test-${GIT_COMMIT_SHA} --wait=false

##@ Module management

.PHONY: verify-serverless
verify-serverless: ## Wait for Serverless CR to be in Ready state.
	# wait some time to make sure operator starts the reconciliation first
	sleep 10

	./verify_serverless_status.sh || \
		(make print-serverless-details && false)
	
	kubectl wait --for condition=Available -n kyma-system deployment serverless-operator --timeout=60s || \
		(make print-serverless-details && false)

	kubectl wait --for condition=Available -n kyma-system deployment serverless-ctrl-mngr --timeout=60s || \
		(make print-serverless-details && false)

.PHONY: print-serverless-details
print-serverless-details: ## Print all pods, deploys and serverless CRs in the kyma-system namespace.
	kubectl get serverless -n kyma-system -oyaml
	kubectl get deploy -n kyma-system -oyaml
	kubectl get pods -n kyma-system -oyaml

.PHONY: verify-kyma
verify-kyma: ## Wait for Kyma CR to be in Ready state.
	@./verify_kyma_status.sh

.PHONY: enable-module
enable-module: ## Patch serverless module in the right Kyma CR.
	${KYMA} alpha enable module serverless -c fast

.PHONY: patch-mod-mgr-role
patch-mod-mgr-role: ## Path lifecycle-manager clusterrole.
	kubectl patch clusterrole \
		lifecycle-manager-manager-role \
		--patch-file=./k3d-patches/patch-k3d-lifecycle-manager-clusterrole.yaml

.PHONY: install-kyma-with-lm
install-kyma-with-lm:
	${KYMA} alpha deploy \
		--ci \
		--force-conflicts

.PHONY: install-module-template
install-module-template: ## Apply k3d-ready moduletemplate.
	kubectl apply -f ${PROJECT_ROOT}/moduletemplate-k3d.yaml


.PHONY: fix-template
fix-template: ## Create moduletemplate-k3d.yaml based on moduletemplate.yaml with right URLs.
	@cat ${PROJECT_ROOT}/moduletemplate.yaml \
	| sed -e 's/remote/control-plane/g' \
		-e 's/${REGISTRY_PORT}/5000/g' \
	      	-e 's/localhost/k3d-${REGISTRY_NAME}.localhost/g' \
	> ${PROJECT_ROOT}/moduletemplate-k3d.yaml

.PHONY: apply-serverless
apply-serverless: ## Apply the k3d serverless CR.
	kubectl apply \
		-f ${PROJECT_ROOT}/config/samples/default-serverless-cr.yaml

### Internal Dependencies

.PHONY: kyma
kyma:
	@make -C ${OPERATOR_ROOT} kyma

.PHONY: module-image-dev
module-image-dev:
	@make -C ${OPERATOR_ROOT} module-image-dev \
		IMG=localhost:${REGISTRY_PORT}/${OPERATOR_IMAGE_NAME}:${OPERATOR_IMAGE_TAG}

.PHONY: module-build
module-build:
	@make -C ${OPERATOR_ROOT} module-build \
		PROW_JOB_ID= \
		IMG=k3d-${REGISTRY_NAME}:${REGISTRY_PORT}/${OPERATOR_IMAGE_NAME}:${OPERATOR_IMAGE_TAG} \
		MODULE_REGISTRY=localhost:${REGISTRY_PORT}

.PHONY: deploy
deploy:
	@make -C ${OPERATOR_ROOT} deploy \
		IMG=k3d-${REGISTRY_NAME}:${REGISTRY_PORT}/${OPERATOR_IMAGE_NAME}:${OPERATOR_IMAGE_TAG}

.PHONY: deploy-k3d
deploy-k3d:
	$(eval HASH_TAG=$(shell docker images localhost:${REGISTRY_PORT}/${OPERATOR_IMAGE_NAME}:${OPERATOR_IMAGE_TAG} --quiet))
	docker tag localhost:${REGISTRY_PORT}/${OPERATOR_IMAGE_NAME}:${OPERATOR_IMAGE_TAG} localhost:${REGISTRY_PORT}/${OPERATOR_IMAGE_NAME}:${HASH_TAG}
	docker push localhost:${REGISTRY_PORT}/${OPERATOR_IMAGE_NAME}:${HASH_TAG}
	@make -C ${OPERATOR_ROOT} deploy \
		IMG=k3d-${REGISTRY_NAME}:${REGISTRY_PORT}/${OPERATOR_IMAGE_NAME}:${HASH_TAG}

.PHONY: deploy-cluster
deploy-cluster:
	@make -C ${OPERATOR_ROOT} deploy
