package capability

import (
	"context"
	"fmt"

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
		// return the error
		return fmt.Errorf("could not create OpCapClient: %v", err)
	}

	// Getting subscription data form the package manifests available in the selected catalog
	subscriptions, err := c.GetSubscriptionData(ctx, capAuditor.CatalogSource, capAuditor.CatalogSourceNamespace, capAuditor.Packages)
	if err != nil {
		return fmt.Errorf("could not get bundles from CatalogSource: %s: %v", capAuditor.CatalogSource, err)
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
			return fmt.Errorf("could not build configuration for subscription: %s: %v", subscription.Name, err)
		}

		// load workqueue with capAudit
		capAuditor.WorkQueue <- *capAudit
	}

	return nil
}

// RunAudits executes all selected functions in order for a given audit at a time
func (capAuditor *CapAuditor) RunAudits(ctx context.Context) error {
	err := capAuditor.buildWorkQueueByCatalog(ctx)
	if err != nil {
		return fmt.Errorf("unable to build workqueue: %v", err)
	}

	// read workqueue for audits
	for audit := range capAuditor.WorkQueue {
		// read a particular audit's auditPlan for functions
		// to be executed against operator
		for _, function := range audit.auditPlan {
			// run function/method by name
			// NOTE: The signature for this method MUST be:
			// func Fn(context.Context) error
			auditFn := newAudit(ctx, function,
				withClient(audit.client),
				withNamespace(audit.namespace),
				withOperatorGroupData(&audit.operatorGroupData),
				withSubscription(&audit.subscription),
				withTimeout(int(audit.csvWaitTime)),
			)
			if auditFn == nil {
				logger.Errorf("invalid audit plan specified: %s", function)
				continue
			}
			err := auditFn(ctx)
			if err != nil {
				logger.Errorf("error in audit: %v", err)
			}
		}
	}
	return nil
}
