package capability

import (
	"context"
	"reflect"

	"github.com/opdev/opcap/internal/logger"
	"github.com/opdev/opcap/internal/operator"
)

// capAuditor implements Auditor
type CapAuditor struct {
	// AuditPlan holds the tests that should be run during an audit
	AuditPlan []string

	// CatalogSource may be built-in OLM or custom
	CatalogSource string
	// CatalogSourceNamespace will be openshift-marketplace or custom
	CatalogSourceNamespace string

	// Packages is a subset of packages to be tested from a catalogSource
	Packages []string

	// WorkQueue holds capAudits in a buffered channel in order to execute them
	WorkQueue chan capAudit

	// AllInstallModes will test all install modes supported by an operator
	AllInstallModes bool
}

// BuildWorkQueueByCatalog fills in the auditor workqueue with all package information found in a specific catalog
func (capAuditor *CapAuditor) buildWorkQueueByCatalog(ctx context.Context) error {
	c, err := operator.NewOpCapClient()
	if err != nil {
		// if it doesn't load the client nothing can be done
		// log and panic
		logger.Panic("Error while creating OpCapClient: %w", err)
	}

	// Getting subscription data form the package manifests available in the selected catalog
	subscriptions, err := c.GetSubscriptionData(ctx, capAuditor.CatalogSource, capAuditor.CatalogSourceNamespace, capAuditor.Packages)
	if err != nil {
		logger.Errorf("Error while getting bundles from CatalogSource %s: %w", capAuditor.CatalogSource, err)
		return err
	}

	// build workqueue as buffered channel based subscriptionData list size
	capAuditor.WorkQueue = make(chan capAudit, len(subscriptions))
	defer close(capAuditor.WorkQueue)

	// packagesToBeAudited is a subset of packages to be tested from a catalogSource
	var packagesToBeAudited []operator.SubscriptionData

	// get all install modes for all operators in the catalog
	// and add them to the packagesToBeAudited list
	if capAuditor.AllInstallModes {
		packagesToBeAudited = subscriptions
	} else {
		packages := make(map[string]bool)
		for _, subscription := range subscriptions {
			if _, exists := packages[subscription.Package]; !exists {
				packages[subscription.Package] = true
				packagesToBeAudited = append(packagesToBeAudited, subscription)
			}
		}
	}

	// add capAudits to the workqueue
	for _, subscription := range packagesToBeAudited {
		capAudit, err := newCapAudit(ctx, c, subscription, capAuditor.AuditPlan)
		if err != nil {
			logger.Debugf("Couldn't build capAudit for subscription %s", "Err:", err)
			return err
		}

		// load workqueue with capAudit
		capAuditor.WorkQueue <- capAudit
	}

	return nil
}

// RunAudits executes all selected functions in order for a given audit at a time
func (capAuditor *CapAuditor) RunAudits(ctx context.Context) error {
	err := capAuditor.buildWorkQueueByCatalog(ctx)
	if err != nil {
		logger.Debugf("Unable to build workqueue err := %s", err.Error())
		return err
	}

	// read workqueue for audits
	for audit := range capAuditor.WorkQueue {
		// read a particular audit's auditPlan for functions
		// to be executed against operator
		for _, function := range audit.auditPlan {
			// run function/method by name
			// NOTE: The signature for this method MUST be:
			// func Fn(context.Context) error
			m := reflect.ValueOf(&audit).MethodByName(function)
			in := []reflect.Value{reflect.ValueOf(ctx)}
			m.Call(in)
		}
	}
	return nil
}
