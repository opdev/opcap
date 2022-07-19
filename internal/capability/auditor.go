package capability

import (
	"opcap/internal/operator"
	"reflect"
	"strings"
)

// Auditor interface represents the object running capability audits against operators
// It has methods to create a workqueue with all the package and audit requirements for a
// particular audit run
type Auditor interface {
	BuildWorkQueueByCatalog(catalogSource string, catalogSourceNamespace string, auditPlan []string) error
	RunAudits() error
}

// capAuditor implements Auditor
type capAuditor struct {

	// Workqueue holds capAudits in a buffered channel in order to execute them
	WorkQueue chan capAudit
}

// BuildAuditorByCatalog creates a new Auditor with workqueue based on a selected catalog
func BuildAuditorByCatalog(catalogSource string, catalogSourceNamespace string, auditPlan []string) (capAuditor, error) {

	var auditor capAuditor
	err := auditor.BuildWorkQueueByCatalog(catalogSource, catalogSourceNamespace, auditPlan)
	if err != nil {
		logger.Fatalf("Unable to build workqueue err := %s", err.Error())
	}
	return auditor, nil
}

// BuildWorkQueueByCatalog fills in the auditor workqueue with all package information found in a specific catalog
func (capAuditor *capAuditor) BuildWorkQueueByCatalog(catalogSource string, catalogSourceNamespace string, auditPlan []string) error {

	c, err := operator.NewOpCapClient()
	if err != nil {
		// if it doesn't load the client nothing can be done
		// log and panic
		logger.Panic("Error while creating OpCapClient: %w", err)
	}

	// Getting subscription data form the package manifests available in the selected catalog
	s, err := c.GetSubscriptionData(catalogSource, catalogSourceNamespace)
	if err != nil {
		logger.Errorf("Error while getting bundles from CatalogSource %s: %w", catalogSource, err)
		return err
	}

	// build workqueue as buffered channel based subscriptionData list size
	capAuditor.WorkQueue = make(chan capAudit, len(s))
	defer close(capAuditor.WorkQueue)

	// add capAudits to the workqueue
	for _, subscription := range s {

		var ca capAudit
		ca.namespace = strings.Join([]string{"opcap", strings.ReplaceAll(subscription.Package, ".", "-")}, "-")
		ca.subscription = subscription
		ca.auditPlan = auditPlan

		// load workqueue with capAudit
		capAuditor.WorkQueue <- ca

	}

	return nil
}

// RunAudits executes all selected functions in order for a given audit at a time
func (capAuditor *capAuditor) RunAudits() error {

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
