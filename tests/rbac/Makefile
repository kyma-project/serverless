PROJECT_ROOT = ../..
include ${PROJECT_ROOT}/hack/help.mk

.PHONY: test
test: ## Run RBAC integration test. Requires RBAC_TEST_SCENARIO=edit|view|precreate, RBAC_TEST_NAMESPACE, and RESOURCES_FILE.
	go run main.go

.PHONY: test-edit
test-edit: ## Run RBAC edit scenario (full CRUD).
	RBAC_TEST_SCENARIO=edit go run main.go

.PHONY: test-view
test-view: ## Run RBAC view scenario (read-only). Resources must be pre-created.
	RBAC_TEST_SCENARIO=view go run main.go

.PHONY: precreate
precreate: ## Pre-create test resources before switching to a lower-privilege role for the view scenario.
	RBAC_TEST_SCENARIO=precreate go run main.go
