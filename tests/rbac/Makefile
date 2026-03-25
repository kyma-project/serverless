PROJECT_ROOT = ../..
include ${PROJECT_ROOT}/hack/help.mk

.PHONY: deploy
deploy: ## Deploy operator, buildless-serverless, and default CR.
	kubectl create namespace kyma-system || true
	make -C ${PROJECT_ROOT}/components/operator deploy
	make -C ${PROJECT_ROOT}/components/buildless-serverless deploy
	kubectl apply -f ${PROJECT_ROOT}/config/samples/default-serverless-cr.yaml

.PHONY: cleanup
cleanup: ## Remove default CR, buildless-serverless, and operator.
	kubectl delete -f ${PROJECT_ROOT}/config/samples/default-serverless-cr.yaml --ignore-not-found
	make -C ${PROJECT_ROOT}/components/buildless-serverless undeploy ignore-not-found=true
	make -C ${PROJECT_ROOT}/components/operator undeploy IGNORE_NOT_FOUND=true

.PHONY: test
test: deploy ## Deploy, run RBAC integration test, then cleanup.
	@go run main.go config.yaml; EXIT_CODE=$$?; \
	$(MAKE) cleanup; \
	exit $$EXIT_CODE


