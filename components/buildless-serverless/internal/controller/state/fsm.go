package state

import (
	"context"
	"fmt"
	serverlessv1alpha2 "github.com/kyma-project/serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/internal/config"
	"reflect"
	"runtime"
	"strings"

	"go.uber.org/zap"
	apimachineryruntime "k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type stateFn func(context.Context, *stateMachine) (stateFn, *ctrl.Result, error)

type systemState struct {
	instance       serverlessv1alpha2.Function
	statusSnapshot serverlessv1alpha2.FunctionStatus
}

func (s *systemState) saveStatusSnapshot() {
	result := s.instance.Status.DeepCopy()
	if result == nil {
		result = &serverlessv1alpha2.FunctionStatus{}
	}
	s.statusSnapshot = *result
}

type stateMachine struct {
	nextFn stateFn
	state  systemState
	log    *zap.SugaredLogger
	//	k8s
	client         client.Client
	functionConfig config.FunctionConfig
	scheme         *apimachineryruntime.Scheme
}

func (m *stateMachine) stateFnName() string {
	fullName := runtime.FuncForPC(reflect.ValueOf(m.nextFn).Pointer()).Name()
	splitFullName := strings.Split(fullName, ".")

	if len(splitFullName) < 3 {
		return fullName
	}

	shortName := splitFullName[2]
	return shortName
}

func (m *stateMachine) Reconcile(ctx context.Context) (ctrl.Result, error) {
	var err error
	var result *ctrl.Result
loop:
	for m.nextFn != nil && err == nil {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			break loop

		default:
			m.log.Info(fmt.Sprintf("switching state: %s", m.stateFnName()))
			m.nextFn, result, err = m.nextFn(ctx, m)
			if updateErr := updateFunctionStatus(ctx, m); updateErr != nil {
				err = updateErr
			}
		}
	}

	if result == nil {
		result = &ctrl.Result{}
	}

	m.log.
		With("error", err).
		With("result", result).
		Info("reconciliation done")

	return *result, err
}

type StateMachineReconciler interface {
	Reconcile(ctx context.Context) (ctrl.Result, error)
}

func NewMachine(client client.Client, functionConfig config.FunctionConfig, instance *serverlessv1alpha2.Function /*recorder record.EventRecorder,*/, scheme *apimachineryruntime.Scheme, log *zap.SugaredLogger) StateMachineReconciler {
	sm := stateMachine{
		nextFn: sFnStart,
		state: systemState{
			instance: *instance,
		},
		log:            log,
		functionConfig: functionConfig,
		client:         client,
		scheme:         scheme,
	}
	sm.state.saveStatusSnapshot()
	return &sm
}

func updateFunctionStatus(ctx context.Context, m *stateMachine) error {
	s := &m.state
	if !reflect.DeepEqual(s.instance.Status, s.statusSnapshot) {
		m.log.Debug(fmt.Sprintf("updating serverless status to '%+v'", s.instance.Status))
		err := m.client.Status().Update(ctx, &s.instance)
		//emitEvent(r, s)
		s.saveStatusSnapshot()
		return err
	}
	return nil
}
