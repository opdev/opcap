package capability

import (
	"context"
	"time"

	"github.com/opdev/opcap/internal/operator"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type options struct {
	Subscription      *operator.SubscriptionData
	operatorGroupData *operator.OperatorGroupData
	namespace         string
	client            operator.Client
	CsvTimeout        bool
	csvWaitTime       time.Duration
	Csv               *v1alpha1.ClusterServiceVersion
	OcpVersion        string
	customResources   []map[string]interface{}
	operands          []unstructured.Unstructured
}

type (
	auditFn        func(context.Context) error
	auditCleanupFn func(context.Context) error
)
