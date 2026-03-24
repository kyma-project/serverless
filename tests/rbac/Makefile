PROJECT_ROOT = ../..
include ${PROJECT_ROOT}/hack/help.mk

.PHONY: test
test: ## Run RBAC aggregation integration test (requires edit-scoped kubeconfig and RBAC_TEST_NAMESPACE env var).
	go run main.go


