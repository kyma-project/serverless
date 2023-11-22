ROOT :=  $(shell pwd)
PROJECT_ROOT ?= $(ROOT)/../..
# Image URL to use all building/pushing image targets
LOCALBIN ?= $(PROJECT_ROOT)/bin
KUSTOMIZE ?= $(LOCALBIN)/kustomize

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)


########## Envtest ###########
ENVTEST ?= $(LOCALBIN)/setup-envtest
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.25.0

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

########## Kustomize ###########
KUSTOMIZE ?= $(LOCALBIN)/kustomize
KUSTOMIZE_VERSION ?= v4.5.5
KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	test -s $(LOCALBIN)/kustomize || { curl -Ss $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN); }

########## Controller-gen ###########
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
CONTROLLER_GEN_VERSION ?= v0.9.2

# find or download controller-gen
# download controller-gen if necessary
.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN)
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/CONTROLLER_GEN || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_GEN_VERSION)

########## other things:) ###########

.PHONY: test
test: KUBEBUILDER_CONTROLPLANE_START_TIMEOUT=2m
test: KUBEBUILDER_CONTROLPLANE_STOP_TIMEOUT=2m
test: KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)"
test: envtest
	go test -race -count=1 -coverprofile=/tmp/artifacts/cover.out ./...
	@echo -n "Total coverage: "
	@go tool cover -func=/tmp/artifacts/cover.out | grep total | awk '{print $$3}'

.PHONY: manifests
manifests: kustomize controller-gen generate-local
	$(CONTROLLER_GEN) rbac:roleName=serverless crd webhook paths="./..." \
		object:headerFile=hack/boilerplate.go.txt \
		output:crd:artifacts:config=config/crd/bases \
		output:rbac:artifacts:config=config/rbac \
		output:webhook:artifacts:config=config/webhook
	$(KUSTOMIZE) build config/crd > config/crd/crd-serverless.yaml
	cp config/crd/crd-serverless.yaml $(PROJECT_ROOT)/config/serverless/templates/crds.yaml
	# TODO: Fix it. Now this docu is in https://kyma-project.io/#/serverless-manager/user/resources/06-10-function-cr?id=custom-resource-parameters. Remove table-gen from kyma.
	# (cd ../../hack/table-gen && make serverless-docs )

# Generate code
.PHONY: generate-local
generate-local: controller-gen
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths="./..."

######## function manager
MANAGER_NAME = function-controller

.PHONY: build-image-function-controller push-image-function-controller
build-image-function-controller:
	docker build -t $(MANAGER_NAME) -f $(ROOT)/deploy/manager/Dockerfile .

install-manager-k3d: build-image-function-controller disable-operator
	$(eval HASH_TAG=$(shell docker images $(MANAGER_NAME):latest --quiet))
	docker tag $(MANAGER_NAME) $(MANAGER_NAME):$(HASH_TAG)

	k3d image import $(MANAGER_NAME):$(HASH_TAG) -c kyma
	kubectl set image deployment serverless-ctrl-mngr -n kyma-system manager=$(MANAGER_NAME):$(HASH_TAG)

######## function webhook
WEBHOOK_NAME = function-webhook

.PHONY: build-image-function-webhook push-function-webhook
build-image-function-webhook:
	docker build -t $(WEBHOOK_NAME) -f $(ROOT)/deploy/webhook/Dockerfile .

install-webhook-k3d: build-image-function-webhook disable-operator
	$(eval HASH_TAG=$(shell docker images $(WEBHOOK_NAME):latest --quiet))
	docker tag $(WEBHOOK_NAME) $(WEBHOOK_NAME):$(HASH_TAG)

	k3d image import $(WEBHOOK_NAME):$(HASH_TAG) -c kyma
	kubectl set image deployment serverless-webhook-svc -n kyma-system webhook=$(WEBHOOK_NAME):$(HASH_TAG)

######## builder init container
JOBINIT_NAME = function-build-init

.PHONY: build-image-function-build-init push-function-build-init
build-image-function-build-init:
	docker build -t $(JOBINIT_NAME) -f $(ROOT)/deploy/jobinit/Dockerfile .

######## registry-gc 
REGISTRY_GC_NAME = registry-gc

.PHONY: build-image-registry-gc push-registry-gc
build-image-registry-gc:
	docker build -t $(REGISTRY_GC_NAME) -f $(ROOT)/deploy/registry-gc/Dockerfile .


######## disable operator to prevent undo of local image update to k3d
disable-operator:
	kubectl scale deployment serverless-operator -n kyma-system --replicas=0
