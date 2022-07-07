package operator

import (
	"context"
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	log "opcap/internal/logger"

	operatorv1 "github.com/operator-framework/api/pkg/operators/v1"
	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"

	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"

	olmclient "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned"
	pkgsclientv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/client/clientset/versioned"

	"k8s.io/client-go/tools/clientcmd"
)

var logger = log.Sugar

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
	WaitForCsvOnNamespace(namespace string) (string, error)
	GetOpenShiftVersion() (string, error)
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
		logger.Errorf("could not get kubeconfig")
		return nil, err
	}

	client, err := runtimeClient.New(kubeconfig, runtimeClient.Options{Scheme: scheme})
	if err != nil {
		logger.Errorf("could not get subscription client")
		return nil, err
	}

	var operatorClient Client = &operatorClient{
		Client: client,
	}
	return operatorClient, nil
}

func NewPackageServerClient() (*pkgsclientv1.Clientset, error) {
	cfg, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	if err != nil {
		logger.Errorf("Unable to build config from flags: %w", err)
	}
	pkgsclient, err := pkgsclientv1.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	return pkgsclient, nil
}

func NewOlmClientset() (*olmclient.Clientset, error) {
	cfg, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	if err != nil {
		return nil, err
	}

	olmClientset, err := olmclient.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	return olmClientset, nil
}
