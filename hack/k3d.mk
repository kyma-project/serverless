CLUSTER_NAME ?= kyma
REGISTRY_PORT ?= 5001
REGISTRY_NAME ?= ${CLUSTER_NAME}-registry.localhost

ifndef PROJECT_ROOT
$(error PROJECT_ROOT is undefined)
endif

##@ K3D

.PHONY: create-k3d
create-k3d: delete-k3d ## Delete old k3d registry and cluster. Create preconfigured k3d with registry
	k3d registry create ${REGISTRY_NAME} --port 5000
	k3d cluster create ${CLUSTER_NAME} --registry-use "k3d-${REGISTRY_NAME}:5000"
	kubectl create namespace kyma-system

.PHONY: delete-k3d
delete-k3d: delete-k3d-cluster delete-k3d-registry ## Delete k3d registry & cluster.

.PHONY: delete-k3d-registry
delete-k3d-registry: ## Delete k3d kyma registry.
	-k3d registry delete ${REGISTRY_NAME}

.PHONY: delete-k3d-cluster
delete-k3d-cluster: ## Delete k3d kyma cluster.
	-k3d cluster delete ${CLUSTER_NAME}
