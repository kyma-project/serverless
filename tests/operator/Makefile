PROJECT_ROOT = ../..
include ${PROJECT_ROOT}/hack/help.mk

.PHONY: test
test: ## Run integration test.
	go run main.go

.PHONY: cluster-info
cluster-info: ## Print useful info about the cluster regarding integration run
	@echo "####################### Operator Logs #######################"
	@kubectl logs -n kyma-system -l app.kubernetes.io/component=serverless-operator.kyma-project.io --tail=-1 || true
	@echo ""

	@echo "####################### Serverless CR #######################"
	@kubectl get serverless -A -oyaml || true
	@echo ""

	@echo "####################### Pods #######################"
	@kubectl get pods -A || true
	@echo ""
