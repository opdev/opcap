package operator

import (
	"context"
	"errors"
	"fmt"
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
	"k8s.io/client-go/tools/clientcmd"
)

var logger = log.Sugar

type Client interface {
	CreateOperatorGroup(ctx context.Context, data OperatorGroupData, namespace string) (*operatorv1.OperatorGroup, error)
	DeleteOperatorGroup(ctx context.Context, name string, namespace string) error
	CreateSubscription(ctx context.Context, data SubscriptionData, namespace string) (*operatorv1alpha1.Subscription, error)
	DeleteSubscription(ctx context.Context, name string, namespace string) error
	GetCompletedCsvWithTimeout(namespace string, delay time.Duration) (operatorv1alpha1.ClusterServiceVersion, error)
	GetOpenShiftVersion() (string, error)
	ListPackageManifests(ctx context.Context, list *pkgserverv1.PackageManifestList, opts OperatorCheckOptions) error
	GetSubscriptionData(options OperatorCheckOptions) ([]SubscriptionData, error)
	ListCRDs(ctx context.Context, list *apiextensionsv1.CustomResourceDefinitionList) error
}

type operatorClient struct {
	Client runtimeClient.Client
}

type OperatorCheckOptions struct {
	// AuditPlan is an ordered list of tests to be run
	// during an operator audit
	AuditPlan []string
	// CatalogSource provides target catalog source
	// from which to list package manifests
	CatalogSource string
	// CatalogSourceNamespace specifies the namespace of the
	// catalog source to be used
	CatalogSourceNamespace string
	// ListPackages operation lists packages in the catalog source
	ListPackages bool
	// FilterPackages provides a list of packages to find
	FilterPackages []string
	// AllInstallModes is passed to test all install modes
	AllInstallModes bool
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

	kubeConfig, err := kubeConfig()
	if err != nil {
		return nil, fmt.Errorf("could not get kubeconfig: %v", err)
	}
	client, err := runtimeClient.New(kubeConfig, runtimeClient.Options{Scheme: scheme})
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
			return nil, fmt.Errorf("please provide kubeconfig before retrying")
		}

		// returned when the kubeconfig has no servers
		if errors.Is(err, clientcmd.ErrEmptyCluster) {
			return nil, fmt.Errorf("malformed kubeconfig. Please check before retrying")
		}

		// any other errors getting kubeconfig would be caught here
		return nil, fmt.Errorf("error getting kubeconfig. Please check before retrying")
	}
	return config, nil
}

func NewOlmClientset() (*olmclient.Clientset, error) {
	kubeConfig, err := kubeConfig()
	if err != nil {
		return nil, err
	}
	return olmclient.NewForConfig(kubeConfig)
}

// NewDynamicClient creates a new dynamic client or returns an error.
func NewDynamicClient() (dynamic.Interface, error) {
	kubeConfig, err := kubeConfig()
	if err != nil {
		return nil, err
	}
	return dynamic.NewForConfig(kubeConfig)
}

// NewKubernetesClient returns a kubernetes clientset
func NewKubernetesClient() (*kubernetes.Clientset, error) {
	kubeConfig, err := kubeConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(kubeConfig)
}

func NewConfigClient() (*configv1.ConfigV1Client, error) {
	// create openshift config clientset
	kubeConfig, err := kubeConfig()
	if err != nil {
		return nil, err
	}
	return configv1.NewForConfig(kubeConfig)
}
