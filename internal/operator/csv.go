package operator

import (
	"context"
	"fmt"

	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

func (c operatorClient) GetCSVPhase(namespace string) (operatorv1alpha1.ClusterServiceVersionPhase, error) {

	clusterServiceVersionList := operatorv1alpha1.ClusterServiceVersionList{}

	listOpts := runtimeClient.ListOptions{
		Namespace: namespace,
	}

	err := c.Client.List(context.Background(), &clusterServiceVersionList, &listOpts)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	// TODO: create a custom error for this
	if len(clusterServiceVersionList.Items) > 1 {
		return "", fmt.Errorf("more than one CSV found in dedicated namespace %s", fmt.Sprint(len(clusterServiceVersionList.Items)))
	}

	if len(clusterServiceVersionList.Items) == 0 {
		return "", fmt.Errorf("no CSV found in namespace %s", fmt.Sprint(len(clusterServiceVersionList.Items)))
	}

	clusterServiceVersion := operatorv1alpha1.ClusterServiceVersion{}

	err = c.Client.Get(context.Background(), types.NamespacedName{Name: clusterServiceVersionList.Items[0].ObjectMeta.Name, Namespace: namespace}, &clusterServiceVersion)

	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return clusterServiceVersion.Status.Phase, nil
}
