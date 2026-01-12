## Location to install dependencies to
ifndef PROJECT_ROOT
$(error PROJECT_ROOT is undefined)
endif
LOCALBIN ?= $(realpath $(PROJECT_ROOT))/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

# Operating system architecture
OS_ARCH=$(shell uname -m)
# Operating system type
OS_TYPE=$(shell uname)

##@ Tools

########## Kyma CLI ###########
KYMA_STABILITY ?= unstable

define os_error
$(error Error: unsuported platform OS_TYPE:$1, OS_ARCH:$2; to mitigate this problem set variable KYMA with absolute path to kyma-cli binary compatible with your operating system and architecture)
endef

KYMA ?= $(LOCALBIN)/kyma-$(KYMA_STABILITY)
kyma: $(LOCALBIN) $(KYMA) ## Download kyma locally if necessary.
$(KYMA):
	$(eval KYMA_FILE_NAME=$(shell ${PROJECT_ROOT}/hack/get_kyma_file_name.sh ${OS_TYPE} ${OS_ARCH}))
	## Detect if operating system
	$(if $(KYMA_FILE_NAME),,$(call os_error, ${OS_TYPE}, ${OS_ARCH}))
	test -f $@ || curl -s -Lo $(KYMA) https://storage.googleapis.com/kyma-cli-$(KYMA_STABILITY)/$(KYMA_FILE_NAME)
	chmod +x $(KYMA)

########## Kustomize ###########
KUSTOMIZE ?= $(LOCALBIN)/kustomize
KUSTOMIZE_VERSION ?= v5.8.0
KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	@if [ -z ${GITHUB_TOKEN} ]; then \
		test -s $(LOCALBIN)/kustomize || { curl -Ss $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN); }; \
	else \
		test -s $(LOCALBIN)/kustomize || { curl -Ss $(KUSTOMIZE_INSTALL_SCRIPT) --header "Authorization: Bearer ${GITHUB_TOKEN}" | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN); }; \
	fi

########## Controller-Gen ###########
CONTROLLER_TOOLS_VERSION ?= v0.20.0
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen

.PHONY: controller-gen $(CONTROLLER_GEN)
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	test "$(shell ${LOCALBIN}/controller-gen --version)" = "Version: ${CONTROLLER_TOOLS_VERSION}" || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

########## Envtest ###########
ENVTEST ?= $(LOCALBIN)/setup-envtest
KUBEBUILDER_ASSETS=$(LOCALBIN)/k8s/kubebuilder_assets

define path_error
$(error Error: path is empty: $1)
endef

# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.35.0

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

# Envtest download binaries to k8s/(K8S_Version)-(arch)-(os) directory which is different on every machine.
# To use the same `envtest` binaries on CI and during local development this target moves it to upfront known directory.
# Additionaly `OS-ARCH` return X86_64, but envtest uses `amd64` name.
.PHONY: kubebuilder-assets
kubebuilder-assets: envtest
	$(eval DOWNLOADED_ASSETS=$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path))
	$(if $(DOWNLOADED_ASSETS),,$(call path_error, ${DOWNLOADED_ASSETS}))
	chmod -R 755 $(DOWNLOADED_ASSETS)
	mkdir -p $(LOCALBIN)/k8s/kubebuilder_assets/
	mv $(DOWNLOADED_ASSETS)/* $(LOCALBIN)/k8s/kubebuilder_assets/
	rm -d $(DOWNLOADED_ASSETS)
