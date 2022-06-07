package operator

import (
	"context"
	"errors"
	"time"

	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

func (c operatorClient) CSVSuceededOnNamespace(namespace string) (*operatorv1alpha1.ClusterServiceVersion, error) {

	clusterServiceVersionList := operatorv1alpha1.ClusterServiceVersionList{}

	listOpts := runtimeClient.ListOptions{
		Namespace: namespace,
	}

	deadline := 1 * time.Minute
	ticker := time.NewTicker(2 * time.Second)
	timeout := time.After(deadline)

loop:
	for {

		select {

		case <-timeout:
			break loop

		case <-ticker.C:
			// list CSVs on namespace
			err := c.Client.List(context.Background(), &clusterServiceVersionList, &listOpts)
			if err != nil {
				logger.Errorf("Unable to list CSVs in namespace %s: %s", namespace, err)
				return nil, err
			}

			if len(clusterServiceVersionList.Items) == 0 {

				continue

			} else if clusterServiceVersionList.Items[0].Status.Phase == operatorv1alpha1.CSVPhaseSucceeded {

				return &clusterServiceVersionList.Items[0], nil

			} else {
				continue
			}
		}

	}

	return nil, errors.New("Deadline exceeded for CSV on namespace: " + namespace)
}
