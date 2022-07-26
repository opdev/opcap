package capability

import (
	"context"
	"strings"
	"time"

	"github.com/opdev/opcap/internal/operator"

	log "github.com/opdev/opcap/internal/logger"
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

	return nil
}
