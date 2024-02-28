PROJECT_ROOT=./
OPERATOR_ROOT=./components/operator
include ${PROJECT_ROOT}/hack/tools/help.Makefile
include ${PROJECT_ROOT}/hack/tools/k3d.Makefile

##@ Installation
.PHONY: install-serverless-main
install-serverless-main: ## Install serverless with operator using default serverless cr
	make -C ${OPERATOR_ROOT} deploy-main apply-default-serverless-cr check-serverless-installation

.PHONY: install-serverless-latest-release
install-serverless-latest-release: ## Install serverless from latest release
	kubectl create namespace kyma-system || true
	kubectl apply -f https://github.com/kyma-project/serverless-manager/releases/latest/download/serverless-operator.yaml
	kubectl apply -f https://github.com/kyma-project/serverless-manager/releases/latest/download/default-serverless-cr.yaml -n kyma-system
	make -C ${OPERATOR_ROOT} check-serverless-installation

.PHONY: remove-serverless
remove-serverless: ## Remove serverless-cr and serverless operator
	make -C ${OPERATOR_ROOT} remove-serverless undeploy

run: create-k3d install-serverless-main ## Create k3d cluster and install serverless from main
