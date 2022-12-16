package capability

import (
	"context"
	"fmt"

	"github.com/opdev/opcap/internal/logger"
)

func operatorCleanup(ctx context.Context, opts ...auditOption) auditCleanupFn {
	var options auditOptions
	for _, opt := range opts {
		err := opt(&options)
		if err != nil {
			return func(_ context.Context) error {
				return fmt.Errorf("option failed: %v", err)
			}
		}
	}

	return func(ctx context.Context) error {
		// delete subscription
		if err := options.client.DeleteSubscription(ctx, options.subscription.Name, options.namespace); err != nil {
			logger.Debugf("Error while deleting Subscription: %w", err)
			return err
		}

		// get csv using csvWatcher
		csv, err := options.client.GetCompletedCsvWithTimeout(ctx, options.namespace, options.csvWaitTime)
		if err != nil {
			return err
		}

		// delete cluster service version
		if err := options.client.DeleteCSV(ctx, csv.ObjectMeta.Name, options.namespace); err != nil {
			logger.Debugf("Error while deleting ClusterServiceVersion: %w", err)
			return err
		}

		// delete operator group
		if err := options.client.DeleteOperatorGroup(ctx, options.operatorGroupData.Name, options.namespace); err != nil {
			logger.Debugf("Error while deleting OperatorGroup: %w", err)
			return err
		}

		// delete target namespaces
		for _, ns := range options.operatorGroupData.TargetNamespaces {
			if err := options.client.DeleteNamespace(ctx, ns); err != nil {
				logger.Debugf("Error deleting target namespace %s", ns)
				return err
			}
		}

		// delete operator's own namespace
		if err := options.client.DeleteNamespace(ctx, options.namespace); err != nil {
			logger.Debugf("Error deleting operator's own namespace %s", options.namespace)
			return err
		}
		return nil
	}
}
