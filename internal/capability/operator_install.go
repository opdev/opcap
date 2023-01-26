package capability

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/opdev/opcap/internal/logger"
	"github.com/opdev/opcap/internal/operator"
	"github.com/opdev/opcap/internal/report"
)

func operatorInstall(ctx context.Context, opts ...auditOption) (auditFn, auditCleanupFn) {
	var options auditOptions
	for _, opt := range opts {
		err := opt(&options)
		if err != nil {
			return func(_ context.Context) error {
					return fmt.Errorf("option failed: %v", err)
				},
				func(_ context.Context) error {
					return nil
				}
		}
	}

	return func(ctx context.Context) error {
		logger.Debugw("installing package", "package", options.subscription.Package, "channel", options.subscription.Channel, "installmode", options.subscription.InstallModeType)

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
		if _, err := options.client.CreateSubscription(ctx, *options.subscription, options.namespace); err != nil {
			logger.Debugf("Error creating subscriptions: %w", err)
			return err
		}

		// Get a Succeeded or Failed CSV with one minute timeout
		resultCSV, err := options.client.GetCompletedCsvWithTimeout(ctx, options.namespace, options.csvWaitTime)
		if err != nil {
			// If error is timeout than don't log phase but timeout
			if errors.Is(err, operator.TimeoutError) {
				options.csvTimeout = true
				options.csv = resultCSV
				// if err = CollectDebugData(ctx, options, "operator_detailed_report_timeout.json"); err != nil {
				// 	return fmt.Errorf("couldn't collect debug data: %s", err)
				// }

			} else {
				return err
			}
		}
		options.csv = resultCSV

		file, err := options.fs.OpenFile("operator_install_report.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		defer file.Close()

		err = report.OperatorInstallJsonReport(file, report.TemplateData{
			OcpVersion:   options.ocpVersion,
			Subscription: *options.subscription,
			Csv:          options.csv,
			CsvTimeout:   options.csvTimeout,
		})
		if err != nil {
			return fmt.Errorf("could not generate operator install JSON report: %v", err)
		}

		err = report.OperatorInstallTextReport(options.reportWriter, report.TemplateData{
			OcpVersion:   options.ocpVersion,
			Subscription: *options.subscription,
			Csv:          options.csv,
			CsvTimeout:   options.csvTimeout,
		})
		if err != nil {
			return fmt.Errorf("could not generate operator install text report: %v", err)
		}
		if options.detailedReports {
			if err = CollectDebugData(ctx, options, "operator_detailed_report_all.json"); err != nil {
				return fmt.Errorf("couldn't collect debug data: %s", err)
			}
		}

		return nil
	}, operatorCleanup(ctx, opts...)
}
