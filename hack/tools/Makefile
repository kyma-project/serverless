## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/$(PROJECT_ROOT)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

##@ Tools

########## Kyma CLI ###########
KYMA_STABILITY ?= unstable

# $(call os_error, os-type, os-architecture)
define os_error
$(error Error: unsuported platform OS_TYPE:$1, OS_ARCH:$2; to mitigate this problem set variable KYMA with absolute path to kyma-cli binary compatible with your operating system and architecture)
endef

KYMA ?= $(LOCALBIN)/kyma-$(KYMA_STABILITY)
kyma: $(LOCALBIN) $(KYMA) ## Download kyma locally if necessary.
$(KYMA):
	# Operating system architecture
	$(eval OS_ARCH=$(shell uname -m))
	# Operating system type
	$(eval OS_TYPE=$(shell uname))
	$(eval KYMA_FILE_NAME=$(shell ${PROJECT_ROOT}/hack/tools/get_kyma_file_name.sh ${OS_TYPE} ${OS_ARCH}))
	## Detect if operating system
	$(if $(KYMA_FILE_NAME),,$(call os_error, ${OS_TYPE}, ${OS_ARCH}))
	test -f $@ || curl -s -Lo $(KYMA) https://storage.googleapis.com/kyma-cli-$(KYMA_STABILITY)/$(KYMA_FILE_NAME)
	chmod 0100 $(KYMA)

########## Kustomize ###########
KUSTOMIZE ?= $(LOCALBIN)/kustomize
KUSTOMIZE_VERSION ?= v4.5.5
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

# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.25.0

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest
