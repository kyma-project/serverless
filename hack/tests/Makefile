# local variables
OPERATOR_ROOT = ../..
PROJECT_ROOT = $(OPERATOR_ROOT)/../..
PROJECT_COMMON = $(OPERATOR_ROOT)/hack/common

include ${PROJECT_ROOT}/hack/tools/help.Makefile
include ${PROJECT_ROOT}/hack/tools/gardener.Makefile

##@ Tests

#TODO: is it possible to call just make -C on tests dir?
.PHONY: integration-test
integration-test: ## Run integration tests on self-prepared k3d cluster.
integration-test:
	cd ${PROJECT_ROOT}/tests/serverless && make serverless-integration serverless-contract-tests

.PHONY: gardener-integration-test
gardener-integration-test: ## Provision gardener cluster and run integration test on it.
	@make provision-gardener && make -c ${OPERATOR_ROOT} install-default-serverless &&
	@make integration-test \
		deprovision-gardener || \
		(make deprovision-gardener && false)

.PHONY: integration-test-on-cluster
integration-test-on-cluster: ## Installs and removes after default serverless and run integration tests on existing k3d cluster.
	make -C ${PROJECT_ROOT} install-default-serverless
	@make integration-test
	make -C ${PROJECT_ROOT} remove-serverless

.PHONY: git-auth-test-on-cluster
git-auth-test-on-cluster:
	make -C ${PROJECT_ROOT} install-default-serverless
	cd ${PROJECT_ROOT}/tests/serverless && make git-auth-integration
	make -C ${PROJECT_ROOT} remove-serverless

.PHONY: upgrade-test
upgrade-test: ## Installs Serverless from latest, upgrades to local and run integration tests
upgrade-test:
	make -C ${PROJECT ROOT} install-latest-serverless install-default-serverless
	@make integration-test