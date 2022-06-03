package operator

import (
	"context"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// NewClient
func GetK8sClient() *kubernetes.Clientset {
	// create k8s client
	cfg, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	if err != nil {
		_ = fmt.Errorf("unable to build config from flags: %v", err)
	}
	clientset, _ := kubernetes.NewForConfig(cfg)

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
		log.Debug(fmt.Errorf("%w: error while creating Namespace: %s", err, name))
		return nil, err
	}
	log.Debugf("Namespace Created: ", name)
	return &nsSpec, nil
}

func DeleteNamespace(ctx context.Context, name string) error {
	operatorClient := GetK8sClient()
	log.Debugf("Deleting namespace: %s", name)
	err := operatorClient.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		log.Debug(fmt.Errorf("%w: error while deleting Namespace: %s", err, name))
		return err
	}
	log.Debugf("Namespace Deleted: ", name)
	return nil
}
