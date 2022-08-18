package operator

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateNamespace
func CreateNamespace(ctx context.Context, name string) (*corev1.Namespace, error) {
	operatorClient, err := NewKubernetesClient()
	if err != nil {
		return nil, err
	}

	nsSpec := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	_, err = operatorClient.CoreV1().Namespaces().Create(ctx, &nsSpec, metav1.CreateOptions{})
	if err != nil {
		logger.Errorf("error while creating Namespace %s: %w", name, err)
		return nil, err
	}
	logger.Debugf("Namespace Created: %s", name)
	return &nsSpec, nil
}

func DeleteNamespace(ctx context.Context, name string) error {
	operatorClient, err := NewKubernetesClient()
	if err != nil {
		return err
	}

	logger.Debugf("Delete namespace: %s", name)
	err = operatorClient.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		logger.Errorf("error while deleting Namespace %s: %w", name, err)
		return err
	}
	logger.Debugf("Namespace Deleted: %s", name)
	return nil
}
