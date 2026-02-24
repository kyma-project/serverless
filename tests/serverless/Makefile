ifneq (,$(wildcard ./.env))
    include .env
    export
endif

PROJECT_ROOT = ../..
include ${PROJECT_ROOT}/hack/help.mk

test: ## Run unit test of integration test engine
	GO111MODULE=on go test -count=1 -v ./internal/...

compile:
	CGO_ENABLED=0 go build -o main.test ./cmd/main.go

.PHONY: serverless-integration
serverless-integration: ## Run serverless integration scenario with kubectl proxy
	./run-local.sh serverless-integration

.PHONY: serverless-integration-fips-mode
serverless-integration-fips-mode: ## Run serverless integration scenario with kubectl proxy in FIPS mode
	./run-local.sh serverless-integration-fips-mode

.PHONY: git-auth-integration
git-auth-integration: ## Run git authorization integration scenario with kubectl proxy
	./run-local.sh git-auth-integration

.PHONY: git-auth-integration-fips
git-auth-integration-fips: ## Run git authorization integration scenario with kubectl proxy in FIPS mode
	./run-local.sh git-auth-integration-fips

.PHONY: serverless-contract-tests
serverless-contract-tests: ## Run serverless contract tests with kubectl proxy
	./run-local.sh serverless-contract-tests

.PHONY: cluster-info
cluster-info: ## Collect logs from the cluster
	./cluster_info.sh
