package capability

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/opdev/opcap/internal/operator"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
)

// CapAudit is an implementation of the Audit interface
type capAudit struct {
	// client has access to all operator methods
	client operator.Client

	// OpenShift Cluster Version under test
	ocpVersion string

	// namespace is the ns where the operator will be installed
	namespace string

	// operatorGroupData contains information to create operator groups
	operatorGroupData operator.OperatorGroupData

	// subscription holds the data to install an operator via OLM
	subscription operator.SubscriptionData

	// Cluster CSV for current operator under test
	csv operatorv1alpha1.ClusterServiceVersion

	// How much time to wait for a CSV before timeout
	csvWaitTime time.Duration

	// If the given CSV timed out on install
	csvTimeout bool

	// auditPlan is a list of functions to be run in sequence in a given audit
	// all of them must be an implemented method of CapAudit and must be part
	// of the Audit interface
	auditPlan []string

	// CustomResources stores CR manifests to deploy operands
	customResources []map[string]interface{}

	// Operands stores a list of unstructured custom resources that were created at the API level
	// This data is used for further analysis on statuses, conditions and other patterns
	operands []unstructured.Unstructured
}

func newCapAudit(ctx context.Context, c operator.Client, subscription operator.SubscriptionData, auditPlan []string, extraCustomResources []map[string]interface{}) (*capAudit, error) {
	ns := strings.Join([]string{"opcap", strings.ReplaceAll(subscription.Package, ".", "-"), strings.ToLower(string(subscription.InstallModeType))}, "-")
	operatorGroupName := strings.Join([]string{subscription.Name, subscription.Channel, "group"}, "-")

	ocpVersion, err := c.GetOpenShiftVersion(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get OpenShift version for testing: %v", err)
	}

	return &capAudit{
		client:            c,
		ocpVersion:        ocpVersion,
		namespace:         ns,
		operatorGroupData: newOperatorGroupData(operatorGroupName, getTargetNamespaces(subscription, ns)),
		subscription:      subscription,
		csvWaitTime:       time.Minute,
		csvTimeout:        false,
		auditPlan:         auditPlan,
		customResources:   extraCustomResources,
	}, nil
}

func newOperatorGroupData(name string, targetNamespaces []string) operator.OperatorGroupData {
	return operator.OperatorGroupData{
		Name:             name,
		TargetNamespaces: targetNamespaces,
	}
}

func getTargetNamespaces(s operator.SubscriptionData, namespace string) []string {
	targetNs1 := strings.Join([]string{namespace, "targetns1"}, "-")
	targetNs2 := strings.Join([]string{namespace, "targetns2"}, "-")

	switch s.InstallModeType {

	case operatorv1alpha1.InstallModeTypeSingleNamespace:

		return []string{targetNs1}

	case operatorv1alpha1.InstallModeTypeOwnNamespace:
		return []string{namespace}

	case operatorv1alpha1.InstallModeTypeMultiNamespace:

		return []string{targetNs1, targetNs2}
	}
	return []string{}
}

// auditOption is the function type for passing an option to an audit
type auditOption func(options *options) error

// WithSubscription adds a subscription object to the audit
func withSubscription(subscription *operator.SubscriptionData) auditOption {
	return func(options *options) error {
		if subscription == nil {
			return fmt.Errorf("subscription data cannot be nil")
		}
		options.Subscription = subscription
		return nil
	}
}

// WithOperatorGroupData adds an operatorgroupdata objec to the audit
func withOperatorGroupData(operatorGroupData *operator.OperatorGroupData) auditOption {
	return func(options *options) error {
		if operatorGroupData == nil {
			return fmt.Errorf("operatorgroupdata cannot be nil")
		}
		options.operatorGroupData = operatorGroupData
		return nil
	}
}

// WithNamespace adds a namespace for the audit
func withNamespace(namespace string) auditOption {
	return func(options *options) error {
		if namespace == "" {
			return fmt.Errorf("namespace cannot be empty")
		}
		options.namespace = namespace
		return nil
	}
}

// WithClient adds a client to the audit
func withClient(client operator.Client) auditOption {
	return func(options *options) error {
		if client == nil {
			return fmt.Errorf("client cannot be nil")
		}
		options.client = client
		return nil
	}
}

// WithTimeout adds a timeout duration to the audit
func withTimeout(csvWaitTime int) auditOption {
	return func(options *options) error {
		options.csvWaitTime = time.Duration(csvWaitTime)
		return nil
	}
}

// WithOcpVersion adds the OCP version to the audit
func withOcpVersion(ocpVersion string) auditOption {
	return func(options *options) error {
		options.OcpVersion = ocpVersion
		return nil
	}
}

// withCustomResources adds existing Custom Resources to the audit
func withCustomResources(customResources []map[string]interface{}) auditOption {
	return func(options *options) error {
		options.customResources = customResources
		return nil
	}
}

type options struct {
	Subscription      *operator.SubscriptionData
	operatorGroupData *operator.OperatorGroupData
	namespace         string
	client            operator.Client
	CsvTimeout        bool
	csvWaitTime       time.Duration
	Csv               v1alpha1.ClusterServiceVersion
	OcpVersion        string
	customResources   []map[string]interface{}
	operands          []unstructured.Unstructured
}

type auditFn func(context.Context) error

// New returns a function corresponding to a passed in audit plan
func newAudit(ctx context.Context, auditType string, opts ...auditOption) auditFn {
	switch strings.ToLower(auditType) {
	case "operatorinstall":
		return operatorInstall(ctx, opts...)
	case "operatorcleanup":
		return operatorCleanUp(ctx, opts...)
	case "operandinstall":
		return operandInstall(ctx, opts...)
	case "operandcleanup":
		return operandCleanUp(ctx, opts...)
	}
	return nil
}
