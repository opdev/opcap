package capability

import (
	"context"
	"fmt"

	"github.com/opdev/opcap/internal/logger"
	"github.com/opdev/opcap/internal/operator"
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
		}

		// get csv using csvWatcher
		csv, err := options.client.GetCompletedCsvWithTimeout(ctx, options.namespace, options.csvWaitTime)
		if err != operator.TimeoutError && err != nil {
			logger.Debugf("Error while deleting CSV: %w", err)
		}

		if csv != nil {
			// delete cluster service version
			if err := options.client.DeleteCSV(ctx, csv.ObjectMeta.Name, options.namespace); err != nil {
				logger.Debugf("Error while deleting ClusterServiceVersion: %w", err)
			}
		}

		// delete operator group
		if err := options.client.DeleteOperatorGroup(ctx, options.operatorGroupData.Name, options.namespace); err != nil {
			logger.Debugf("Error while deleting OperatorGroup: %w", err)
		}

		// delete target namespaces
		for _, ns := range options.operatorGroupData.TargetNamespaces {
			if err := options.client.DeleteNamespace(ctx, ns); err != nil {
				logger.Debugf("Error deleting target namespace %s", ns)
			}
		}

		// delete operator's own namespace
		if err := options.client.DeleteNamespace(ctx, options.namespace); err != nil {
			logger.Debugf("Error deleting operator's own namespace %s", options.namespace)
		}
		return nil
	}
}
