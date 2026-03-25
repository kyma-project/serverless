.PHONY: install-monitoring
install-monitoring:
	@echo "Installing monitoring tools..."
	helm upgrade --install monitoring -n monitoring --create-namespace config/

.PHONY: start-test
start-test:
	@echo "Recreating test namespace..."
	@kubectl delete namespace performance-tests &>/dev/null || true
	@kubectl create namespace performance-tests &>/dev/null || true
	@echo "Running test..."
	@echo ""
	@NAMESPACE=performance-tests ./test_runtimes.sh

.PHONY: start-rapid-test
start-rapid-test:
	@RAPID=true make start-test

.PHONY: start-mandelbrot-test
start-mandelbrot-test:
	@SCENARIO=mandelbrot make start-test

.PHONY: start-rapid-mandelbrot-test
start-rapid-mandelbrot-test:
	@RAPID=true make start-mandelbrot-test
