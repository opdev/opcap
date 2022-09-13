package capability

import (
	"context"
	"fmt"
	"os"

	"github.com/opdev/opcap/internal/logger"
)

func operatorInstall(ctx context.Context, opts ...auditOption) auditFn {
	var options options
	for _, opt := range opts {
		err := opt(&options)
		if err != nil {
			return func(_ context.Context) error {
				return fmt.Errorf("option failed: %v", err)
			}
		}
	}

	return func(ctx context.Context) error {
		logger.Debugw("installing package", "package", options.Subscription.Package, "channel", options.Subscription.Channel, "installmode", options.Subscription.InstallModeType)

		// create operator's own namespace
		if _, err := options.client.CreateNamespace(ctx, options.namespace); err != nil {
			return err
		}

		// create remaining target namespaces watched by the operator
		for _, ns := range options.operatorGroupData.TargetNamespaces {
			if ns != options.namespace {
				options.client.CreateNamespace(ctx, ns)
			}
		}

		// create operator group for operator package/channel
		options.client.CreateOperatorGroup(ctx, *options.operatorGroupData, options.namespace)

		// create subscription for operator package/channel
		if _, err := options.client.CreateSubscription(ctx, *options.Subscription, options.namespace); err != nil {
			logger.Debugf("Error creating subscriptions: %w", err)
			return err
		}

		// Get a Succeeded or Failed CSV with one minute timeout
		var err error
		options.Csv, err = options.client.GetCompletedCsvWithTimeout(ctx, options.namespace, options.csvWaitTime)

		if err != nil {
			// If error is timeout than don't log phase but timeout
			if err.Error() == "operator install timeout" {
				options.CsvTimeout = true
			} else {
				return err
			}
		}

		file, err := os.OpenFile("operator_install_report.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		defer file.Close()

		// TODO: What to do with reports?
		_ = operatorInstallJsonReport(file, options)

		_ = operatorInstallTextReport(os.Stdout, options)

		return nil
	}
}
