PROJECT_ROOT=.
OPERATOR_ROOT=./components/operator
include ${PROJECT_ROOT}/hack/help.mk
include ${PROJECT_ROOT}/hack/k3d.mk

##@ Installation
.PHONY: install-dockerregistry-main
install-dockerregistry-main: ## Install dockerregistry with operator using default dockerregistry cr
	make -C ${OPERATOR_ROOT} deploy-main apply-default-dockerregistry-cr check-dockerregistry-installation

.PHONY: install-dockerregistry-custom-operator
install-dockerregistry-custom-operator: ## Install dockerregistry with operator from IMG env using default dockerregistry cr
	$(call check-var,IMG)
	make -C ${OPERATOR_ROOT} deploy apply-default-dockerregistry-cr check-dockerregistry-installation

.PHONY: install-dockerregistry-latest-release
install-dockerregistry-latest-release: ## Install dockerregistry from latest release
	kubectl create namespace kyma-system || true
	kubectl apply -f https://github.com/kyma-project/serverless-manager/releases/latest/download/dockerregistry-operator.yaml
	kubectl apply -f https://github.com/kyma-project/serverless-manager/releases/latest/download/default-dockerregistry-cr.yaml -n kyma-system
	make -C ${OPERATOR_ROOT} check-dockerregistry-installation

.PHONY: remove-dockerregistry
remove-dockerregistry: ## Remove dockerregistry-cr and dockerregistry operator
	make -C ${OPERATOR_ROOT} remove-dockerregistry undeploy

.PHONY: run
run: create-k3d install-dockerregistry-main ## Create k3d cluster and install dockerregistry from main

check-var = $(if $(strip $($1)),,$(error "$1" is not defined))

