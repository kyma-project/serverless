# incoming variables

# Image URL to use all building/pushing image targets
IMG ?= op-skr-registry.localhost:8888/unsigned/operator-images/serverless-operator:0.0.0

# Operating system architecture
OS_ARCH ?= $(shell uname -m)

# Operating system type
OS_TYPE ?= $(shell uname)


# local variables
PROJECT_ROOT = ../..

# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.27.1

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development
CONFIGOPERATOR = $(PROJECT_ROOT)/config/operator

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=operator-role crd webhook paths="./..." output:crd:artifacts:config=$(CONFIGOPERATOR)/crd/bases output:rbac:artifacts:config=$(CONFIGOPERATOR)/rbac

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="$(PROJECT_ROOT)/hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	 go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: manifests generate fmt vet envtest ## Run unit tests.
	KUBEBUILDER_CONTROLPLANE_START_TIMEOUT=2m KUBEBUILDER_CONTROLPLANE_STOP_TIMEOUT=2m KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test ./... -coverprofile cover.out

##@ Build

.PHONY: build
build: generate fmt vet ## Build operator binary.
	go build -o bin/operator main.go

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	go run ./main.go

.PHONY: docker-build-dev
docker-build-dev: manifests generate ## Build docker dev image with the operator.
	docker build -t ${IMG} -f Dockerfile.dev $(PROJECT_ROOT) --build-arg=IMG_DIRECTORY="$(IMG_DIRECTORY)" --build-arg=IMG_VERSION="$(IMG_VERSION)"

.PHONY: docker-build-local
docker-build-local: manifests generate ## Build docker local image with the operator.
	docker build -t ${IMG} -f Dockerfile.local $(PROJECT_ROOT) --build-arg=IMG_DIRECTORY="$(IMG_DIRECTORY)" --build-arg=IMG_VERSION="$(IMG_VERSION)"

.PHONY: docker-push
docker-push: ## Push docker image with the operator.
	docker push ${IMG}

##@ Deployment
IGNORE_NOT_FOUND = false

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build $(CONFIGOPERATOR)/crd | kubectl apply -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with IGNORE_NOT_FOUND=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build $(CONFIGOPERATOR)/crd | kubectl delete --ignore-not-found=$(IGNORE_NOT_FOUND) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	kubectl create namespace kyma-system || true
	cd $(CONFIGOPERATOR)/operator && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build $(CONFIGOPERATOR)/default | kubectl apply -f -

.PHONY: render-manifest
render-manifest: manifests kustomize ## Render serverless-operator.yaml manifest.
	cd $(CONFIGOPERATOR)/operator && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build $(CONFIGOPERATOR)/default > $(PROJECT_ROOT)/serverless-operator.yaml

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with IGNORE_NOT_FOUND=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build $(CONFIGOPERATOR)/default | kubectl delete --ignore-not-found=$(IGNORE_NOT_FOUND) -f -

##@ Module

.PHONY: module-image-dev
module-image-dev: docker-build-dev docker-push ## Build the dev Module Image and push it to a registry defined in IMG.
	echo "built and pushed module image $(IMG)"

.PHONY: module-image-local
module-image-local: docker-build-local docker-push ## Build the local Module Image and push it to a registry defined in IMG.
	echo "built and pushed module image $(IMG)"

##@ Tools

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/$(PROJECT_ROOT)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

########## Kustomize ###########
KUSTOMIZE ?= $(LOCALBIN)/kustomize
KUSTOMIZE_VERSION ?= v5.1.0
KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	test -s $(LOCALBIN)/kustomize || { curl -Ss $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN); }

########## Controller-Gen ###########
CONTROLLER_TOOLS_VERSION ?= v0.12.0
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

########## Envtest ###########
ENVTEST ?= $(LOCALBIN)/setup-envtest

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

########## Kyma CLI ###########
KYMA_STABILITY ?= unstable

# $(call os_error, os-type, os-architecture)
define os_error
$(error Error: unsuported platform OS_TYPE:$1, OS_ARCH:$2; to mitigate this problem set variable KYMA with absolute path to kyma-cli binary compatible with your operating system and architecture)
endef

KYMA_FILE_NAME ?= $(shell ./hack/get_kyma_file_name.sh ${OS_TYPE} ${OS_ARCH})

KYMA ?= $(LOCALBIN)/kyma-$(KYMA_STABILITY)
kyma: $(LOCALBIN) $(KYMA) ## Download kyma locally if necessary.
$(KYMA):
	## Detect if operating system
	$(if $(KYMA_FILE_NAME),,$(call os_error, ${OS_TYPE}, ${OS_ARCH}))
	test -f $@ || curl -s -Lo $(KYMA) https://storage.googleapis.com/kyma-cli-$(KYMA_STABILITY)/$(KYMA_FILE_NAME)
	chmod 0100 $(KYMA)
