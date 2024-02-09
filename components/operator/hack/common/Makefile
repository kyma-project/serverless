# local variables
OPERATOR_ROOT = ../..
PROJECT_ROOT = $(OPERATOR_ROOT)/../..

KYMA = ${PROJECT_ROOT}/bin/kyma-unstable

CLUSTER_NAME = kyma
REGISTRY_PORT = 5001
REGISTRY_NAME = ${CLUSTER_NAME}-registry
OPERATOR_IMAGE_NAME = serverless-operator-dev-local
OPERATOR_IMAGE_TAG = 0.0.0
GARDENER_INFRASTRUCTURE = az


##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: run
run: ## Create k3d cluster and deploy module.
run: kyma \
	create-k3d \
	module-image-dev \
	deploy \
	apply-serverless \
	verify-serverless

.PHONY: run-on-cluster
run-on-cluster: ## Deploy module on existing cluster.
run-on-cluster: \
	deploy-cluster \
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

.PHONY: module-image-local
module-image-local:
	@make -C ${OPERATOR_ROOT} module-image-local \
		IMG=localhost:${REGISTRY_PORT}/${OPERATOR_IMAGE_NAME}:${OPERATOR_IMAGE_TAG}

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
