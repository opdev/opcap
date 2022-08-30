package capability

import (
	"reflect"

	"github.com/opdev/opcap/internal/operator"
)

// Auditor interface represents the object running capability audits against operators
// It has methods to create a workqueue with all the package and audit requirements for a
// particular audit run
type Auditor interface {
	RunAudits() error
}

// capAuditor implements Auditor
type CapAuditor struct {

	// AuditPlan holds the tests that should be run during an audit
	AuditPlan []string

	// CatalogSource may be built-in OLM or custom
	CatalogSource string
	// CatalogSourceNamespace will be openshift-marketplace or custom
	CatalogSourceNamespace string

	// FilterPackages is a subset of packages to be tested from a catalogSource
	FilterPackages []string

	// Workqueue holds capAudits in a buffered channel in order to execute them
	WorkQueue chan capAudit
}

// BuildWorkQueueByCatalog fills in the auditor workqueue with all package information found in a specific catalog
func (capAuditor *CapAuditor) buildWorkQueueByCatalog() error {

	c, err := operator.NewOpCapClient()
	if err != nil {
		// if it doesn't load the client nothing can be done
		// log and panic
		logger.Panic("Error while creating OpCapClient: %w", err)
	}

	// Getting subscription data form the package manifests available in the selected catalog
	subscriptions, err := c.GetSubscriptionData(capAuditor.CatalogSource, capAuditor.CatalogSourceNamespace, capAuditor.FilterPackages)
	if err != nil {
		logger.Errorf("Error while getting bundles from CatalogSource %s: %w", capAuditor.CatalogSource, err)
		return err
	}

	// build workqueue as buffered channel based subscriptionData list size
	capAuditor.WorkQueue = make(chan capAudit, len(subscriptions))
	defer close(capAuditor.WorkQueue)

	// add capAudits to the workqueue
	for _, subscription := range subscriptions {

		capAudit, err := newCapAudit(c, subscription, capAuditor.AuditPlan)
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
func (capAuditor *CapAuditor) RunAudits() error {

	err := capAuditor.buildWorkQueueByCatalog()
	if err != nil {
		logger.Fatalf("Unable to build workqueue err := %s", err.Error())
	}

	// read workqueue for audits
	for audit := range capAuditor.WorkQueue {

		// read a particular audit's auditPlan for functions
		// to be executed against operator
		for _, function := range audit.auditPlan {

			// run function/method by name
			m := reflect.ValueOf(&audit).MethodByName(function)
			m.Call(nil)
		}

	}
	return nil
}
