# local variables
PROJECT_ROOT = ../../../..

include ${PROJECT_ROOT}/hack/tools/help.Makefile

.PHONY: replace-only-main-chart-images
replace-only-main-chart-images:
	./replace_serverless_chart_images.sh main-only

.PHONY: replace-all-serverless-chart-images
replace-all-serverless-chart-images:
	./replace_serverless_chart_images.sh
