package capability

import (
	"strings"

	"github.com/opdev/opcap/internal/operator"

	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
)

// Audit defines all the methods used to run a full audit Plan against a single operator
// All new capability tests should be added to this interface and used with a capAudit
// instance and as part of an auditPlan
type Audit interface {
	OperatorInstall() error
	OperandInstall() error
	OperandCleanUp() error
	OperatorCleanUp() error
}

// capAudit is an implementation of the Audit interface
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

	// auditPlan is a list of functions to be run in sequence in a given audit
	// all of them must be an implemented method of capAudit and must be part
	// of the Audit interface
	auditPlan []string

	// customResources is a map of string interface that has all the CR(almExamples) that needs to be installed
	// as part of the OperandInstall function
	customResources []map[string]interface{}
}

func newCapAudit(c operator.Client, subscription operator.SubscriptionData, auditPlan []string) (capAudit, error) {

	ns := strings.Join([]string{"opcap", strings.ReplaceAll(subscription.Package, ".", "-")}, "-")
	operatorGroupName := strings.Join([]string{subscription.Name, subscription.Channel, "group"}, "-")

	ocpVersion, err := c.GetOpenShiftVersion()
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
