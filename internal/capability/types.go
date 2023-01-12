package capability

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/opdev/opcap/internal/operator"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/spf13/afero"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type auditOptions struct {
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
	reportWriter      io.Writer
	csvEvents         *corev1.EventList
	detailedReports   bool
}

type auditorOptions struct {
	// AuditPlan holds the tests that should be run during an audit
	auditPlan []string

	// CatalogSource may be built-in OLM or custom
	catalogSource string
	// CatalogSourceNamespace will be openshift-marketplace or custom
	catalogSourceNamespace string

	// Packages is a subset of packages to be tested from a catalogSource
	packages []string

	// WorkQueue holds capAudits in a buffered channel in order to execute them
	workQueue chan capAudit

	// AllInstallModes will test all install modes supported by an operator
	allInstallModes bool

	// extraCustomResources associates packages to a list of Custom Resources (in addition to ALMExamples)
	// to be audited by the OperandInstall AuditPlan.
	extraCustomResources string

	// OpCapClient is the main OpenShift client interface
	opCapClient operator.Client

	// Fs is an afero filesystem used by the auditor
	fs afero.Fs

	// Timeout is the audit timeout
	timeout time.Duration

	//  ReportWriter is any io.Writer for the text reports
	reportWriter io.Writer

	// DetailedReports creates reports containing events and logs
	detailedReports bool
}

type (
	auditFn        func(context.Context) error
	auditCleanupFn func(context.Context) error

	// auditOption is the function type for passing an option to an audit
	auditOption func(options *auditOptions) error

	// auditorOption is the function type for passing an option to RunAudits
	auditorOption func(options *auditorOptions) error
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
