package operator

import (
	"context"
	"os"

	"k8s.io/apimachinery/pkg/runtime"

	log "github.com/opdev/opcap/internal/logger"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	operatorv1 "github.com/operator-framework/api/pkg/operators/v1"
	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"

	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"

	olmclient "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned"
	pkgserverv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

var logger = log.Sugar

type Client interface {
	CreateOperatorGroup(ctx context.Context, data OperatorGroupData, namespace string) (*operatorv1.OperatorGroup, error)
	DeleteOperatorGroup(ctx context.Context, name string, namespace string) error
	CreateSubscription(ctx context.Context, data SubscriptionData, namespace string) (*operatorv1alpha1.Subscription, error)
	DeleteSubscription(ctx context.Context, name string, namespace string) error
	WaitForCsvOnNamespace(namespace string) (string, error)
	GetOpenShiftVersion() (string, error)
	ListPackageManifests(ctx context.Context, list *pkgserverv1.PackageManifestList) error
	GetSubscriptionData(source string, namespace string) ([]SubscriptionData, error)
	ListCRDs(ctx context.Context, list *apiextensionsv1.CustomResourceDefinitionList) error
}

type operatorClient struct {
	Client runtimeClient.Client
}

func NewOpCapClient() (Client, error) {
	scheme := runtime.NewScheme()

	if err := operatorv1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	if err := operatorv1alpha1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	if err := pkgserverv1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	if err := apiextensionsv1.AddToScheme(scheme); err != nil {
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

// NewDynamicClient creates a new dynamic client or returns an error.
func NewDynamicClient() (dynamic.Interface, error) {
	config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	if err != nil {
		return nil, err
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}
