package operator

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

// NewClient
func GetK8sClient() *kubernetes.Clientset {
	// create k8s client
	kubeconfig, err := ctrl.GetConfig()
	if err != nil {
		logger.Errorf("unable to build config from flags: %w", err)
	}
	clientset, _ := kubernetes.NewForConfig(kubeconfig)

	return clientset
}

// CreateNamespace
func CreateNamespace(ctx context.Context, name string) (*corev1.Namespace, error) {
	operatorClient := GetK8sClient()
	nsSpec := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	_, err := operatorClient.CoreV1().Namespaces().Create(ctx, &nsSpec, metav1.CreateOptions{})
	if err != nil {
		logger.Errorf("error while creating Namespace %s: %w", name, err)
		return nil, err
	}
	logger.Debugf("Namespace Created: %s", name)
	return &nsSpec, nil
}

func DeleteNamespace(ctx context.Context, name string) error {
	operatorClient := GetK8sClient()
	logger.Debugf("Delete namespace: %s", name)
	err := operatorClient.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		logger.Errorf("error while deleting Namespace %s: %w", name, err)
		return err
	}
	logger.Debugf("Namespace Deleted: %s", name)
	return nil
}
