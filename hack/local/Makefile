PROJECT_COMMON ?= ../common

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: run-with-lifecycle-manager
run: ## Create k3d cluster with module installed by the lifecycle-manager
	@make -C ${PROJECT_COMMON} run

.PHONY: run-without-lifecycle-manager
run-manual: ## Create k3d cluster with module
	@make -C ${PROJECT_COMMON} run-manual

.PHONY: stop
stop: ## Delete k3d cluster and registry
	@make -C ${PROJECT_COMMON} stop
