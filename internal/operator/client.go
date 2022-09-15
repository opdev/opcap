package operator

import (
	"context"
	"errors"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/opdev/opcap/internal/logger"
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
	"k8s.io/client-go/tools/clientcmd"
)

type Client interface {
	CreateOperatorGroup(ctx context.Context, data OperatorGroupData, namespace string) (*operatorv1.OperatorGroup, error)
	DeleteOperatorGroup(ctx context.Context, name string, namespace string) error
	CreateSubscription(ctx context.Context, data SubscriptionData, namespace string) (*operatorv1alpha1.Subscription, error)
	DeleteSubscription(ctx context.Context, name string, namespace string) error
	GetCompletedCsvWithTimeout(ctx context.Context, namespace string, delay time.Duration) (operatorv1alpha1.ClusterServiceVersion, error)
	GetOpenShiftVersion(ctx context.Context) (string, error)
	ListPackageManifests(ctx context.Context, list *pkgserverv1.PackageManifestList, catalogSource string, filter []string) error
	GetSubscriptionData(ctx context.Context, source string, namespace string, filter []string) ([]SubscriptionData, error)
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
		return nil, fmt.Errorf("could not get kubeconfig: %v", err)
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
	config, err := ctrl.GetConfig()
	if err != nil {
		// returned when there is no kubeconfig
		if errors.Is(err, clientcmd.ErrEmptyConfig) {
			return nil, fmt.Errorf("please provide kubeconfig before retrying: %v", err)
		}

		// returned when the kubeconfig has no servers
		if errors.Is(err, clientcmd.ErrEmptyCluster) {
			return nil, fmt.Errorf("malformed kubeconfig. Please check before retrying: %v", err)
		}

		// any other errors getting kubeconfig would be caught here
		return nil, fmt.Errorf("error getting kubeocnfig. Please check before retrying: %v", err)
	}
	return config, nil
}

func NewOlmClientset() (*olmclient.Clientset, error) {
	kubeconfig, err := kubeConfig()
	if err != nil {
		return nil, fmt.Errorf("could not get kubeconfig: %v", err)
	}

	return olmclient.NewForConfig(kubeconfig)
}

// NewDynamicClient creates a new dynamic client or returns an error.
func NewDynamicClient() (dynamic.Interface, error) {
	kubeconfig, err := kubeConfig()
	if err != nil {
		return nil, fmt.Errorf("could not get kubeconfig: %v", err)
	}

	return dynamic.NewForConfig(kubeconfig)
}

// NewKubernetesClient returns a kubernetes clientset
func NewKubernetesClient() (*kubernetes.Clientset, error) {
	kubeconfig, err := kubeConfig()
	if err != nil {
		return nil, fmt.Errorf("could not get kubeconfig: %v", err)
	}

	return kubernetes.NewForConfig(kubeconfig)
}

func NewConfigClient() (*configv1.ConfigV1Client, error) {
	// create openshift config clientset
	kubeconfig, err := kubeConfig()
	if err != nil {
		return nil, fmt.Errorf("could not get kubeconfig: %v", err)
	}
	return configv1.NewForConfig(kubeconfig)
}
