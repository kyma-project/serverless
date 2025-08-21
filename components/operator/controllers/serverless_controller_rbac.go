package controllers

// TODO: serverless-manager doesn't need almost half of these rbscs. It uses them only to create another rbacs ( is there any onther option? - investigate )

//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups="",resources=services;secrets;serviceaccounts;configmaps,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups="",resources=nodes,verbs=list;watch;get
//+kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete;deletecollection

//+kubebuilder:rbac:groups=apps,resources=replicasets,verbs=list
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get
//+kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;patch;delete;deletecollection

//+kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete;deletecollection

//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get

//+kubebuilder:rbac:groups=policy,resources=podsecuritypolicies,verbs=use

//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles;clusterrolebindings,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings;roles,verbs=get;list;watch;create;update;patch;delete;deletecollection

//+kubebuilder:rbac:groups=serverless.kyma-project.io,resources=functions,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups=serverless.kyma-project.io,resources=functions/status,verbs=get;list;watch;create;update;patch;delete;deletecollection

//+kubebuilder:rbac:groups=operator.kyma-project.io,resources=serverlesses,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups=operator.kyma-project.io,resources=serverlesses/status,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups=operator.kyma-project.io,resources=serverlesses/finalizers,verbs=get;list;watch;create;update;patch;delete;deletecollection

//+kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=validatingwebhookconfigurations;mutatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete;deletecollection

//+kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch;create;update;patch;delete;deletecollection

//+kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups=scheduling.k8s.io,resources=priorityclasses,verbs=get;list;watch;create;update;patch;delete;deletecollection

//+kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete;deletecollection
