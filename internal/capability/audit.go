package capability

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/opdev/opcap/internal/operator"
	"github.com/spf13/afero"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

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

func generateNamespace(packageName string, installMode string) string {
	installModeString := string(installMode)

	packageNameMaxLength := 63 - len("opcap-") - len(installModeString) - 1

	if len(packageName) > packageNameMaxLength {
		packageName = packageName[:packageNameMaxLength]
	}

	return strings.Join([]string{
		"opcap",
		packageName,
		installModeString,
	}, "-")
}

func newCapAudit(ctx context.Context, c operator.Client, subscription operator.SubscriptionData, auditPlan []string, extraCustomResources []map[string]interface{}) (*capAudit, error) {
	ns := generateNamespace(strings.ReplaceAll(subscription.Package, ".", "-"), strings.ToLower(string(subscription.InstallModeType)))
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
		csvWaitTime:       2 * time.Minute,
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

// WithSubscription adds a subscription object to the audit
func withSubscription(subscription *operator.SubscriptionData) auditOption {
	return func(options *auditOptions) error {
		if subscription == nil {
			return fmt.Errorf("subscription data cannot be nil")
		}
		options.subscription = subscription
		return nil
	}
}

// WithOperatorGroupData adds an operatorgroupdata objec to the audit
func withOperatorGroupData(operatorGroupData *operator.OperatorGroupData) auditOption {
	return func(options *auditOptions) error {
		if operatorGroupData == nil {
			return fmt.Errorf("operatorgroupdata cannot be nil")
		}
		options.operatorGroupData = operatorGroupData
		return nil
	}
}

// WithNamespace adds a namespace for the audit
func withNamespace(namespace string) auditOption {
	return func(options *auditOptions) error {
		if namespace == "" {
			return fmt.Errorf("namespace cannot be empty")
		}
		options.namespace = namespace
		return nil
	}
}

// WithClient adds a client to the audit
func withClient(client operator.Client) auditOption {
	return func(options *auditOptions) error {
		if client == nil {
			return fmt.Errorf("client cannot be nil")
		}
		options.client = client
		return nil
	}
}

// WithTimeout adds a timeout duration to the audit
func withTimeout(csvWaitTime time.Duration) auditOption {
	return func(options *auditOptions) error {
		options.csvWaitTime = csvWaitTime
		return nil
	}
}

// WithOcpVersion adds the OCP version to the audit
func withOcpVersion(ocpVersion string) auditOption {
	return func(options *auditOptions) error {
		options.ocpVersion = ocpVersion
		return nil
	}
}

// withCustomResources adds existing Custom Resources to the audit
func withCustomResources(customResources []map[string]interface{}) auditOption {
	return func(options *auditOptions) error {
		options.customResources = customResources
		return nil
	}
}

// withFilesystem adds a filesystem to be used for writing files
func withFilesystem(fs afero.Fs) auditOption {
	return func(options *auditOptions) error {
		options.fs = fs
		return nil
	}
}

// withReportWriter adds an io.Writer to be used for outputing the test reports
func withReportWriter(w io.Writer) auditOption {
	return func(options *auditOptions) error {
		if w == nil {
			return fmt.Errorf("report writer cannot be nil")
		}
		options.reportWriter = w
		return nil
	}
}

func withDetailedReports(detailedReports bool) auditOption {
	return func(options *auditOptions) error {
		options.detailedReports = detailedReports
		return nil
	}
}

// New returns a function corresponding to a passed in audit plan
func newAudit(ctx context.Context, auditType string, opts ...auditOption) (auditFn, auditCleanupFn) {
	switch strings.ToLower(auditType) {
	case "operatorinstall":
		return operatorInstall(ctx, opts...)
	case "operandinstall":
		return operandInstall(ctx, opts...)
	case "fakeplan":
		return func(ctx context.Context) error { return nil }, func(ctx context.Context) error { return nil }
	}
	return nil, nil
}
