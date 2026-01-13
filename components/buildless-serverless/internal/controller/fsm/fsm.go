package fsm

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"time"

	serverlessv1alpha2 "github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/config"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/git"
	serverlessmetrics "github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/metrics"
	"github.com/kyma-project/serverless/components/buildless-serverless/internal/controller/resources"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	apimachineryruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type StateFn func(context.Context, *StateMachine) (StateFn, *ctrl.Result, error)

type SystemState struct {
	Function          serverlessv1alpha2.Function
	statusSnapshot    serverlessv1alpha2.FunctionStatus
	BuiltDeployment   *resources.Deployment
	ClusterDeployment *appsv1.Deployment
	Commit            string
	GitAuth           *git.GitAuth
}

func (s *SystemState) saveStatusSnapshot() {
	result := s.Function.Status.DeepCopy()
	if result == nil {
		result = &serverlessv1alpha2.FunctionStatus{}
	}
	s.statusSnapshot = *result
}

type StateMachine struct {
	nextFn         StateFn
	State          SystemState
	Log            *zap.SugaredLogger
	Client         client.Client
	FunctionConfig config.FunctionConfig
	Scheme         *apimachineryruntime.Scheme
	GitChecker     git.AsyncLatestCommitChecker
	EventRecorder  record.EventRecorder
}

func (m *StateMachine) stateFnName() string {
	fullName := runtime.FuncForPC(reflect.ValueOf(m.nextFn).Pointer()).Name()
	splitFullName := strings.Split(fullName, ".")

	if len(splitFullName) < 3 {
		return fullName
	}

	shortName := splitFullName[2]
	return shortName
}

func (m *StateMachine) Reconcile(ctx context.Context) (ctrl.Result, error) {
	var startReconciliationTime = time.Now()
	var err error
	var result *ctrl.Result
loop:
	for m.nextFn != nil && err == nil {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			break loop

		default:
			m.Log.Info(fmt.Sprintf("switching state: %s", m.stateFnName()))
			m.nextFn, result, err = m.nextFn(ctx, m)
			if updateErr := updateFunctionStatus(ctx, m); updateErr != nil {
				err = updateErr
			}
		}
	}

	if result == nil {
		result = &ctrl.Result{}
	}

	m.Log.
		With("error", err).
		With("result", result).
		Info("reconciliation done")
	serverlessmetrics.PublishReconciliationTime(m.State.Function, startReconciliationTime)

	return *result, err
}

type StateMachineReconciler interface {
	Reconcile(ctx context.Context) (ctrl.Result, error)
}

func New(client client.Client, functionConfig config.FunctionConfig, instance *serverlessv1alpha2.Function, startState StateFn, recorder record.EventRecorder, gitChecker git.AsyncLatestCommitChecker, scheme *apimachineryruntime.Scheme, log *zap.SugaredLogger) StateMachineReconciler {
	sm := StateMachine{
		nextFn: startState,
		State: SystemState{
			Function: *instance,
		},
		Log:            log,
		FunctionConfig: functionConfig,
		Client:         client,
		Scheme:         scheme,
		GitChecker:     gitChecker,
		EventRecorder:  recorder,
	}
	sm.State.saveStatusSnapshot()
	return &sm
}

func updateFunctionStatus(ctx context.Context, m *StateMachine) error {
	s := &m.State
	if !reflect.DeepEqual(s.Function.Status, s.statusSnapshot) {
		m.Log.Debug(fmt.Sprintf("updating serverless status to '%+v'", s.Function.Status))
		err := m.Client.Status().Update(ctx, &s.Function)
		emitEvent(m)
		s.saveStatusSnapshot()
		return err
	}
	return nil
}
