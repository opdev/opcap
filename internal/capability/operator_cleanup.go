package capability

import (
	"context"

	"github.com/opdev/opcap/internal/logger"
)

func (ca *capAudit) OperatorCleanUp(ctx context.Context) error {
	// delete subscription
	if err := ca.client.DeleteSubscription(ctx, ca.subscription.Name, ca.namespace); err != nil {
		logger.Debugf("Error while deleting Subscription: %w", err)
		return err
	}

	// delete operator group
	if err := ca.client.DeleteOperatorGroup(ctx, ca.operatorGroupData.Name, ca.namespace); err != nil {
		logger.Debugf("Error while deleting OperatorGroup: %w", err)
		return err
	}

	// delete target namespaces
	for _, ns := range ca.operatorGroupData.TargetNamespaces {
		if err := ca.client.DeleteNamespace(ctx, ns); err != nil {
			logger.Debugf("Error deleting target namespace %s", ns)
			return err
		}
	}

	// delete operator's own namespace
	if err := ca.client.DeleteNamespace(ctx, ca.namespace); err != nil {
		logger.Debugf("Error deleting operator's own namespace %s", ca.namespace)
		return err
	}
	return nil
}
