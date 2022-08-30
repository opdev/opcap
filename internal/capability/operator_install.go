package capability

import (
	"context"

	log "github.com/opdev/opcap/internal/logger"
	"github.com/opdev/opcap/internal/operator"
)

var logger = log.Sugar

func (ca *capAudit) OperatorInstall() error {
	logger.Debugw("installing package", "package", ca.subscription.Package, "channel", ca.subscription.Channel, "installmode", ca.subscription.InstallModeType)

	// create operator's own namespace
	if _, err := operator.CreateNamespace(context.Background(), ca.namespace); err != nil {
		return err
	}

	// create remaining target namespaces watched by the operator
	for _, ns := range ca.operatorGroupData.TargetNamespaces {
		if ns != ca.namespace {
			operator.CreateNamespace(context.Background(), ns)
		}
	}

	// create operator group for operator package/channel
	ca.client.CreateOperatorGroup(context.Background(), ca.operatorGroupData, ca.namespace)

	// create subscription for operator package/channel
	if _, err := ca.client.CreateSubscription(context.Background(), ca.subscription, ca.namespace); err != nil {
		logger.Debugf("Error creating subscriptions: %w", err)
		return err
	}

	// Get a Succeeded or Failed CSV with one minute timeout
	var err error
	ca.csv, err = ca.client.GetCompletedCsvWithTimeout(ca.namespace, ca.csvWaitTime)

	if err != nil {
		// If error is timeout than don't log phase but timeout
		if err.Error() == "operator install timeout" {
			ca.csvTimeout = true
		} else {
			return err
		}
	}

	ca.Report(OperatorInstallRptOptionFile{FilePath: "operator_install_report.json"}, OperatorInstallRptOptionPrint{})

	return nil
}
