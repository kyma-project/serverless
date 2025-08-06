resources:
- ../base
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
# To overwrite image in base it has to point to the image in base kustomization.yaml
images:
- name: europe-docker.pkg.dev/kyma-project/prod/serverless-operator
  newName: local-registry
  newTag: local
patches:
- path: default-images-patch.yaml
