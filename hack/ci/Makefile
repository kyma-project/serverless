# local variables
PROJECT_ROOT=../..
TEST_ROOT=${PROJECT_ROOT}/tests/serverless
OPERATOR_ROOT=${PROJECT_ROOT}/components/operator

include ${PROJECT_ROOT}/hack/tools/help.Makefile
include ${PROJECT_ROOT}/hack/tools/gardener.Makefile

##@ CI Tests
.PHONY: integration-test
integration-test: ## Run integration tests on self-prepared k3d cluster.
	make -C ${TEST_ROOT} serverless-integration serverless-contract-tests

.PHONY: gardener-integration-test
gardener-integration-test: ## Provision gardener cluster and run integration test on it.
	make provision-gardener && make -c ${OPERATOR_ROOT} install-serverless-main && \
	make integration-test \
		deprovision-gardener || \
		(make deprovision-gardener && false)

.PHONY: integration-test-on-cluster
integration-test-on-cluster: ## Install serverless with default serverless-cr, run integration tests and remove serverless-cr
	make -C ${PROJECT_ROOT} install-serverless-main
	make integration-test
	make -C ${PROJECT_ROOT} remove-serverless

.PHONY: git-auth-test-on-cluster
git-auth-test-on-cluster: # Install serverless with default serverless-cr, run git auth integration and remove serverless-cr
	make -C ${PROJECT_ROOT} install-serverless-main
	cd ${TEST_ROOT} && make git-auth-integration
	make -C ${PROJECT_ROOT} remove-serverless

.PHONY: upgrade-test
upgrade-test: ## Installs Serverless from latest, upgrades to version specified in IMG and run integration tests
upgrade-test:
	make -C ${PROJECT_ROOT} install-serverless-latest-release install-serverless-main
	make integration-test
	make -C ${PROJECT_ROOT} remove-serverless