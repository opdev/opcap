package operator

import (
	"context"

	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func NewSubscriptionList() *[]SubscriptionData {

	s := &[]SubscriptionData{{
		Name:                   "subscription-test",
		Channel:                "stable",
		CatalogSource:          "certified-operators",
		CatalogSourceNamespace: "openshift-marketplace",
		Package:                "gpu-operator-certified",
	}}
	return s
}

func (c subscriptionClient) Create(ctx context.Context, data SubscriptionData) (*operatorv1alpha1.Subscription, error) {
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

func (c subscriptionClient) Delete(ctx context.Context, name string) error {
	subscription := &operatorv1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	return c.client.Delete(ctx, subscription)
}

func (c subscriptionClient) Get(ctx context.Context, name string) (*operatorv1alpha1.Subscription, error) {
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
		log.Error("could not get kubeconfig")
		return nil, err
	}
	client, err := client.New(kubeconfig, runtimeclient.Options{Scheme: scheme})
	if err != nil {
		log.Error("could not get subscription client")
		return nil, err
	}

	return &subscriptionClient{
		client: runtimeclient.NewNamespacedClient(client, namespace),
	}, nil
}
