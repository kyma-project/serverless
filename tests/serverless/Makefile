PROJECT_ROOT = ../..
include ${PROJECT_ROOT}/hack/help.mk

test: ## Run unit test of integration test engine
	GO111MODULE=on go test -count=1 -v ./internal/...

compile:
	CGO_ENABLED=0 go build -o main.test ./cmd/main.go

.PHONY: serverless-integration
serverless-integration: ## Run serverless integration scenario with kubectl proxy
	./run-local.sh serverless-integration

.PHONY: git-auth-integration
git-auth-integration: ## Run git authorization integration scenario with kubectl proxy
	./run-local.sh git-auth-integration

.PHONY: serverless-contract-tests
serverless-contract-tests: ## Run serverless contract tests with kubectl proxy
	./run-local.sh serverless-contract-tests
