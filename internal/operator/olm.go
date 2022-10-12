package operator

import (
	"context"

	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

func (c operatorClient) ListClusterServiceVersions(ctx context.Context, namespace string) (*operatorv1alpha1.ClusterServiceVersionList, error) {
	var csvs operatorv1alpha1.ClusterServiceVersionList
	err := c.Client.List(ctx, &csvs, &runtimeClient.ListOptions{Namespace: namespace})
	return &csvs, err
}
