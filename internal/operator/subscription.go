package operator

import (
	"context"

	"strings"

	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// operator client
// NewSubscrition
// client.create(NewSubscription)

type subscriptionClient struct {
	client runtimeclient.Client
}

type SubscriptionData struct {
	Name                   string
	Channel                string
	CatalogSource          string
	CatalogSourceNamespace string
	Package                string
}

// SubscriptionList represent the set of operators
// to be installed and tested
// It's a unique list of package/channels for operator install
func subscriptions(catalogSource string, catalogSourceNamespace string) []SubscriptionData {

	SubscriptionList := []SubscriptionData{}

	for _, b := range bundleList() {
		s := SubscriptionData{
			Name:                   strings.Join([]string{b.PackageName, b.ChannelName, "subscription"}, "-"),
			Channel:                b.ChannelName,
			CatalogSource:          catalogSource,
			CatalogSourceNamespace: catalogSourceNamespace,
			Package:                b.PackageName,
		}
		SubscriptionList = append(SubscriptionList, s)
	}
	return uniqueElementsOf(SubscriptionList)
}

func uniqueElementsOf(s []SubscriptionData) []SubscriptionData {
	unique := make(map[SubscriptionData]bool, len(s))
	uniqueSubscriptionData := make([]SubscriptionData, len(unique))
	for _, elem := range s {
		if !unique[elem] {
			uniqueSubscriptionData = append(uniqueSubscriptionData, elem)
			unique[elem] = true
		}
	}
	return uniqueSubscriptionData
}

func (c subscriptionClient) CreateSubscription(ctx context.Context, data SubscriptionData) (*operatorv1alpha1.Subscription, error) {
	subscription := &operatorv1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name: data.Name,
		},
		Spec: &operatorv1alpha1.SubscriptionSpec{
			CatalogSource:          data.CatalogSource,
			CatalogSourceNamespace: data.CatalogSourceNamespace,
			Channel:                data.Channel,
			Package:                data.Package,
		},
	}
	err := c.client.Create(ctx, subscription)
	return subscription, err
}

func (c subscriptionClient) DeleteSubscription(ctx context.Context, name string) error {
	subscription := &operatorv1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	return c.client.Delete(ctx, subscription)
}

func (c subscriptionClient) GetSubscription(ctx context.Context, name string) (*operatorv1alpha1.Subscription, error) {
	subscription := &operatorv1alpha1.Subscription{}
	err := c.client.Get(ctx, runtimeclient.ObjectKey{
		Name: name,
	}, subscription)

	return subscription, err
}

func SubscriptionClient(namespace string) (*subscriptionClient, error) {
	scheme := runtime.NewScheme()
	operatorv1alpha1.AddToScheme(scheme)
	kubeconfig, err := ctrl.GetConfig()
	if err != nil {
		logger.Error("could not get kubeconfig")
		return nil, err
	}
	client, err := runtimeclient.New(kubeconfig, runtimeclient.Options{Scheme: scheme})
	if err != nil {
		logger.Error("could not get subscription client")
		return nil, err
	}

	return &subscriptionClient{
		client: runtimeclient.NewNamespacedClient(client, namespace),
	}, nil
}
