package capability

import (
	"context"
	"errors"
	"time"

	"github.com/opdev/opcap/internal/operator"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/spf13/afero"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type options struct {
	subscription      *operator.SubscriptionData
	operatorGroupData *operator.OperatorGroupData
	namespace         string
	client            operator.Client
	csvTimeout        bool
	csvWaitTime       time.Duration
	csv               *v1alpha1.ClusterServiceVersion
	ocpVersion        string
	customResources   []map[string]interface{}
	operands          []unstructured.Unstructured
	fs                afero.Fs
}

type (
	auditFn        func(context.Context) error
	auditCleanupFn func(context.Context) error
)

type Stack[T any] struct {
	stack *element[T]
}

type element[T any] struct {
	previous *element[T]
	val      T
}

var StackEmptyError = errors.New("Stack empty")

func (s *Stack[T]) Push(v T) {
	e := &element[T]{
		previous: s.stack,
		val:      v,
	}
	s.stack = e
}

func (s *Stack[T]) Pop() (T, error) {
	if s.stack == nil {
		var r T
		return r, StackEmptyError
	}
	e := *s.stack
	s.stack = e.previous

	return e.val, nil
}

func (s *Stack[T]) Empty() bool {
	return s.stack == nil
}
