.PHONY: install-monitoring
install-monitoring:
	@echo "Installing monitoring tools..."
	helm upgrade --install monitoring -n monitoring --create-namespace config/

.PHONY: apply-funcs
apply-funcs:
	@echo "Applying function configurations..."
	@kubectl apply -f funcs/
	@echo "Waiting for functions to be ready..."
	@kubectl wait function --for='jsonpath={.status.conditions[?(@.type=="Running")].status}=True' -l "test=performance"

.PHONY: start-test
start-test: apply-funcs
	@echo "Running performance tests..."
	@kubectl apply -f jobs/