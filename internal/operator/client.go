package operator

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	configv1 "github.com/openshift/api/config/v1"
	operatorv1 "github.com/operator-framework/api/pkg/operators/v1"
	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	olmclient "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned"
	pkgserverv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

type Client interface {
	CreateNamespace(ctx context.Context, name string) (*corev1.Namespace, error)
	DeleteNamespace(ctx context.Context, name string) error
	CreateOperatorGroup(ctx context.Context, data OperatorGroupData, namespace string) (*operatorv1.OperatorGroup, error)
	DeleteOperatorGroup(ctx context.Context, name string, namespace string) error
	CreateSubscription(ctx context.Context, data SubscriptionData, namespace string) (*operatorv1alpha1.Subscription, error)
	DeleteSubscription(ctx context.Context, name string, namespace string) error
	GetCompletedCsvWithTimeout(ctx context.Context, namespace string, delay time.Duration) (*operatorv1alpha1.ClusterServiceVersion, error)
	GetOpenShiftVersion(ctx context.Context) (string, error)
	ListPackageManifests(ctx context.Context, list *pkgserverv1.PackageManifestList, catalogSource string, filter []string) error
	GetSubscriptionData(ctx context.Context, source string, namespace string, filter []string) ([]SubscriptionData, error)
	ListCRDs(ctx context.Context, list *apiextensionsv1.CustomResourceDefinitionList) error
	CreateUnstructured(ctx context.Context, namespace string, obj *unstructured.Unstructured, gvr schema.GroupVersionResource) (*unstructured.Unstructured, error)
	GetUnstructured(ctx context.Context, namespace, name string, gvr schema.GroupVersionResource) (*unstructured.Unstructured, error)
	DeleteUnstructured(ctx context.Context, namespace, name string, gvr schema.GroupVersionResource) error
	ListClusterServiceVersions(ctx context.Context, namespace string) (*operatorv1alpha1.ClusterServiceVersionList, error)
}

type operatorClient struct {
	Client        runtimeClient.Client
	OlmClient     olmclient.Interface
	DynamicClient dynamic.Interface
}

func addSchemes(scheme *runtime.Scheme) error {
	if err := operatorv1.AddToScheme(scheme); err != nil {
		return err
	}

	if err := operatorv1alpha1.AddToScheme(scheme); err != nil {
		return err
	}

	if err := pkgserverv1.AddToScheme(scheme); err != nil {
		return err
	}

	if err := apiextensionsv1.AddToScheme(scheme); err != nil {
		return err
	}

	if err := corev1.AddToScheme(scheme); err != nil {
		return err
	}

	if err := configv1.AddToScheme(scheme); err != nil {
		return err
	}

	return nil
}

func NewOpCapClient(kubeconfig *rest.Config) (Client, error) {
	scheme := runtime.NewScheme()

	if err := addSchemes(scheme); err != nil {
		return nil, fmt.Errorf("could not add schemes to client: %v", err)
	}

	client, err := runtimeClient.New(kubeconfig, runtimeClient.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("could not get subscription client: %v", err)
	}

	olmClient, err := newOlmClientset(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("could not create OLM clientset: %v", err)
	}

	dynamicClient, err := newDynamicClient(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("could not create dynamic client: %v", err)
	}

	var operatorClient Client = &operatorClient{
		Client:        client,
		OlmClient:     olmClient,
		DynamicClient: dynamicClient,
	}
	return operatorClient, nil
}

func newOlmClientset(kubeconfig *rest.Config) (*olmclient.Clientset, error) {
	return olmclient.NewForConfig(kubeconfig)
}

// NewDynamicClient creates a new dynamic client or returns an error.
func newDynamicClient(kubeconfig *rest.Config) (dynamic.Interface, error) {
	return dynamic.NewForConfig(kubeconfig)
}
