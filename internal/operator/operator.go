package operator

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
)

// TODO: InstallOperatorsTest creates all subscriptions for a catalogSource sequencially
// We will need other arguments that can tweak how many to test at a time
// And possibly indicate a specific condition

func InstallOperatorsTest(catalogSource string, catalogSourceNamespace string) error {

	s := subscriptions(catalogSource, catalogSourceNamespace)

	c, err := SubscriptionClient("test")
	if err != nil {
		log.Fatal(err)
	}

	for _, subscription := range s {

		// TODO: implement this with goroutines for concurrent testing
		// TODO: transform subscriptions list in a queuing mechanism
		// for the test work. Run all individual tests under the umbrella
		// of it's operator dedicated goroutine
		installOperator(subscription, c)

	}

	return nil
}

func installOperator(s SubscriptionData, c *subscriptionClient) error {

	// create operatorGroup per operator package/channel

	// create subscription per operator package/channel
	_, err := c.CreateSubscription(context.Background(), s)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Test subscription for %s created successfully\n", s.Name)

	// check/approve install plan

	// check CSV/operator status

	// generate and send report

	// delete subscription

	// delete operator group

	// delete namespace ?

	return nil
}

// Installer:
// 1. openshift-install create cluster --install-config myfile.yaml
// 2. wait for install to complete
//
// ---- for each operator on queue ----
//
// Bundle:
// 3. run bundle (create operator group, subscription, approve install plan)
// 4. wait for operator to be ready - check CSV to be ready
//
// CR:
// 5. create CR
// 6. wait for CR to be ready
//
// CAPABILITY:
// 7. run tests (split multiple times in each of the 5 levels)
// 8. retrieve data
// 9. repeat until finish
//
// REPORT:
// 10. generate and publish report
//
// CLEAN UP OPERATOR:
// 11. clean up operator and operand
//
// ---- go next until queue is complete ---
//
// DELETE CLUSTER:
// 12. clean up cluster and exit
