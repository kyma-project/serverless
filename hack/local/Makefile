PROJECT_COMMON ?= ../common

.PHONY: run-k3d-with-serverless
run-k3d-with-serverless:
	@make -C ${PROJECT_COMMON} run

.PHONY: verify
verify:
	@make -C ${PROJECT_COMMON} verify

.PHONY: run
run: run-k3d-with-serverless

.PHONY: test
test: run verify
	@echo '### serverless operator opertational'

.PHONY: stop
stop:
	@make -C ${PROJECT_COMMON} stop
