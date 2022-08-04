package capability

import (
	"context"
	"strings"
	"time"

	log "github.com/opdev/opcap/internal/logger"
	"github.com/opdev/opcap/internal/operator"
)

var logger = log.Sugar

func (ca *CapAudit) OperatorInstall() error {
	logger.Debugw("installing package", "package", ca.Subscription.Package, "channel", ca.Subscription.Channel, "installmode", ca.Subscription.InstallModeType)

	// create operator's own namespace
	operator.CreateNamespace(context.Background(), ca.Namespace)

	// create remaining target namespaces watched by the operator
	for _, ns := range ca.OperatorGroupData.TargetNamespaces {
		if ns != ca.Namespace {
			operator.CreateNamespace(context.Background(), ns)
		}
	}

	// create operator group for operator package/channel
	ca.Client.CreateOperatorGroup(context.Background(), ca.OperatorGroupData, ca.Namespace)

	// create subscription for operator package/channel
	_, err := ca.Client.CreateSubscription(context.Background(), ca.Subscription, ca.Namespace)
	if err != nil {
		logger.Debugf("Error creating subscriptions: %w", err)
		return err
	}

	// Get a Succeeded or Failed CSV with one minute timeout
	ca.Csv, err = ca.Client.GetCompletedCsvWithTimeout(ca.Namespace, 1*time.Minute)

	if err != nil {

		// If error is timeout than don't log phase but timeout
		if err.Error() == "operator install timeout" {
			logger.Infow("timeout", "package", ca.Subscription.Package, "channel", ca.Subscription.Channel, "installmode", ca.Subscription.InstallModeType)
			return nil
		} else {
			return err
		}
	}

	// if operator completed log Succeeded or Failed according to status field
	logger.Infow(strings.ToLower(string(ca.Csv.Status.Phase)), "package", ca.Subscription.Package, "channel", ca.Subscription.Channel, "installmode", ca.Subscription.InstallModeType)

	ca.Report(RptOptionFile{FilePath: "operator_install_report.json"}, RptOptionPrint{})

	return nil
}
