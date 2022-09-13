package capability

import (
	"context"
	"fmt"
)

func (ca *capAudit) OperatorCleanUp(ctx context.Context) error {
	// delete subscription
	if err := ca.client.DeleteSubscription(ctx, ca.subscription.Name, ca.namespace); err != nil {
		return fmt.Errorf("could not delete subscription: %s: %v", ca.subscription.Name, err)
	}

	// delete operator group
	if err := ca.client.DeleteOperatorGroup(ctx, ca.operatorGroupData.Name, ca.namespace); err != nil {
		return fmt.Errorf("could not delete OperatorGroup: %s: %v", ca.operatorGroupData.Name, err)
	}

	// delete target namespaces
	for _, ns := range ca.operatorGroupData.TargetNamespaces {
		if err := ca.client.DeleteNamespace(ctx, ns); err != nil {
			return fmt.Errorf("could not delete target namespace: %s: %v", ns, err)
		}
	}

	// delete operator's own namespace
	if err := ca.client.DeleteNamespace(ctx, ca.namespace); err != nil {
		return fmt.Errorf("could not delete opeator's own namespace: %s: %v", ca.namespace, err)
	}
	return nil
}
