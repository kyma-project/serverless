START_TIME := $(shell date +%s)
IMG ?= app:$(START_TIME)

##@ General
.DEFAULT_GOAL=help
.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: docker-build
docker-build: ## Build the Docker image.
	docker build -t $(IMG) .

.PHONY: docker-push
docker-push: ## Push the Docker image to the registry.
	docker push $(IMG)

##@ Development

.PHONY: k3d-deploy
k3d-deploy: docker-build ## Build and deploy the image to k3d cluster.
	k3d image import $(IMG)
	yq '.spec.template.spec.containers[] |= select(.name == "function") |= .image="$(IMG)"' resources/deployment.yaml | kubectl apply -f -
	kubectl apply -f resources/service.yaml

.PHONY: run
run: ## Run the application locally.
	python3 -m pip install -r requirements.txt --user --break-system-packages
	python3 server.py
