START_TIME := $(shell date +%s)
IMG ?= app:$(START_TIME)

.PHONY: k3d-deploy
k3d-deploy:
	docker build -t $(IMG) .
	k3d image import $(IMG)
	yq '.spec.template.spec.containers[] |= select(.name == "function") |= .image="$(IMG)"' resources/deployment.yaml | kubectl apply -f -
	kubectl apply -f resources/service.yaml

.PHONY: run
run:
	npm install
	npm start
