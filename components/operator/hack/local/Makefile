# local variables
PROJECT_COMMON = ../common
PROJECT_LOCAL = ../local
PROJECT_ROOT = ../../../..
include ${PROJECT_ROOT}/hack/tools/help.Makefile

.PHONY: deploy-k3d
deploy-k3d: ## Build image and deploy operator on a k3d cluster.
	@make -C ${PROJECT_COMMON} \
		module-image-local \
		deploy-k3d \
		apply-serverless \
		verify-serverless

.PHONY: deploy-cluster
deploy-cluster: ## Deploy operator on a cluster using IMG env.
	@IMG=${IMG} make -C ${PROJECT_COMMON} \
		deploy-cluster \
		apply-serverless \
		verify-serverless

.PHONY: stop
stop: ## Delete k3d cluster and registry.
	@make -C ${PROJECT_COMMON} delete-k3d

.PHONY: run
run: ## Create k3d cluster and deploy operator.
	@IMG_VERSION="main" IMG_DIRECTORY="prod" make -C ${PROJECT_COMMON} \
		kyma \
		create-k3d \
		module-image-local \
		deploy \
		apply-serverless \
		verify-serverless

.PHONY: run-for-img
run-for-img: ## Create k3d cluster with operator using the IMG env.
	@echo "this target requires IMG env"
	@IMG=${IMG} make -C ${PROJECT_COMMON} \
		create-k3d \
		deploy-cluster \
		apply-serverless \
		verify-serverless

.PHONY: build-and-run
build-and-run: ## Create k3d cluster with operator using locally built images.
	@echo "[WARNING] THIS IS EXPERIMENTAL FEATURE! USE ON YOUR OWN RISK!"
	@echo "  --> Functions build is not working properly\n"

	@make -C ${PROJECT_COMMON} kyma \
		create-k3d

	@${PROJECT_LOCAL}/build_local_images.sh

	@IMG_VERSION="local" IMG_DIRECTORY="dev" \
		make -C ${PROJECT_COMMON} module-image-local \
		deploy \
		apply-serverless \
		verify-serverless
