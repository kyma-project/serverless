PROJECT_ROOT=./
OPERATOR_ROOT=./components/operator
include ${PROJECT_ROOT}/hack/tools/help.Makefile

##@ Installation
install-serverless-main: ## Install serverless with operator using default serverless cr
	make -C ${OPERATOR_ROOT} deploy apply-default-serverless-cr check-serverless-installation

.PHONY: install-latest-serverless
install-serverless-latest-release:
	kubectl create namespace kyma-system || true
	kubectl apply -f https://github.com/kyma-project/serverless-manager/releases/latest/download/serverless-operator.yaml
	kubectl apply -f https://github.com/kyma-project/serverless-manager/releases/latest/download/default-serverless-cr.yaml -n kyma-system
	@make -C ${OPERATOR_ROOT} verify-serverless

remove-serverless:
	make -C ${OPERATOR_ROOT} remove-serverless undeploy