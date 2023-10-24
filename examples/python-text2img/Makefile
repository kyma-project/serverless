deploy:
	kyma apply function --dry-run -oyaml > resources/function.yaml
	kubectl apply -k resources 