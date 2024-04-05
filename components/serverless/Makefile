PROJECT_ROOT = ../..
COMPONENT_ROOT=$(PROJECT_ROOT)/components/serverless

include ${PROJECT_ROOT}/hack/tools.mk
include ${PROJECT_ROOT}/hack/help.mk

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

##@ Verification
.PHONY: test
test: KUBEBUILDER_CONTROLPLANE_START_TIMEOUT=2m
test: KUBEBUILDER_CONTROLPLANE_STOP_TIMEOUT=2m
test: manifests kubebuilder-assets ## Run unit tests.
	KUBEBUILDER_ASSETS="$(KUBEBUILDER_ASSETS)" go test -race -count=1 -coverprofile=cover.out ./...
	@echo -n "Total coverage: "
	@go tool cover -func=cover.out | grep total | awk '{print $$3}'

##@ Development
.PHONY: manifests
manifests: kustomize controller-gen ## Render CRDs
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."
	$(CONTROLLER_GEN) rbac:roleName=serverless crd webhook paths="./..." \
		object:headerFile=hack/boilerplate.go.txt \
		output:crd:artifacts:config=config/crd/bases \
		output:rbac:artifacts:config=config/rbac \
		output:webhook:artifacts:config=config/webhook
	$(KUSTOMIZE) build config/crd > config/crd/crd-serverless.yaml
	cp config/crd/crd-serverless.yaml $(PROJECT_ROOT)/config/serverless/templates/crds.yaml
	# TODO: Fix it. Now this docu is in https://kyma-project.io/#/serverless-manager/user/resources/06-10-function-cr?id=custom-resource-parameters. Remove table-gen from kyma.
	# (cd ../../hack/table-gen && make serverless-docs )

##@ Deployment

install: manifests ## Install CRDS into the k8s cluster specified in ~/.kube/config
	kubectl apply -f ../../config/serverless/templates/crds.yaml
######## function controller
CONTROLLER_NAME = function-controller

.PHONY: build-image-function-controller
build-image-function-controller:
	docker build -t $(CONTROLLER_NAME) -f $(COMPONENT_ROOT)/deploy/manager/Dockerfile $(PROJECT_ROOT)

install-controller-k3d: build-image-function-controller disable-operator ## Build and replace serverless manager from local sources on k3d
	$(eval HASH_TAG=$(shell docker images $(CONTROLLER_NAME):latest --quiet))
	docker tag $(CONTROLLER_NAME) $(CONTROLLER_NAME):$(HASH_TAG)

	k3d image import $(CONTROLLER_NAME):$(HASH_TAG) -c kyma
	kubectl set image deployment serverless-ctrl-mngr -n kyma-system manager=$(CONTROLLER_NAME):$(HASH_TAG)

######## function webhook
WEBHOOK_NAME = function-webhook

.PHONY: build-image-function-webhook push-function-webhook
build-image-function-webhook:
	docker build -t $(WEBHOOK_NAME) -f $(COMPONENT_ROOT)/deploy/webhook/Dockerfile $(PROJECT_ROOT)

install-webhook-k3d: build-image-function-webhook disable-operator ## Build and replace serverless webhook from local sources on k3d
	$(eval HASH_TAG=$(shell docker images $(WEBHOOK_NAME):latest --quiet))
	docker tag $(WEBHOOK_NAME) $(WEBHOOK_NAME):$(HASH_TAG)

	k3d image import $(WEBHOOK_NAME):$(HASH_TAG) -c kyma
	kubectl set image deployment serverless-webhook-svc -n kyma-system webhook=$(WEBHOOK_NAME):$(HASH_TAG)

######## builder init container
JOBINIT_NAME = function-build-init

.PHONY: build-image-function-build-init push-function-build-init
build-image-function-build-init:
	docker build -t $(JOBINIT_NAME) -f $(COMPONENT_ROOT)/deploy/jobinit/Dockerfile $(PROJECT_ROOT)

######## registry-gc 
REGISTRY_GC_NAME = registry-gc

.PHONY: build-image-registry-gc push-registry-gc
build-image-registry-gc:
	docker build -t $(REGISTRY_GC_NAME) -f $(COMPONENT_ROOT)/deploy/registry-gc/Dockerfile .


######## disable operator to prevent undo of local image update to k3d
disable-operator:
	kubectl scale deployment serverless-operator -n kyma-system --replicas=0
