# local variables
OPERATOR_ROOT = ../..
PROJECT_ROOT = $(OPERATOR_ROOT)/../..

#include other makefiles
include ${PROJECT_ROOT}/hack/tools/Makefile

CLUSTER_NAME = kyma
REGISTRY_PORT = 5001
REGISTRY_NAME = ${CLUSTER_NAME}-registry
OPERATOR_IMAGE_NAME = serverless-operator-dev-local
OPERATOR_IMAGE_TAG = 0.0.0

.PHONY: run
run: ## Create k3d cluster and deploy module.
run: kyma \
	create-k3d \
	module-image-dev \
	deploy \
	apply-serverless \
	verify-serverless

### Internal Dependencies

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
