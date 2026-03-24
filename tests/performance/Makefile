.PHONY: install-monitoring
install-monitoring:
	@echo "Installing monitoring tools..."
	helm upgrade --install monitoring -n monitoring --create-namespace config/

.PHONY: start-test
start-test:
	@kubectl create namespace performance-tests &>/dev/null || true
	@NAMESPACE=performance-tests ./test_runtimes.sh

.PHONY: start-rapid-test
start-rapid-test:
	@RAPID=true make start-test
