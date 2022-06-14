package capability

import (
	"context"
	log "opcap/internal/logger"
	"opcap/internal/operator"
	"strings"

	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
)

var logger = log.Sugar

// TODO: InstallOperatorsTest creates all subscriptions for a catalogSource sequencially
// We will need other arguments that can tweak how many to test at a time
// And possibly indicate a specific condition

func OperatorInstallAllFromCatalog(catalogSource string, catalogSourceNamespace string) error {

	s, err := operator.Subscriptions(catalogSource, catalogSourceNamespace)
	if err != nil {
		logger.Errorf("Error while getting bundles from CatalogSource %s: %w", catalogSource, err)
		return err
	}

	c, err := operator.NewClient()
	if err != nil {
		logger.Errorf("Error while creating PackageServerClient: %w", err)
		return err
	}

	for _, subscription := range s {

		// TODO: implement this with goroutines for concurrent testing
		// TODO: transform subscriptions list in a queuing mechanism
		// for the test work. Run all individual tests under the umbrella
		// of it's operator dedicated goroutine
		err := OperatorInstall(subscription, c)
		if err != nil {
			logger.Errorw("installing operator", "package", subscription.Package, "channel", subscription.Channel, "installmode", subscription.InstallModeType)
		}

	}

	return nil
}

func OperatorInstall(s operator.SubscriptionData, c operator.Client) error {

	logger.Debugw("installing package", "package", s.Package, "channel", s.Channel, "installmode", s.InstallModeType)

	namespace := strings.Join([]string{"opcap", strings.ReplaceAll(s.Package, ".", "-")}, "-")
	targetNs1 := strings.Join([]string{namespace, "targetNs1"}, "-")
	targetNs2 := strings.Join([]string{namespace, "targetNs2"}, "-")
	operatorGroup := strings.Join([]string{s.Name, s.Channel, "group"}, "-")

	// create operator namespace
	operator.CreateNamespace(context.Background(), namespace)

	// Checking install modes and
	// creating operatorGroup per operator package/channel
	installedNS := []string{namespace}
	switch s.InstallModeType {

	case operatorv1alpha1.InstallModeTypeAllNamespaces:
		opGroupData := operator.OperatorGroupData{
			Name:             operatorGroup,
			TargetNamespaces: []string{},
		}
		c.CreateOperatorGroup(context.Background(), opGroupData, namespace)

	case operatorv1alpha1.InstallModeTypeSingleNamespace:

		operator.CreateNamespace(context.Background(), targetNs1)
		opGroupData := operator.OperatorGroupData{
			Name:             operatorGroup,
			TargetNamespaces: []string{targetNs1},
		}
		installedNS = append(installedNS, targetNs1)
		c.CreateOperatorGroup(context.Background(), opGroupData, namespace)

	case operatorv1alpha1.InstallModeTypeOwnNamespace:
		opGroupData := operator.OperatorGroupData{
			Name:             operatorGroup,
			TargetNamespaces: []string{namespace},
		}
		c.CreateOperatorGroup(context.Background(), opGroupData, namespace)

	case operatorv1alpha1.InstallModeTypeMultiNamespace:

		operator.CreateNamespace(context.Background(), targetNs1)
		operator.CreateNamespace(context.Background(), targetNs2)
		opGroupData := operator.OperatorGroupData{
			Name:             operatorGroup,
			TargetNamespaces: []string{targetNs1, targetNs2},
		}
		installedNS = append(installedNS, targetNs1)
		installedNS = append(installedNS, targetNs2)
		c.CreateOperatorGroup(context.Background(), opGroupData, namespace)

	}

	// create subscription per operator package/channel
	sub, err := c.CreateSubscription(context.Background(), s, namespace)
	if err != nil {
		logger.Debugf("Error creating subscriptions: %w", err)
		return err
	}

	if err = c.WaitForInstallPlan(context.Background(), sub); err != nil {
		logger.Debugf("Waiting for InstallPlan: %w", err)
		return err
	}
	// check/approve install plan
	// TODO: check the name standard for installPlan
	err = c.InstallPlanApprove(namespace)
	if err != nil {
		logger.Debugf("Error creating subscriptions: %w", err)
		return err
	}

	_, err = c.CSVSuceededOnNamespace(namespace)

	if err != nil {
		logger.Infow("failed", "package", s.Package, "channel", s.Channel, "installmode", s.InstallModeType)
	} else {
		logger.Infow("succeeded", "package", s.Package, "channel", s.Channel, "installmode", s.InstallModeType)
	}

	// generate and send report

	// delete subscription
	err = c.DeleteSubscription(context.Background(), s.Name, namespace)
	if err != nil {
		logger.Debugf("Error while deleting Subscription: %w", err)
		return err
	}

	// delete operator group
	err = c.DeleteOperatorGroup(context.Background(), operatorGroup, namespace)
	if err != nil {
		logger.Debugf("Error while deleting OperatorGroup: %w", err)
		return err
	}

	// delete namespaces
	for _, ns := range installedNS {
		operator.DeleteNamespace(context.Background(), ns)
	}
	return nil
}
