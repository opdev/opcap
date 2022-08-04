package capability

import (
	"context"

	"github.com/opdev/opcap/internal/operator"
)

func (ca *CapAudit) OperatorCleanUp() error {

	// delete subscription
	err := ca.Client.DeleteSubscription(context.Background(), ca.Subscription.Name, ca.Namespace)
	if err != nil {
		logger.Debugf("Error while deleting Subscription: %w", err)
		return err
	}

	// delete operator group
	err = ca.Client.DeleteOperatorGroup(context.Background(), ca.OperatorGroupData.Name, ca.Namespace)
	if err != nil {
		logger.Debugf("Error while deleting OperatorGroup: %w", err)
		return err
	}

	// delete target namespaces
	for _, ns := range ca.OperatorGroupData.TargetNamespaces {
		err := operator.DeleteNamespace(context.Background(), ns)
		if err != nil {
			logger.Debugf("Error deleting target namespace %s", ns)
			return err
		}
	}

	// delete operator's own namespace
	err = operator.DeleteNamespace(context.Background(), ca.Namespace)
	if err != nil {
		logger.Debugf("Error deleting operator's own namespace %s", ca.Namespace)
		return err
	}
	return nil
}
