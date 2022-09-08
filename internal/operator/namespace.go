package operator

import (
	"context"

	"github.com/opdev/opcap/internal/logger"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

// CreateNamespace
func (o *operatorClient) CreateNamespace(ctx context.Context, name string) (*corev1.Namespace, error) {
	logger.Debugf("Create namespace: %s", name)
	nsSpec := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	if err := o.Client.Create(ctx, &nsSpec, &runtimeClient.CreateOptions{}); err != nil {
		logger.Errorf("error while creating Namespace %s: %s", name, err.Error())
		return nil, err
	}
	logger.Debugf("Namespace Created: %s", name)
	return &nsSpec, nil
}

// DeleteNamespace
func (o *operatorClient) DeleteNamespace(ctx context.Context, name string) error {
	logger.Debugf("Delete namespace: %s", name)
	nsSpec := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	if err := o.Client.Delete(ctx, &nsSpec, &runtimeClient.DeleteOptions{}); err != nil {
		logger.Errorf("error while deleting Namespace %s: %s", name, err.Error())
		return err
	}
	logger.Debugf("Namespace Deleted: %s", name)
	return nil
}
