package capability

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/opdev/opcap/internal/operator"

	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Audit defines all the methods used to run a full audit Plan against a single operator
// All new capability tests should be added to this interface and used with a capAudit
// instance and as part of an auditPlan
type Audit interface {
	OperatorInstall() error
	GetAlmExamples() error
	OperandInstall() error
	OperatorCleanUp() error
}

// capAudit is an implementation of the Audit interface
type capAudit struct {

	// client has access to all operator methods
	client operator.Client

	// namespace is the ns where the operator will be installed
	namespace string

	// operatorGroupData contains information to create operator groups
	operatorGroupData operator.OperatorGroupData

	// subscription holds the data to install an operator via OLM
	subscription operator.SubscriptionData

	// auditPlan is a list of functions to be run in sequence in a given audit
	// all of them must be an implemented method of capAudit and must be part
	// of the Audit interface
	auditPlan []string

	// customResources is a map of string interface that has all the CR(almExamples) that needs to be installed
	// as part of the OperandInstall function
	customResources []map[string]interface{}
}

func newCapAudit(c operator.Client, subscription operator.SubscriptionData, auditPlan []string) capAudit {

	ns := strings.Join([]string{"opcap", strings.ReplaceAll(subscription.Package, ".", "-")}, "-")
	operatorGroupName := strings.Join([]string{subscription.Name, subscription.Channel, "group"}, "-")

	return capAudit{
		client:            c,
		namespace:         ns,
		operatorGroupData: newOperatorGroupData(operatorGroupName, getTargetNamespaces(subscription, ns)),
		subscription:      subscription,
		auditPlan:         auditPlan,
	}
}

func newOperatorGroupData(name string, targetNamespaces []string) operator.OperatorGroupData {

	return operator.OperatorGroupData{
		Name:             name,
		TargetNamespaces: targetNamespaces,
	}
}

// function to get all the ALMExamples present in a given CSV
func (ca *capAudit) GetAlmExamples() error {
	ctx := context.Background()

	olmClientset, err := operator.NewOlmClientset()
	if err != nil {
		return err
	}

	opts := v1.ListOptions{}

	// gets the list of CSVs present in a particular namespace
	CSVList, err := olmClientset.OperatorsV1alpha1().ClusterServiceVersions(ca.namespace).List(ctx, opts)
	if err != nil {
		return err
	}

	// map of string interface which consist of ALM examples from the CSVList
	almExamples := CSVList.Items[0].ObjectMeta.Annotations["alm-examples"]

	var almList []map[string]interface{}

	err = json.Unmarshal([]byte(almExamples), &almList)
	if err != nil {
		return err
	}

	ca.customResources = almList

	return nil
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
