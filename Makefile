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
install-serverless-latest-release: ## Install serverless from latest release
	kubectl create namespace kyma-system || true
	kubectl apply -f https://github.com/kyma-project/serverless-manager/releases/latest/download/serverless-operator.yaml
	kubectl apply -f https://github.com/kyma-project/serverless-manager/releases/latest/download/default-dockerregistry-cr.yaml -n kyma-system
	make -C ${OPERATOR_ROOT} check-serverless-installation

.PHONY: remove-serverless
remove-serverless: ## Remove serverless-cr and serverless operator
	make -C ${OPERATOR_ROOT} remove-serverless undeploy

.PHONY: run
run: create-k3d install-serverless-main ## Create k3d cluster and install serverless from main

check-var = $(if $(strip $($1)),,$(error "$1" is not defined))

