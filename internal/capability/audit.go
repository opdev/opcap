package capability

import (
	"context"
	"strings"
	"time"

	"github.com/opdev/opcap/internal/logger"
	"github.com/opdev/opcap/internal/operator"
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

func newCapAudit(ctx context.Context, c operator.Client, subscription operator.SubscriptionData, auditPlan []string) (capAudit, error) {
	ns := strings.Join([]string{"opcap", strings.ReplaceAll(subscription.Package, ".", "-")}, "-")
	operatorGroupName := strings.Join([]string{subscription.Name, subscription.Channel, "group"}, "-")

	ocpVersion, err := c.GetOpenShiftVersion(ctx)
	if err != nil {
		logger.Debugw("Couldn't get OpenShift version for testing", "Err:", err)
		return capAudit{}, err
	}

	return capAudit{
		client:            c,
		ocpVersion:        ocpVersion,
		namespace:         ns,
		operatorGroupData: newOperatorGroupData(operatorGroupName, getTargetNamespaces(subscription, ns)),
		subscription:      subscription,
		csvWaitTime:       time.Minute,
		csvTimeout:        false,
		auditPlan:         auditPlan,
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
