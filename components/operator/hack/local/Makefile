OPERATOR_ROOT ?= ../..
PROJECT_ROOT ?= $(OPERATOR_ROOT)/../..
PROJECT_COMMON ?= ../common

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: run-with-lifecycle-manager
run-with-lifecycle-manager: ## Create k3d cluster with module installed by the lifecycle-manager.
	make -C ${PROJECT_COMMON} kyma \
	create-k3d

	make build-and-replace-chart-images

	make -C ${PROJECT_COMMON} module-image \
	module-build \
	fix-template \
	install-kyma-with-lm \
	patch-mod-mgr-role \
	install-module-template \
	enable-module \
	verify-kyma

.PHONY: deploy
deploy: ## Build image and deploy manager on a cluster
	@make -C ${PROJECT_COMMON} module-image deploy-k3d

.PHONY: run-without-lifecycle-manager
run-without-lifecycle-manager: ## Create k3d cluster with module.
	@make -C ${PROJECT_COMMON} kyma \
	create-k3d

	@make build-and-replace-chart-images

	@make -C ${PROJECT_COMMON} module-image \
	deploy \
	apply-serverless \
	verify-serverless

.PHONY: stop
stop: ## Delete k3d cluster and registry.
	@make -C ${PROJECT_COMMON} delete-k3d

.PHONY: build-and-replace-chart-images
build-and-replace-chart-images:
	@${PROJECT_ROOT}/components/serverless/hack/poc_script_replace_local_images.sh
