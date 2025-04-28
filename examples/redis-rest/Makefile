deploy:
	helm install my-redis oci://registry-1.docker.io/bitnamicharts/redis
	kubectl apply -f k8s-resources

undeploy:
	helm uninstall my-redis
	kubectl delete -f k8s-resources

read-logs:
	kubectl logs -l serverless.kyma-project.io/function-name=rest-fn -f
