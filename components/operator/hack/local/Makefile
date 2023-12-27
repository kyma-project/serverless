PROJECT_COMMON ?= ../common
PROJECT_LOCAL ?= ../local
PROJECT_CI ?= ../ci
IMG_DIRECTORY ?= "dev" # image directory ( prod / dev )
IMG_VERSION ?= "local" # image version ( release tag / commit hash / PR-<number> )

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: deploy
deploy: ## Build image and deploy manager on a cluster
	@make -C ${PROJECT_COMMON} module-image-release deploy-k3d

.PHONY: stop
stop: ## Delete k3d cluster and registry.
	@make -C ${PROJECT_COMMON} delete-k3d

.PHONY: run
run: ## Create k3d cluster with module.
	@make -C ${PROJECT_COMMON} run-without-lifecycle-manager

.PHONY: run-for-img
run-for-img: ## Create k3d cluster with module using local images.
	@echo "this target requires IMG_DIRECTORY and IMG_VERSION envs"
	@IMG_VERSION=${IMG_VERSION} IMG_DIRECTORY=${IMG_DIRECTORY} \
		make -C ${PROJECT_COMMON} replace-chart-images run-without-lifecycle-manager

.PHONY: build-and-run
build-and-run: ## Create k3d cluster with module using locally built images.
	@make -C ${PROJECT_COMMON} kyma \
		create-k3d

	@${PROJECT_LOCAL}/build_local_images.sh

	@IMG_VERSION=${IMG_VERSION} IMG_DIRECTORY=${IMG_DIRECTORY} \
		make -C ${PROJECT_COMMON} replace-chart-images \
		module-image \
		deploy \
		apply-serverless \
		verify-serverless

.PHONY: run-with-lm
run-with-lm: ## Create k3d cluster with module installed by the lifecycle-manager.
	@make -C ${PROJECT_COMMON} run-with-lifecycle-manager

.PHONY: run-with-lm-for-img
run-with-lm-for-img: ## Create k3d cluster with module installed by the lifecycle-manager using local images.
	@echo "this target requires IMG_DIRECTORY and IMG_VERSION envs"
	@IMG_VERSION=${IMG_VERSION} IMG_DIRECTORY=${IMG_DIRECTORY} \
		make -C ${PROJECT_COMMON} replace-chart-images run-with-lifecycle-manager

.PHONY: build-and-run-with-lm
build-and-run-with-lm: ## Create k3d cluster with module installed by the lifecycle-manager using locally built images.
	@make -C ${PROJECT_COMMON} kyma \
		create-k3d

	@${PROJECT_LOCAL}/build_local_images.sh

	@IMG_VERSION=${IMG_VERSION} IMG_DIRECTORY=${IMG_DIRECTORY} \
		make -C ${PROJECT_COMMON} replace-chart-images \
		module-image \
		module-build \
		fix-template \
		install-kyma-with-lm \
		patch-mod-mgr-role \
		install-module-template \
		enable-module \
		verify-kyma
