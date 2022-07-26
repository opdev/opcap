package capability

import (
	"context"
	"strings"

	log "opcap/internal/logger"
	"opcap/internal/operator"

	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
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
	sub, err := ca.client.CreateSubscription(context.Background(), ca.subscription, ca.namespace)
	if err != nil {
		logger.Debugf("Error creating subscriptions: %w", err)
		return err
	}

	// wait for installPlan and approve it if it's in manual mode
	if sub.Spec.InstallPlanApproval == operatorv1alpha1.ApprovalManual {
		if err = ca.client.ApproveInstallPlan(context.Background(), sub); err != nil {
			logger.Debugf("Waiting for InstallPlan: %w", err)
			return err
		}
	}

	csvStatus, err := ca.client.WaitForCsvOnNamespace(ca.namespace)

	if err != nil {
		logger.Infow("failed", "package", ca.subscription.Package, "channel", ca.subscription.Channel, "installmode", ca.subscription.InstallModeType)
	} else {
		logger.Infow(strings.ToLower(csvStatus), "package", ca.subscription.Package, "channel", ca.subscription.Channel, "installmode", ca.subscription.InstallModeType)
	}

	return nil
}
