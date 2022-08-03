package capability

import (
	"context"
	"strings"
	"time"

	log "github.com/opdev/opcap/internal/logger"
	"github.com/opdev/opcap/internal/operator"
	"github.com/opdev/opcap/internal/report"
)

var logger = log.Sugar

func (ca *capAudit) OperatorInstall() error {
	logger.Debugw("installing package", "package", ca.subscription.Package, "channel", ca.subscription.Channel, "installmode", ca.subscription.InstallModeType)

	// create operator's own namespace
	operator.CreateNamespace(context.Background(), ca.namespace)

	// create remaining target namespaces watched by the operator
	for _, ns := range ca.operatorGroupData.TargetNamespaces {
		if ns != ca.namespace {
			operator.CreateNamespace(context.Background(), ns)
		}
	}

	// create operator group for operator package/channel
	ca.client.CreateOperatorGroup(context.Background(), ca.operatorGroupData, ca.namespace)

	// create subscription for operator package/channel
	_, err := ca.client.CreateSubscription(context.Background(), ca.subscription, ca.namespace)
	if err != nil {
		logger.Debugf("Error creating subscriptions: %w", err)
		return err
	}

	// Get a Succeeded or Failed CSV with one minute timeout
	csv, err := ca.client.GetCompletedCsvWithTimeout(ca.namespace, 1*time.Minute)

	if err != nil {

		// If error is timeout than don't log phase but timeout
		if err.Error() == "operator install timeout" {
			logger.Infow("timeout", "package", ca.subscription.Package, "channel", ca.subscription.Channel, "installmode", ca.subscription.InstallModeType)
			return nil
		} else {
			return err
		}
	}

	// if operator completed log Succeeded or Failed according to status field
	logger.Infow(strings.ToLower(string(csv.Status.Phase)), "package", ca.subscription.Package, "channel", ca.subscription.Channel, "installmode", ca.subscription.InstallModeType)

	ocpVersion, err := ca.client.GetOpenShiftVersion()
	if err != nil {
		return err
	}

	// Initializing new operator report
	// TODO: consolidate data in capAudit object and pass the whole object
	r := report.NewOperatorInstallReport().Init(ocpVersion,
		ca.subscription.Package, ca.subscription.Channel, ca.subscription.CatalogSource,
		string(ca.subscription.InstallModeType), csv.Status, report.OpInstallRptOptPrint{},
		report.OpInstallRptOptFile{FilePath: "operator_install_report.json"})

	err = r.Report()
	if err != nil {
		return err
	}

	return nil
}
