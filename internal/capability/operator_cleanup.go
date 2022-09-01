package capability

import (
	"context"

	"github.com/opdev/opcap/internal/operator"
)

func (ca *capAudit) OperatorCleanUp() error {

	// delete subscription
	if err := ca.client.DeleteSubscription(context.Background(), ca.subscription.Name, ca.namespace); err != nil {
		logger.Debugf("Error while deleting Subscription: %w", err)
		return err
	}

	// delete operator group
	if err := ca.client.DeleteOperatorGroup(context.Background(), ca.operatorGroupData.Name, ca.namespace); err != nil {
		logger.Debugf("Error while deleting OperatorGroup: %w", err)
		return err
	}

	// delete target namespaces
	for _, ns := range ca.operatorGroupData.TargetNamespaces {
		if err := operator.DeleteNamespace(context.Background(), ns); err != nil {
			logger.Debugf("Error deleting target namespace %s", ns)
			return err
		}
	}

	// delete operator's own namespace
	if err := operator.DeleteNamespace(context.Background(), ca.namespace); err != nil {
		logger.Debugf("Error deleting operator's own namespace %s", ca.namespace)
		return err
	}
	return nil
}
