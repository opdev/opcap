package capability

import (
	"context"
	"opcap/internal/operator"
)

func (ca *capAudit) OperatorCleanUp() error {

	// delete subscription
	err := ca.client.DeleteSubscription(context.Background(), ca.subscription.Name, ca.namespace)
	if err != nil {
		logger.Debugf("Error while deleting Subscription: %w", err)
		return err
	}

	// delete operator group
	err = ca.client.DeleteOperatorGroup(context.Background(), ca.operatorGroupData.Name, ca.namespace)
	if err != nil {
		logger.Debugf("Error while deleting OperatorGroup: %w", err)
		return err
	}

	// delete target namespaces
	for _, ns := range ca.operatorGroupData.TargetNamespaces {
		err := operator.DeleteNamespace(context.Background(), ns)
		if err != nil {
			logger.Debugf("Error deleting target namespace %s", ns)
			return err
		}
	}

	// delete operator's own namespace
	err = operator.DeleteNamespace(context.Background(), ca.namespace)
	if err != nil {
		logger.Debugf("Error deleting operator's own namespace %s", ca.namespace)
		return err
	}
	return nil
}
