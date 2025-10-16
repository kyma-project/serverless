<!-- This document is a part of the Secure Development in the Kyma Environment section on HP -->
# Function Security

To eliminate potential security risks when using Functions, bear in mind these facts:

- By default, JSON Web Tokens (JWTs) issued by an OpenID Connect-compliant identity provider do not provide the scope parameter for Functions. This means that if you expose your Function and secure it with a JWT, you can use the token to validate access to all Functions within the cluster as well as other JWT-protected services.

- Kyma provides base images for serverless runtimes. Those default runtimes are maintained with regards to commonly known security advisories. It is possible to use a custom runtime image. For more information, see [Override Runtime Image](tutorials/01-110-override-runtime-image.md). In such a case, you are responsible for security compliance and assessment of exploitability of any potential vulnerabilities of the custom runtime image.

- Kyma does not run any security scans against Functions and their images. Before you store any sensitive data in Functions, consider the potential risk of data leakage.

- Kyma does not define any authorization policies that would restrict Functions' access to other resources within the namespace. If you deploy a Function in a given namespace, it can freely access all events and APIs of services within this namespace.

- Since Kubernetes is [moving from PodSecurityPolicies to PodSecurity Admission Controller](https://kubernetes.io/docs/tasks/configure-pod-container/migrate-from-psp/), traditional Kyma Functions with build jobs require running in namespaces with the `baseline` Pod security level. However, buildless mode eliminates build jobs, allowing Functions to run in namespaces with more restrictive Pod security levels such as `restricted`.

- All administrators and regular users who have access to a specific namespace in a cluster can also access:

  - Source code of all Functions within this namespace
