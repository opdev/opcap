package capability

import (
	"strings"

	"github.com/opdev/opcap/internal/operator"

	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
)

// Audit defines all the methods used to run a full audit Plan against a single operator
// All new capability tests should be added to this interface and used with a CapAudit
// instance and as part of an auditPlan
type Audit interface {
	OperatorInstall() error
	OperandInstall() error
	OperandCleanUp() error
	OperatorCleanUp() error
	Report() error
}

// CapAudit is an implementation of the Audit interface
type CapAudit struct {

	// client has access to all operator methods
	Client operator.Client

	// OpenShift Cluster Version under test
	OcpVersion string

	// namespace is the ns where the operator will be installed
	Namespace string

	// operatorGroupData contains information to create operator groups
	OperatorGroupData operator.OperatorGroupData

	// subscription holds the data to install an operator via OLM
	Subscription operator.SubscriptionData

	// Cluster CSV for current operator under test
	Csv operatorv1alpha1.ClusterServiceVersion

	// auditPlan is a list of functions to be run in sequence in a given audit
	// all of them must be an implemented method of CapAudit and must be part
	// of the Audit interface
	AuditPlan []string

	// customResources is a map of string interface that has all the CR(almExamples) that needs to be installed
	// as part of the OperandInstall function
	CustomResources []map[string]interface{}
}

func newCapAudit(c operator.Client, subscription operator.SubscriptionData, auditPlan []string) (CapAudit, error) {

	ns := strings.Join([]string{"opcap", strings.ReplaceAll(subscription.Package, ".", "-")}, "-")
	operatorGroupName := strings.Join([]string{subscription.Name, subscription.Channel, "group"}, "-")

	ocpVersion, err := c.GetOpenShiftVersion()
	if err != nil {
		logger.Debugw("Couldn't get OpenShift version for testing", "Err:", err)
		return CapAudit{}, err
	}

	return CapAudit{
		Client:            c,
		OcpVersion:        ocpVersion,
		Namespace:         ns,
		OperatorGroupData: newOperatorGroupData(operatorGroupName, getTargetNamespaces(subscription, ns)),
		Subscription:      subscription,
		AuditPlan:         auditPlan,
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
