.PHONY: run-k3d-with-serverless
run-k3d-with-serverless:
	@make -C ./serverless run

.PHONY: verify
verify:
	@make -C ./serverless verify

.PHONY: run
run: run-k3d-with-serverless

.PHONY: test
test: run verify
	@echo '### serverless operator opertational'

.PHONY: stop
stop:
	@make -C ./serverless stop
