package operator

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/runtime"

	log "github.com/opdev/opcap/internal/logger"
	configv1 "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	operatorv1 "github.com/operator-framework/api/pkg/operators/v1"
	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	olmclient "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned"
	pkgserverv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var logger = log.Sugar

type Client interface {
	CreateOperatorGroup(ctx context.Context, data OperatorGroupData, namespace string) (*operatorv1.OperatorGroup, error)
	DeleteOperatorGroup(ctx context.Context, name string, namespace string) error
	CreateSubscription(ctx context.Context, data SubscriptionData, namespace string) (*operatorv1alpha1.Subscription, error)
	DeleteSubscription(ctx context.Context, name string, namespace string) error
	GetCompletedCsvWithTimeout(namespace string, delay time.Duration) (operatorv1alpha1.ClusterServiceVersion, error)
	GetOpenShiftVersion() (string, error)
	ListPackageManifests(ctx context.Context, list *pkgserverv1.PackageManifestList, filter []string) error
	GetSubscriptionData(source string, namespace string, filter []string) ([]SubscriptionData, error)
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

	kubeconfig, err := kubeConfig()
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

// kubeConfig return kubernetes cluster config
func kubeConfig() (*rest.Config, error) {
	return ctrl.GetConfig()
}

// mustGetConfig returns only the kubernetes cluster config.
// nil is returned when an error is encountered
func mustGetConfig() *rest.Config {
	config, err := kubeConfig()
	if err != nil {
		return nil
	}
	return config
}

func NewOlmClientset() (*olmclient.Clientset, error) {
	cfg, err := kubeConfig()
	if err != nil {
		return nil, err
	}

	return olmclient.NewForConfig(cfg)
}

// NewDynamicClient creates a new dynamic client or returns an error.
func NewDynamicClient() (dynamic.Interface, error) {
	config, err := kubeConfig()
	if err != nil {
		return nil, err
	}
	return dynamic.NewForConfig(config)
}

// NewKubernetesClient returns a kubernetes clientset
func NewKubernetesClient() (*kubernetes.Clientset, error) {
	config, err := kubeConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func NewConfigClient() (*configv1.ConfigV1Client, error) {
	// create openshift config clientset
	cfg, err := kubeConfig()
	if err != nil {
		return nil, err
	}

	return configv1.NewForConfig(cfg)
}
