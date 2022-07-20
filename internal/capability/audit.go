package capability

import (
	"opcap/internal/operator"
)

// Audit defines all the methods used to run a full audit Plan against a single operator
// All new capability tests should be added to this interface and used with a capAudit
// instance and as part of an auditPlan
type Audit interface {
	OperatorInstall() error
	OperatorCleanUp() error
}

// capAudit is an implementation of the Audit interface
type capAudit struct {

	// client has access to all operator methods
	client operator.Client

	// namespace is the ns where the operator will be installed
	namespace string

	// subscription holds the data to install an operator via OLM
	subscription operator.SubscriptionData

	// auditPlan is a list of functions to be run in sequence in a given audit
	// all of them must be an implemented method of capAudit and must be part
	// of the Audit interface
	auditPlan []string
}

// Temporary fake install for testing
// will remove before merging this PR
// func (ca *capAudit) OperatorCleanUp() error {

// 	fmt.Printf("Cleaning up package %s\n", ca.subscription.Package)
// 	return nil
// }
