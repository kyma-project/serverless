PROJECT_ROOT=.
OPERATOR_ROOT=./components/operator
include ${PROJECT_ROOT}/hack/help.mk
include ${PROJECT_ROOT}/hack/k3d.mk

##@ Installation
.PHONY: install-serverless-main
install-serverless-main: ## Install serverless with operator using default serverless cr
	make -C ${OPERATOR_ROOT} deploy-main apply-default-serverless-cr check-serverless-installation

.PHONY: install-serverless-custom-operator
install-serverless-custom-operator: ## Install serverless with operator from IMG env using default serverless cr
	$(call check-var,IMG)
	make -C ${OPERATOR_ROOT} deploy apply-default-serverless-cr check-serverless-installation

.PHONY: install-serverless-latest-release
install-serverless-latest-release:## Install serverless from latest release
	make -C ${OPERATOR_ROOT} deploy-release
	kubectl apply -f https://github.com/kyma-project/serverless-manager/releases/latest/download/default-serverless-cr.yaml -n kyma-system
	make -C ${OPERATOR_ROOT} check-serverless-installation

.PHONY: install-serverless-local-sources
install-serverless-local-sources: ## Install serverless from local sources.
	$(eval IMG_VERSION=local-$(shell date +'%Y%m%d-%H%M%S'))
	IMG_VERSION=${IMG_VERSION} ./hack/build_all.sh

	$(eval IMG=europe-docker.pkg.dev/kyma-project/dev/serverless-operator:${IMG_VERSION})
	IMG_DIRECTORY="kyma-project" IMG_VERSION=${IMG_VERSION} IMG=${IMG} make -C ${OPERATOR_ROOT} docker-build-local

	k3d image import "${IMG}" -c kyma
	IMG=${IMG} make install-serverless-custom-operator

.PHONY: remove-serverless
remove-serverless: ## Remove serverless-cr and serverless operator
	make -C ${OPERATOR_ROOT} remove-serverless undeploy

.PHONY: run-main
run-main: create-k3d install-serverless-main ## Create k3d cluster and install serverless from main

check-var = $(if $(strip $($1)),,$(error "$1" is not defined))

##@ Actions
.PHONY: module-config
module-config:
	yq ".channel = \"${CHANNEL}\" | .version = \"${MODULE_VERSION}\""\
    	module-config-template.yaml > module-config.yaml