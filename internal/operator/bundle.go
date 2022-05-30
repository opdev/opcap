package operator

import (
	"context"
	"fmt"

	registryApi "github.com/operator-framework/operator-registry/pkg/api"
	registryClient "github.com/operator-framework/operator-registry/pkg/client"
)

// List bundles
// TODO: for now the list method will get data from any catalogsource
// exposed via oc port-forward method on a local machine for testing
// In a future issue we need to address this by discovering the catalogsources
// present in the cluster. We're currently assuming only certified operators
// and will extend that for all the others.
// This is why the newClient for registry is created with localhost:50051
// It requires a manual step of getting the catalog source exposed through
// port forward like below:
// oc port-forward -n openshift-marketplace certified-operators-q7nc8  50051:50051
func bundleList() []registryApi.Bundle {

	c, err := registryClient.NewClient("localhost:50051")
	if err != nil {
		fmt.Println(err)
	}
	bundleIterator, err := c.ListBundles(context.Background())
	if err != nil {
		fmt.Println(err)
	}

	bundles := []registryApi.Bundle{}

	for {
		b := bundleIterator.Next()
		if b == nil {
			break
		}

		bundles = append(bundles, *b)
	}
	return bundles
}
