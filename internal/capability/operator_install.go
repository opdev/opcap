package capability

import (
	"context"
	"strings"

	log "opcap/internal/logger"
	"opcap/internal/operator"

	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
)

var logger = log.Sugar

// TODO: InstallOperatorsTest creates all subscriptions for a catalogSource sequencially
// We will need other arguments that can tweak how many to test at a time
// And possibly indicate a specific condition

type operatorData struct {
	namespace     string
	targetNs1     string
	targetNs2     string
	operatorGroup string
	installedNS   []string
}

func (ca *capAudit) OperatorInstall() error {
	logger.Debugw("installing package", "package", ca.subscription.Package, "channel", ca.subscription.Channel, "installmode", ca.subscription.InstallModeType)

	od := new(operatorData)

	od.namespace = strings.Join([]string{"opcap", strings.ReplaceAll(ca.subscription.Package, ".", "-")}, "-")
	od.targetNs1 = strings.Join([]string{od.namespace, "targetns1"}, "-")
	od.targetNs2 = strings.Join([]string{od.namespace, "targetns2"}, "-")
	od.operatorGroup = strings.Join([]string{ca.subscription.Name, ca.subscription.Channel, "group"}, "-")

	// create operator namespace
	operator.CreateNamespace(context.Background(), od.namespace)

	// Checking install modes and
	// creating operatorGroup per operator package/channel
	od.installedNS = []string{od.namespace}
	createGroupByInstallMode(ca.subscription, ca.client, *od)

	// create subscription per operator package/channel
	sub, err := ca.client.CreateSubscription(context.Background(), ca.subscription, od.namespace)
	if err != nil {
		logger.Debugf("Error creating subscriptions: %w", err)
		return err
	}

	if err = ca.client.WaitForInstallPlan(context.Background(), sub); err != nil {
		logger.Debugf("Waiting for InstallPlan: %w", err)
		return err
	}
	// check/approve install plan
	// TODO: check the name standard for installPlan
	err = ca.client.InstallPlanApprove(od.namespace)
	if err != nil {
		logger.Debugf("Error creating subscriptions: %w", err)
		return err
	}

	csvStatus, err := ca.client.WaitForCsvOnNamespace(od.namespace)

	if err != nil {
		logger.Infow("failed", "package", ca.subscription.Package, "channel", ca.subscription.Channel, "installmode", ca.subscription.InstallModeType)
	} else {
		logger.Infow(strings.ToLower(csvStatus), "package", ca.subscription.Package, "channel", ca.subscription.Channel, "installmode", ca.subscription.InstallModeType)
	}

	cleanUp(ca.subscription, ca.client, *od)

	return nil
}

func createGroupByInstallMode(s operator.SubscriptionData, c operator.Client, m operatorData) {

	switch s.InstallModeType {

	case operatorv1alpha1.InstallModeTypeAllNamespaces:
		opGroupData := operator.OperatorGroupData{
			Name:             m.operatorGroup,
			TargetNamespaces: []string{},
		}
		c.CreateOperatorGroup(context.Background(), opGroupData, m.namespace)

	case operatorv1alpha1.InstallModeTypeSingleNamespace:

		operator.CreateNamespace(context.Background(), m.targetNs1)
		opGroupData := operator.OperatorGroupData{
			Name:             m.operatorGroup,
			TargetNamespaces: []string{m.targetNs1},
		}
		m.installedNS = append(m.installedNS, m.targetNs1)
		c.CreateOperatorGroup(context.Background(), opGroupData, m.namespace)

	case operatorv1alpha1.InstallModeTypeOwnNamespace:
		opGroupData := operator.OperatorGroupData{
			Name:             m.operatorGroup,
			TargetNamespaces: []string{m.namespace},
		}
		c.CreateOperatorGroup(context.Background(), opGroupData, m.namespace)

	case operatorv1alpha1.InstallModeTypeMultiNamespace:

		operator.CreateNamespace(context.Background(), m.targetNs1)
		operator.CreateNamespace(context.Background(), m.targetNs2)
		opGroupData := operator.OperatorGroupData{
			Name:             m.operatorGroup,
			TargetNamespaces: []string{m.targetNs1, m.targetNs2},
		}
		m.installedNS = append(m.installedNS, m.targetNs1)
		m.installedNS = append(m.installedNS, m.targetNs2)
		c.CreateOperatorGroup(context.Background(), opGroupData, m.namespace)

	}

}

func cleanUp(s operator.SubscriptionData, c operator.Client, m operatorData) {

	// delete subscription
	err := c.DeleteSubscription(context.Background(), s.Name, m.namespace)
	if err != nil {
		logger.Debugf("Error while deleting Subscription: %w", err)
		return
	}

	// delete operator group
	err = c.DeleteOperatorGroup(context.Background(), m.operatorGroup, m.namespace)
	if err != nil {
		logger.Debugf("Error while deleting OperatorGroup: %w", err)
		return
	}

	// delete namespaces
	for _, ns := range m.installedNS {
		operator.DeleteNamespace(context.Background(), ns)
	}
}
