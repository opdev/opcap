package operator

import (
	"context"
	"fmt"

	"github.com/operator-framework/operator-registry/pkg/api"
	registryClient "github.com/operator-framework/operator-registry/pkg/client"
)

// list Bundles

func BundleList() {

	c, err := registryClient.NewClient("localhost:50051")
	if err != nil {
		fmt.Println(err)
	}
	bundleIterator, err := c.ListBundles(context.Background())
	if err != nil {
		fmt.Println(err)
	}

	bundles := []api.Bundle{}

	for {
		b := bundleIterator.Next()
		if b == nil {
			break
		}

		bundles = append(bundles, *b)
	}
	for _, bundle := range bundles {
		fmt.Println("-----------------------")
		fmt.Println(bundle.BundlePath)
		fmt.Println(bundle.ChannelName)
		fmt.Println("-----------------------")
	}

}

// list Packages
