# local variables
PROJECT_ROOT = ../..

include ${PROJECT_ROOT}/hack/tools/Makefile
include ${PROJECT_ROOT}/hack/tools/help.Makefile

# incoming variables

# Image URL to use all building/pushing image targets
IMG ?= op-skr-registry.localhost:8888/unsigned/operator-images/serverless-operator:0.0.0

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

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
	docker build -t ${IMG} -f Dockerfile $(PROJECT_ROOT) --build-arg=PURPOSE="dev" --build-arg=IMG_DIRECTORY="$(IMG_DIRECTORY)" --build-arg=IMG_VERSION="$(IMG_VERSION)"

.PHONY: docker-build-local
docker-build-local: manifests generate ## Build docker local image with the operator.
	docker build -t ${IMG} -f Dockerfile $(PROJECT_ROOT) --build-arg=PURPOSE="local" --build-arg=IMG_DIRECTORY="$(IMG_DIRECTORY)" --build-arg=IMG_VERSION="$(IMG_VERSION)"

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
