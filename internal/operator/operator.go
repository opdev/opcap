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

// GetPackageManifests

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
		log.Error(fmt.Errorf("%w: error while creating Namespace: %s", err, name))
		return nil, err
	}
	log.Info("Namespace Created: ", name)
	return &nsSpec, nil
}

func DeleteNamespace(ctx context.Context, name string) error {
	operatorClient := GetK8sClient()
	log.Debugf("Deleting namespace: %s", name)
	err := operatorClient.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		log.Error(fmt.Errorf("%w: error while creating Namespace: %s", err, name))
		return err
	}
	log.Info("Namespace Deleted: ", name)
	return nil
}

// CreateOperatorGroup

// CreateSubscription

// ApproveInstallPlan

// GetCSVStatus

// OperatorCleanUp

// // DeleteSubscription

// // DeleteOperatorGroup

// // DeleteNamespace

// Installer:
// 1. openshift-install create cluster --install-config myfile.yaml
// 2. wait for install to complete
//
// ---- for each operator on queue ----
//
// Bundle:
// 3. run bundle (create operator group, subscription, approve install plan)
// 4. wait for operator to be ready - check CSV to be ready
//
// CR:
// 5. create CR
// 6. wait for CR to be ready
//
// CAPABILITY:
// 7. run tests (split multiple times in each of the 5 levels)
// 8. retrieve data
// 9. repeat until finish
//
// REPORT:
// 10. generate and publish report
//
// CLEAN UP OPERATOR:
// 11. clean up operator and operand
//
// ---- go next until queue is complete ---
//
// DELETE CLUSTER:
// 12. clean up cluster and exit
