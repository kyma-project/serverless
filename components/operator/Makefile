# local variables
PROJECT_ROOT = ../..

include ${PROJECT_ROOT}/hack/tools.mk
include ${PROJECT_ROOT}/hack/help.mk

# incoming variables

# Image URL to use all building/pushing image targets
IMG ?= europe-docker.pkg.dev/kyma-project/prod/serverless-operator:main

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

##@ Development
CONFIG_OPERATOR = $(PROJECT_ROOT)/config/operator/base
CONFIG_OPERATOR_DEV = $(PROJECT_ROOT)/config/operator/dev
CONFIG_OPERATOR_MAIN = https://github.com/kyma-project/serverless/config/operator/base?ref=main

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=operator-role crd webhook paths="./..." output:crd:artifacts:config=$(CONFIG_OPERATOR)/crd/bases output:rbac:artifacts:config=$(CONFIG_OPERATOR)/rbac

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="$(PROJECT_ROOT)/hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	 go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

##@ Verification

.PHONY: test
test: manifests generate fmt vet kubebuilder-assets ## Run unit tests.
	KUBEBUILDER_CONTROLPLANE_START_TIMEOUT=2m KUBEBUILDER_CONTROLPLANE_STOP_TIMEOUT=2m KUBEBUILDER_ASSETS="$(KUBEBUILDER_ASSETS)" go test ./... -coverprofile cover.out

##@ Build

.PHONY: build
build: generate fmt vet ## Build operator binary.
	go build -o bin/operator main.go

.PHONY: run
run: manifests generate fmt vet ## Run a operator from your host.
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
OPERATOR_NAME = serverless-operator

.PHONY: build-image-operator
build-image-operator: ## Build serverless operator from local sources on k3d
	docker build -t $(OPERATOR_NAME) -f Dockerfile $(PROJECT_ROOT)

.PHONY: install-operator-k3d
install-operator-k3d: build-image-operator ## Build and replace serverless operator from local sources on k3d
	$(eval HASH_TAG=$(shell docker images $(OPERATOR_NAME):latest --quiet))
	docker tag $(OPERATOR_NAME) $(OPERATOR_NAME):$(HASH_TAG)

	k3d image import $(OPERATOR_NAME):$(HASH_TAG) -c kyma
	kubectl set image deployment serverless-operator -n kyma-system manager=$(OPERATOR_NAME):$(HASH_TAG)

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build $(CONFIG_OPERATOR)/crd | kubectl apply -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with IGNORE_NOT_FOUND=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build $(CONFIG_OPERATOR)/crd | kubectl delete --ignore-not-found=$(IGNORE_NOT_FOUND) -f -

.PHONY: deploy-main
deploy-main: manifests kustomize ## Deploy operator to the K8s cluster specified in ~/.kube/config.
	kubectl create namespace kyma-system || true
	$(KUSTOMIZE) build $(CONFIG_OPERATOR_MAIN) | kubectl apply -f -

.PHONY: generate-kustomization-dev
generate-kustomization-dev:
	cp $(CONFIG_OPERATOR_DEV)/kustomization.yaml.tpl $(CONFIG_OPERATOR_DEV)/kustomization.yaml
	cp $(CONFIG_OPERATOR_DEV)/default-images-patch.yaml.tpl $(CONFIG_OPERATOR_DEV)/default-images-patch.yaml
	PROJECT_ROOT=$(PROJECT_ROOT) IMG_DIRECTORY=$(IMG_DIRECTORY) IMG_VERSION=$(IMG_VERSION) $(CONFIG_OPERATOR_DEV)/replcase_images_patch.sh

.PHONY: deploy-release
deploy-release: ## Deploy operator to the k8s cluster specified in ~/.kube/config with image from latest release.
	kubectl create namespace kyma-system || true
	kubectl apply -f https://github.com/kyma-project/serverless-manager/releases/latest/download/serverless-operator.yaml

.PHONY: deploy
deploy: manifests kustomize generate-kustomization-dev ## Deploy operator to the K8s cluster specified in ~/.kube/config with image from IMG env.
	kubectl create namespace kyma-system || true
	cd $(CONFIG_OPERATOR_DEV) && $(KUSTOMIZE) edit set image europe-docker.pkg.dev/kyma-project/prod/serverless-operator=${IMG}
	$(KUSTOMIZE) build $(CONFIG_OPERATOR_DEV) | kubectl apply -f -

.PHONY: deploy-local
deploy-local: build-image-operator ## Deploy operator to the K8s cluster specified in ~/.kube/config with locally build image.
	$(eval HASH_TAG=$(shell docker images $(OPERATOR_NAME):latest --quiet))
	docker tag $(OPERATOR_NAME) $(OPERATOR_NAME):$(HASH_TAG)

	$(eval IMG=$(OPERATOR_NAME):$(HASH_TAG))
	k3d image import $(IMG) -c kyma
	IMG=$(IMG) make deploy

.PHONY: undeploy
undeploy: ## Undeploy operator from the K8s cluster specified in ~/.kube/config. Call with IGNORE_NOT_FOUND=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build $(CONFIG_OPERATOR) | kubectl delete --ignore-not-found=$(IGNORE_NOT_FOUND) -f -

.PHONY: render-manifest
render-manifest: kustomize ## Render serverless-operator.yaml manifest.
	$(KUSTOMIZE) build $(CONFIG_OPERATOR) > $(PROJECT_ROOT)/serverless-operator.yaml

.PHONY: apply-default-serverless-cr
apply-default-serverless-cr: ## Apply the serverless CR.
	kubectl apply \
		-f ${PROJECT_ROOT}/config/samples/default-serverless-cr.yaml
		
.PHONY: apply-legacy-serverless-cr
apply-legacy-serverless-cr: ## Apply the legacy serverless CR (opt-out from buildless mode, using legacy function container in-cluster build).
	kubectl apply \
		-f ${PROJECT_ROOT}/config/samples/legacy-serverless-cr.yaml

.PHONY: apply-custom-serverless-cr
apply-custom-serverless-cr: ## Apply the serverless CR by path.
	kubectl apply \
		-f ${PROJECT_ROOT}/${SERVERLESS_CR_PATH}
		
.PHONY: remove-default-serverless-cr
remove-serverless: ## Remove Serverless CR
	kubectl delete serverless -n kyma-system default --timeout 2m || (kubectl get serverless -n kyma-system -oyaml && false)


.PHONY: check-serverless-installation
check-serverless-installation: ## Wait for Serverless CR to be in Ready state.
	# wait some time to make sure operator starts the reconciliation first
	sleep 10

	./hack/verify_serverless_status.sh || \
		(make print-serverless-details && false)

	kubectl wait --for condition=Available -n kyma-system deployment serverless-operator --timeout=60s || \
		(make print-serverless-details && false)

	kubectl wait --for condition=Available -n kyma-system deployment -l app.kubernetes.io/name=serverless --timeout=60s || \
	  (make print-serverless-details && false)

.PHONY: print-serverless-details
print-serverless-details: ## Print all pods, deploys and serverless CRs in the kyma-system namespace.
	kubectl get serverless -n kyma-system -oyaml
	kubectl get deploy -n kyma-system -oyaml
	kubectl get pods -n kyma-system -oyaml

##@ Module

.PHONY: module-image-dev
module-image-dev: docker-build-dev docker-push ## Build the dev Module Image and push it to a registry defined in IMG.
	echo "built and pushed module image $(IMG)"

.PHONY: module-image-local
module-image-local: docker-build-local docker-push ## Build the local Module Image and push it to a registry defined in IMG.
	echo "built and pushed module image $(IMG)"
