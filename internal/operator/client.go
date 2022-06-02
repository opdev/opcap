package operator

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	operatorv1 "github.com/operator-framework/api/pkg/operators/v1"
	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	log "github.com/sirupsen/logrus"
	ctrl "sigs.k8s.io/controller-runtime"

	"context"

	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

type Client interface {
	CreateOperatorGroup(ctx context.Context, data OperatorGroupData, namespace string) (*operatorv1.OperatorGroup, error)
	DeleteOperatorGroup(ctx context.Context, name string, namespace string) error
	CreateSecret(ctx context.Context, name string, content map[string]string, secretType corev1.SecretType, namespace string) (*corev1.Secret, error)
	DeleteSecret(ctx context.Context, name string, namespace string) error
	CreateSubscription(ctx context.Context, data SubscriptionData, namespace string) (*operatorv1alpha1.Subscription, error)
	DeleteSubscription(ctx context.Context, name string, namespace string) error
	GetSubscription(ctx context.Context, name string, namespace string) (*operatorv1alpha1.Subscription, error)
	InstallPlanApprove(namespace string) error
	WaitForInstallPlan(ctx context.Context, sub *operatorv1alpha1.Subscription) error
	CSVSuceededOnNamespace(namespace string) (*operatorv1alpha1.ClusterServiceVersion, error)
}

type operatorClient struct {
	Client runtimeClient.Client
}

func NewClient() (Client, error) {

	scheme := runtime.NewScheme()

	if err := operatorv1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	if err := operatorv1alpha1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	kubeconfig, err := ctrl.GetConfig()

	if err != nil {
		log.Error("could not get kubeconfig")
		return nil, err
	}

	client, err := runtimeClient.New(kubeconfig, runtimeClient.Options{Scheme: scheme})
	if err != nil {
		log.Error("could not get subscription client")
		return nil, err
	}

	var operatorClient Client = &operatorClient{
		Client: client,
	}
	return operatorClient, nil
}
