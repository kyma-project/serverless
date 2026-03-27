PROJECT_ROOT=../..
include ${PROJECT_ROOT}/hack/help.mk

RUNNER_NAMESPACE?=performance-tests
RUNNER_IMAGE?=europe-docker.pkg.dev/kyma-project/dev/performance-test-runner:main

GRAFANA_LOCAL_PORT?=3000

##@ Cluster

.PHONY: install-monitoring
install-monitoring: ## Install monitoring tools in the cluster.
	@echo "[MAKEFILE] Installing monitoring tools..."
	helm upgrade --install monitoring -n monitoring --create-namespace config/

.PHONY: start-remote-test
start-remote-test: ## Start test in the cluster using test runner image.
	@echo "[MAKEFILE] Recreating test namespace..."
	@kubectl delete namespace $(RUNNER_NAMESPACE) &>/dev/null || true
	@kubectl create namespace $(RUNNER_NAMESPACE) &>/dev/null || true
	@echo "[MAKEFILE] Running remote test..."
	@cat templates/test_runner_job.yaml | resource_namespace=$(RUNNER_NAMESPACE) test_runner_image=$(RUNNER_IMAGE) envsubst | kubectl apply -f -
	@kubectl wait --for=condition=ready pod -l job-name=performance-test-runner -n $(RUNNER_NAMESPACE) --timeout=5m

.PHONY: forward-grafana
forward-grafana: ## Forward grafana port to localhost.
	@echo "[MAKEFILE] Getting grafana auth..."
	@echo "         > Username: admin"
	@echo "         > Password: $(shell kubectl -n monitoring get secret monitoring-grafana -o jsonpath="{.data.admin-password}" | base64 --decode)"
	@echo "[MAKEFILE] Forwarding grafana port to localhost:$(GRAFANA_LOCAL_PORT)..."
	@kubectl -n monitoring port-forward svc/monitoring-grafana $(GRAFANA_LOCAL_PORT):80

.PHONY: follow-remote-test
follow-remote-test: ## Follow logs of remote test runner.
	@echo "[MAKEFILE] Following test logs..."
	@kubectl logs -l job-name=performance-test-runner -n $(RUNNER_NAMESPACE) -f --tail=-1
	@echo "[MAKEFILE] Test completed."

##@ Dev

.PHONY: start-remote-k3d-test
start-remote-k3d-test: ## Build runner locally and start it on k3d cluster.
	@echo "[MAKEFILE] Building test runner..."
	@docker build -t performance-test-runner:local .
	@k3d image import performance-test-runner:local
	@RUNNER_IMAGE=performance-test-runner:local make start-remote-test 

.PHONY: start-test
start-test: ## Start test locally.
	@echo "[MAKEFILE] Recreating test namespace..."
	@kubectl delete namespace $(RUNNER_NAMESPACE) &>/dev/null || true
	@kubectl create namespace $(RUNNER_NAMESPACE) &>/dev/null || true
	@echo "[MAKEFILE] Running test..."
	@echo ""
	@NAMESPACE=$(RUNNER_NAMESPACE) ./test_runtimes.sh

.PHONY: start-rapid-test
start-rapid-test: ## Start all rapidly locally.
	@RAPID=true make start-test

.PHONY: start-mandelbrot-test
start-mandelbrot-test: ## Start mandelbrot test locally.
	@SCENARIO=mandelbrot make start-test

.PHONY: start-rapid-mandelbrot-test
start-rapid-mandelbrot-test: ## Start mandelbrot test rapidly locally.
	@RAPID=true make start-mandelbrot-test
