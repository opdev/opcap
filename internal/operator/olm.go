package operator

import (
	"context"

	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c operatorClient) ListClusterServiceVersions(ctx context.Context, namespace string) (*operatorv1alpha1.ClusterServiceVersionList, error) {
	return c.OlmClient.OperatorsV1alpha1().ClusterServiceVersions(namespace).List(ctx, v1.ListOptions{})
}
