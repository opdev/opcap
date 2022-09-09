package capability

import (
	"context"
	"os"

	"github.com/opdev/opcap/internal/logger"
	"github.com/opdev/opcap/internal/operator"
)

func (ca *capAudit) OperatorInstall(ctx context.Context) error {
	logger.Debugw("installing package", "package", ca.subscription.Package, "channel", ca.subscription.Channel, "installmode", ca.subscription.InstallModeType)
	// create operator's own namespace
	if _, err := operator.CreateNamespace(ctx, ca.namespace); err != nil {
		return err
	}

	// create remaining target namespaces watched by the operator
	for _, ns := range ca.operatorGroupData.TargetNamespaces {
		if ns != ca.namespace {
			operator.CreateNamespace(ctx, ns)
		}
	}

	// create operator group for operator package/channel
	ca.client.CreateOperatorGroup(ctx, ca.operatorGroupData, ca.namespace)

	// create subscription for operator package/channel
	if _, err := ca.client.CreateSubscription(ctx, ca.subscription, ca.namespace); err != nil {
		logger.Debugf("Error creating subscriptions: %w", err)
		return err
	}

	// Get a Succeeded or Failed CSV with one minute timeout
	var err error
	ca.csv, err = ca.client.GetCompletedCsvWithTimeout(ctx, ca.namespace, ca.csvWaitTime)

	if err != nil {
		// If error is timeout than don't log phase but timeout
		if err.Error() == "operator install timeout" {
			ca.csvTimeout = true
		} else {
			return err
		}
	}

	file, err := os.OpenFile("operator_install_report.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		file.Close()
		return err
	}
	defer file.Close()

	if err := ca.OperatorInstallJsonReport(file); err != nil {
		logger.Debugf("Error during the OperatorInstall Report: %w", err)
		return err
	}

	if err := ca.OperatorInstallTextReport(os.Stdout); err != nil {
		logger.Debugf("Error during the OperatorInstall Text Report: %w", err)
		return err
	}

	return nil
}
