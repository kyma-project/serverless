package fsm

import (
	"context"
	"fmt"
	serverlessv1alpha2 "github.com/kyma-project/serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/internal/config"
	"github.com/kyma-project/serverless/internal/controller/git"
	"github.com/kyma-project/serverless/internal/controller/resources"
	appsv1 "k8s.io/api/apps/v1"
	"reflect"
	"runtime"
	"strings"

	"go.uber.org/zap"
	apimachineryruntime "k8s.io/apimachinery/pkg/runtime"
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
	GitChecker     git.LastCommitChecker
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

	return *result, err
}

type StateMachineReconciler interface {
	Reconcile(ctx context.Context) (ctrl.Result, error)
}

// TODO: Add emiting events
func New(client client.Client, functionConfig config.FunctionConfig, instance *serverlessv1alpha2.Function, startState StateFn /*recorder record.EventRecorder,*/, scheme *apimachineryruntime.Scheme, log *zap.SugaredLogger) StateMachineReconciler {
	sm := StateMachine{
		nextFn: startState,
		State: SystemState{
			Function: *instance,
		},
		Log:            log,
		FunctionConfig: functionConfig,
		Client:         client,
		Scheme:         scheme,
		GitChecker:     git.GoGitCommitChecker{},
	}
	sm.State.saveStatusSnapshot()
	return &sm
}

func updateFunctionStatus(ctx context.Context, m *StateMachine) error {
	s := &m.State
	if !reflect.DeepEqual(s.Function.Status, s.statusSnapshot) {
		m.Log.Debug(fmt.Sprintf("updating serverless status to '%+v'", s.Function.Status))
		err := m.Client.Status().Update(ctx, &s.Function)
		//emitEvent(r, s)
		s.saveStatusSnapshot()
		return err
	}
	return nil
}
