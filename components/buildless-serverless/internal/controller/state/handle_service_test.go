package state

import (
	"context"
	"testing"
	"time"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/fsm"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/resources"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func Test_sFnHandleService(t *testing.T) {
	t.Run("when service does not exist on kubernetes should create service and apply it", func(t *testing.T) {
		// Arrange
		// some service on k8s, but it is not the service we expect
		someSvc := corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "angry-archimedes-name",
				Namespace: "affectionate-agnesi-ns"},
			Spec: corev1.ServiceSpec{
				ExternalName: "agitated-albattani-external"}}
		// scheme and fake client
		scheme := runtime.NewScheme()
		require.NoError(t, serverlessv1alpha2.AddToScheme(scheme))
		require.NoError(t, corev1.AddToScheme(scheme))
		updateWasCalled := false
		k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&someSvc).WithInterceptorFuncs(interceptor.Funcs{
			Update: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
				updateWasCalled = true
				return nil
			},
		}).Build()
		// machine with our function
		m := fsm.StateMachine{
			State: fsm.SystemState{
				Function: serverlessv1alpha2.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "brave-babbage-name",
						Namespace: "busy-banach-ns"}}},
			Log:    zap.NewNop().Sugar(),
			Client: k8sClient,
			Scheme: scheme}

		// Act
		next, result, err := sFnHandleService(context.Background(), &m)

		// Assert
		// no errors
		require.Nil(t, err)
		// we expect stop and requeue
		require.NotNil(t, result)
		require.Equal(t, ctrl.Result{Requeue: true}, *result)
		// no next state (we will stop)
		require.Nil(t, next)
		// service has not been updated
		require.False(t, updateWasCalled)
		// function has proper condition
		requireContainsCondition(t, m.State.Function.Status,
			serverlessv1alpha2.ConditionRunning,
			metav1.ConditionUnknown,
			serverlessv1alpha2.ConditionReasonServiceCreated,
			"Service brave-babbage-name created")
		// service has been applied to k8s
		appliedSvc := &corev1.Service{}
		getErr := k8sClient.Get(context.Background(), client.ObjectKey{
			Name:      "brave-babbage-name",
			Namespace: "busy-banach-ns",
		}, appliedSvc)
		require.NoError(t, getErr)
		// service should have owner ref to our function
		require.NotEmpty(t, appliedSvc.OwnerReferences)
		require.Equal(t, "Function", appliedSvc.OwnerReferences[0].Kind)
		require.Equal(t, "brave-babbage-name", appliedSvc.OwnerReferences[0].Name)
	})
	t.Run("when cannot get service from kubernetes should stop processing", func(t *testing.T) {
		// Arrange
		// scheme and fake client
		scheme := runtime.NewScheme()
		require.NoError(t, serverlessv1alpha2.AddToScheme(scheme))
		require.NoError(t, corev1.AddToScheme(scheme))
		createOrUpdateWasCalled := false
		k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithInterceptorFuncs(interceptor.Funcs{
			Get: func(ctx context.Context, client client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
				return errors.New("angry-cerf error message")
			},
			Create: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
				createOrUpdateWasCalled = true
				return nil
			},
			Update: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
				createOrUpdateWasCalled = true
				return nil
			},
		}).Build()
		// machine with our function
		m := fsm.StateMachine{
			State: fsm.SystemState{
				Function: serverlessv1alpha2.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "brave-babbage-name",
						Namespace: "busy-banach-ns"}}},
			Log:    zap.NewNop().Sugar(),
			Client: k8sClient,
			Scheme: scheme}

		// Act
		next, result, err := sFnHandleService(context.Background(), &m)

		// Assert
		// we expect error
		require.NotNil(t, err)
		require.ErrorContains(t, err, "angry-cerf error message")
		// no result because of error
		require.Nil(t, result)
		// no next state (we will stop)
		require.Nil(t, next)
		// TODO: should we set condition in this case?
		// service has not been created or updated
		require.False(t, createOrUpdateWasCalled)
	})
	t.Run("when service does not exist on kubernetes and create fails should stop processing", func(t *testing.T) {
		// Arrange
		// scheme and fake client
		scheme := runtime.NewScheme()
		require.NoError(t, serverlessv1alpha2.AddToScheme(scheme))
		require.NoError(t, corev1.AddToScheme(scheme))
		k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithInterceptorFuncs(interceptor.Funcs{
			Create: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
				return errors.New("sweet-dirac error message")
			},
		}).Build()
		// machine with our function
		m := fsm.StateMachine{
			State: fsm.SystemState{
				Function: serverlessv1alpha2.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "mystifying-mayer-name",
						Namespace: "inspiring-bhaskara-ns"}}},
			Log:    zap.NewNop().Sugar(),
			Client: k8sClient,
			Scheme: scheme}

		// Act
		next, result, err := sFnHandleService(context.Background(), &m)

		// Assert
		// we expect error
		require.NotNil(t, err)
		require.ErrorContains(t, err, "sweet-dirac error message")
		// no result because of error
		require.Nil(t, result)
		// no next state (we will stop)
		require.Nil(t, next)
		// function has proper condition
		requireContainsCondition(t, m.State.Function.Status,
			serverlessv1alpha2.ConditionRunning,
			metav1.ConditionFalse,
			serverlessv1alpha2.ConditionReasonServiceFailed,
			"Service mystifying-mayer-name create failed: sweet-dirac error message")
	})
	t.Run("when service exists on kubernetes but we do not need changes should keep it without changes and go to the next state", func(t *testing.T) {
		// Arrange
		f := serverlessv1alpha2.Function{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "funny-sanderson-name",
				Namespace: "wonderful-herschel-ns"}}
		// identical service will be generated inside sFnHandleService
		svc := resources.NewService(&f).Service
		// scheme and fake client
		scheme := runtime.NewScheme()
		require.NoError(t, serverlessv1alpha2.AddToScheme(scheme))
		require.NoError(t, corev1.AddToScheme(scheme))
		createOrUpdateWasCalled := false
		k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(svc).WithInterceptorFuncs(interceptor.Funcs{
			Create: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
				createOrUpdateWasCalled = true
				return nil
			},
			Update: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
				createOrUpdateWasCalled = true
				return nil
			},
		}).Build()
		// machine with our function
		m := fsm.StateMachine{
			State: fsm.SystemState{
				Function: f},
			Log:    zap.NewNop().Sugar(),
			Client: k8sClient,
			Scheme: scheme}

		// Act
		next, result, err := sFnHandleService(context.Background(), &m)

		// Assert
		// no errors
		require.Nil(t, err)
		// without stopping processing
		require.Nil(t, result)
		// with expected next state
		require.NotNil(t, next)
		requireEqualFunc(t, sFnDeploymentStatus, next)
		// service has not been created or updated
		require.False(t, createOrUpdateWasCalled)
		// function conditions remain unchanged
		require.Empty(t, m.State.Function.Status.Conditions)
	})
	t.Run("when service exists on kubernetes and we need changes should update it and requeue", func(t *testing.T) {
		// Arrange
		// service which will be returned from kubernetes - empty so there will be a difference
		svc := corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "wizardly-allen-name",
				Namespace: "kind-tu-ns"}}
		// scheme and fake client
		scheme := runtime.NewScheme()
		require.NoError(t, serverlessv1alpha2.AddToScheme(scheme))
		require.NoError(t, corev1.AddToScheme(scheme))
		createWasCalled := false
		k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&svc).WithInterceptorFuncs(interceptor.Funcs{
			Create: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
				createWasCalled = true
				return nil
			},
		}).Build()
		// machine with our function
		m := fsm.StateMachine{
			State: fsm.SystemState{
				Function: serverlessv1alpha2.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "wizardly-allen-name",
						Namespace: "kind-tu-ns"}}},
			Log:    zap.NewNop().Sugar(),
			Client: k8sClient,
			Scheme: scheme}

		// Act
		next, result, err := sFnHandleService(context.Background(), &m)

		// Assert
		// no errors
		require.Nil(t, err)
		// we expect stop and requeue
		require.NotNil(t, result)
		require.Equal(t, ctrl.Result{Requeue: true}, *result)
		// no next state (we will stop)
		require.Nil(t, next)
		// function has proper condition
		requireContainsCondition(t, m.State.Function.Status,
			serverlessv1alpha2.ConditionRunning,
			metav1.ConditionUnknown,
			serverlessv1alpha2.ConditionReasonServiceUpdated,
			"Service wizardly-allen-name updated")
		// service has not been created (updated only)
		require.False(t, createWasCalled)
		// service has been updated on k8s
		updatedSvc := &corev1.Service{}
		getErr := k8sClient.Get(context.Background(), client.ObjectKey{
			Name:      "wizardly-allen-name",
			Namespace: "kind-tu-ns",
		}, updatedSvc)
		require.NoError(t, getErr)
		// service should have updated some specific fields
		require.Contains(t, updatedSvc.Spec.Selector, serverlessv1alpha2.FunctionNameLabel)
		require.Equal(t, "wizardly-allen-name", updatedSvc.Spec.Selector[serverlessv1alpha2.FunctionNameLabel])
	})
	t.Run("when service exists on kubernetes and update fails should stop processing", func(t *testing.T) {
		// Arrange
		// service which will be returned from kubernetes - empty so there will be a difference
		svc := corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "youthful-gates-name",
				Namespace: "zen-newton-ns"}}
		// scheme and fake client
		scheme := runtime.NewScheme()
		require.NoError(t, serverlessv1alpha2.AddToScheme(scheme))
		require.NoError(t, corev1.AddToScheme(scheme))
		k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&svc).WithInterceptorFuncs(interceptor.Funcs{
			Update: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
				return errors.New("quirky-elion error message")
			},
		}).Build()
		// machine with our function
		m := fsm.StateMachine{
			State: fsm.SystemState{
				Function: serverlessv1alpha2.Function{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "youthful-gates-name",
						Namespace: "zen-newton-ns"}}},
			Log:    zap.NewNop().Sugar(),
			Client: k8sClient,
			Scheme: scheme}

		// Act
		next, result, err := sFnHandleService(context.Background(), &m)

		// Assert
		// we expect error
		require.NotNil(t, err)
		require.ErrorContains(t, err, "quirky-elion error message")
		// no result because of error
		require.Nil(t, result)
		// no next state (we will stop)
		require.Nil(t, next)
		// function has proper condition
		requireContainsCondition(t, m.State.Function.Status,
			serverlessv1alpha2.ConditionRunning,
			metav1.ConditionFalse,
			serverlessv1alpha2.ConditionReasonServiceFailed,
			"Service youthful-gates-name update failed: quirky-elion error message")
	})
}

func Test_serviceChanged(t *testing.T) {
	type args struct {
		a *corev1.Service
		b *corev1.Service
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "when all compared fields are equal should return false",
			args: args{
				a: &corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"gauss":    "sad",
							"williams": "elegant"}},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{{
							Name:        "quizzical-joliot",
							Protocol:    "xenodochial-kepler",
							AppProtocol: ptr.To[string]("boring-keller"),
							Port:        int32(6543),
							TargetPort: intstr.IntOrString{
								StrVal: "amazing-shannon"},
							NodePort: int32(2345)}},
						Selector: map[string]string{
							"ishizaka": "zealous",
							"easley":   "angry"}}},
				b: &corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"williams": "elegant",
							"gauss":    "sad"}},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{{
							Name:        "quizzical-joliot",
							Protocol:    "xenodochial-kepler",
							AppProtocol: ptr.To[string]("boring-keller"),
							Port:        int32(6543),
							TargetPort: intstr.IntOrString{
								StrVal: "amazing-shannon"},
							NodePort: int32(2345)}},
						Selector: map[string]string{
							"easley":   "angry",
							"ishizaka": "zealous"}}},
			},
			want: false,
		},
		{
			name: "when ports are different should return true",
			args: args{
				a: &corev1.Service{
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{{
							Name: "quizzical-joliot"}}}},
				b: &corev1.Service{
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{{
							Name: "elastic-curie"}}}},
			},
			want: true,
		},
		{
			// TODO: why?
			name: "when more than one port should return true",
			args: args{
				a: &corev1.Service{
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{
							{Name: "eager-khorana"},
							{Name: "laughing-swartz"}}}},
				b: &corev1.Service{
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{
							{Name: "eager-khorana"},
							{Name: "laughing-swartz"}}}},
			},
			want: true,
		},
		{
			// TODO: why?
			name: "when less than one port should return true",
			args: args{
				a: &corev1.Service{
					Spec: corev1.ServiceSpec{}},
				b: &corev1.Service{
					Spec: corev1.ServiceSpec{}},
			},
			want: true,
		},
		{
			name: "when labels are different should return true",
			args: args{
				a: &corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"gauss":    "sad",
							"williams": "elegant"}},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{{Name: "festive-williams"}}}},
				b: &corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"williams": "serene",
							"gauss":    "sad"}},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{{Name: "festive-williams"}}}},
			},
			want: true,
		},
		{
			name: "when selectors are different should return true",
			args: args{
				a: &corev1.Service{
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{{Name: "festive-williams"}},
						Selector: map[string]string{
							"easley": "angry"}}},
				b: &corev1.Service{
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{{Name: "festive-williams"}},
						Selector: map[string]string{
							"easley":   "angry",
							"ishizaka": "zealous"}}},
			},
			want: true,
		},
		{
			name: "when not compared fields are different should return false",
			args: args{
				a: &corev1.Service{
					TypeMeta: metav1.TypeMeta{
						Kind:       "gifted-nash",
						APIVersion: "gifted-nash"},
					ObjectMeta: metav1.ObjectMeta{
						Name:                       "gifted-nash",
						GenerateName:               "gifted-nash",
						Namespace:                  "gifted-nash",
						UID:                        "gifted-nash",
						ResourceVersion:            "gifted-nash",
						Generation:                 123,
						CreationTimestamp:          metav1.Time{Time: time.Date(123, 1, 2, 3, 1, 2, 3, &time.Location{})},
						DeletionTimestamp:          &metav1.Time{Time: time.Date(123, 1, 2, 3, 1, 2, 3, &time.Location{})},
						DeletionGracePeriodSeconds: ptr.To[int64](123),
						Annotations:                map[string]string{"nash": "gifted"},
						OwnerReferences:            []metav1.OwnerReference{{Name: "gifted-nash"}},
						Finalizers:                 []string{"gifted-nash"},
						ManagedFields:              []metav1.ManagedFieldsEntry{{Operation: "gifted-nash"}}},
					Spec: corev1.ServiceSpec{
						Ports:                    []corev1.ServicePort{{Name: "festive-williams"}},
						ClusterIP:                "gifted-nash",
						ClusterIPs:               []string{"gifted-nash"},
						Type:                     "gifted-nash",
						ExternalIPs:              []string{"gifted-nash"},
						SessionAffinity:          "gifted-nash",
						LoadBalancerSourceRanges: []string{"gifted-nash"},
						ExternalName:             "gifted-nash",
						ExternalTrafficPolicy:    "gifted-nash",
						HealthCheckNodePort:      123,
						PublishNotReadyAddresses: false,
						SessionAffinityConfig: &corev1.SessionAffinityConfig{
							ClientIP: &corev1.ClientIPConfig{TimeoutSeconds: ptr.To[int32](123)}},
						IPFamilies:                    []corev1.IPFamily{"gifted-nash"},
						IPFamilyPolicy:                ptr.To[corev1.IPFamilyPolicy]("gifted-nash"),
						AllocateLoadBalancerNodePorts: ptr.To[bool](false),
						LoadBalancerClass:             ptr.To[string]("gifted-nash"),
						InternalTrafficPolicy:         ptr.To[corev1.ServiceInternalTrafficPolicy]("gifted-nash"),
						TrafficDistribution:           ptr.To[string]("gifted-nash"),
					},
					Status: corev1.ServiceStatus{
						LoadBalancer: corev1.LoadBalancerStatus{
							Ingress: []corev1.LoadBalancerIngress{{Hostname: "gifted-nash"}},
						},
						Conditions: []metav1.Condition{{
							Type:   "gifted-nash",
							Status: "gifted-nash",
							Reason: "gifted-nash"}}}},
				b: &corev1.Service{
					TypeMeta: metav1.TypeMeta{
						Kind:       "pedantic-bartik",
						APIVersion: "pedantic-bartik"},
					ObjectMeta: metav1.ObjectMeta{
						Name:                       "pedantic-bartik",
						GenerateName:               "pedantic-bartik",
						Namespace:                  "pedantic-bartik",
						UID:                        "pedantic-bartik",
						ResourceVersion:            "pedantic-bartik",
						Generation:                 789,
						CreationTimestamp:          metav1.Time{Time: time.Date(789, 7, 8, 9, 7, 8, 9, &time.Location{})},
						DeletionTimestamp:          &metav1.Time{Time: time.Date(789, 7, 8, 9, 7, 8, 9, &time.Location{})},
						DeletionGracePeriodSeconds: ptr.To[int64](789),
						Annotations:                map[string]string{"bartik": "pedantic"},
						OwnerReferences:            []metav1.OwnerReference{{Name: "pedantic-bartik"}},
						Finalizers:                 []string{"pedantic-bartik"},
						ManagedFields:              []metav1.ManagedFieldsEntry{{Operation: "pedantic-bartik"}}},
					Spec: corev1.ServiceSpec{
						Ports:                    []corev1.ServicePort{{Name: "festive-williams"}},
						ClusterIP:                "pedantic-bartik",
						ClusterIPs:               []string{"pedantic-bartik"},
						Type:                     "pedantic-bartik",
						ExternalIPs:              []string{"pedantic-bartik"},
						SessionAffinity:          "pedantic-bartik",
						LoadBalancerSourceRanges: []string{"pedantic-bartik"},
						ExternalName:             "pedantic-bartik",
						ExternalTrafficPolicy:    "pedantic-bartik",
						HealthCheckNodePort:      789,
						PublishNotReadyAddresses: true,
						SessionAffinityConfig: &corev1.SessionAffinityConfig{
							ClientIP: &corev1.ClientIPConfig{TimeoutSeconds: ptr.To[int32](789)}},
						IPFamilies:                    []corev1.IPFamily{"pedantic-bartik"},
						IPFamilyPolicy:                ptr.To[corev1.IPFamilyPolicy]("pedantic-bartik"),
						AllocateLoadBalancerNodePorts: ptr.To[bool](true),
						LoadBalancerClass:             ptr.To[string]("pedantic-bartik"),
						InternalTrafficPolicy:         ptr.To[corev1.ServiceInternalTrafficPolicy]("pedantic-bartik"),
						TrafficDistribution:           ptr.To[string]("pedantic-bartik"),
					},
					Status: corev1.ServiceStatus{
						LoadBalancer: corev1.LoadBalancerStatus{
							Ingress: []corev1.LoadBalancerIngress{{Hostname: "pedantic-bartik"}},
						},
						Conditions: []metav1.Condition{{
							Type:   "pedantic-bartik",
							Status: "pedantic-bartik",
							Reason: "pedantic-bartik"}}}},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := serviceChanged(tt.args.a, tt.args.b)
			require.Equal(t, tt.want, r)
		})
	}
}
