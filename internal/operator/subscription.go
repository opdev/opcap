package operator

import (
	"context"
	"strings"

	pkgserverv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"

	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SubscriptionData struct {
	Name                   string
	Channel                string
	CatalogSource          string
	CatalogSourceNamespace string
	Package                string
	InstallModeType        operatorv1alpha1.InstallModeType
	InstallPlanApproval    operatorv1alpha1.Approval
}

// SubscriptionList represent the set of operators
// to be installed and tested
// It's a unique list of package/channels for operator install
func (c operatorClient) GetSubscriptionData(catalogSource string, catalogSourceNamespace string) ([]SubscriptionData, error) {
	var packageManifests pkgserverv1.PackageManifestList
	err := c.ListPackageManifests(context.Background(), &packageManifests)
	if err != nil {
		logger.Errorf("Error while listing new PackageManifest Objects: %w", err)
		return nil, err
	}

	SubscriptionList := []SubscriptionData{}

	for _, pkgm := range packageManifests.Items {
		if pkgm.Status.CatalogSource == catalogSource {
			for _, pkgch := range pkgm.Status.Channels {
				if pkgch.IsDefaultChannel(pkgm) {
					for _, installMode := range pkgch.CurrentCSVDesc.InstallModes {
						if installMode.Supported {
							s := SubscriptionData{
								Name:                   strings.Join([]string{pkgch.Name, pkgm.Name, "subscription"}, "-"),
								Channel:                pkgch.Name,
								CatalogSource:          catalogSource,
								CatalogSourceNamespace: catalogSourceNamespace,
								Package:                pkgm.Name,
								InstallModeType:        installMode.Type,
								InstallPlanApproval:    operatorv1alpha1.ApprovalAutomatic,
							}

							SubscriptionList = append(SubscriptionList, s)
							break
						}
					}
				}
			}
		}
	}

	return SubscriptionList, nil
}

func (c operatorClient) CreateSubscription(ctx context.Context, data SubscriptionData, namespace string) (*operatorv1alpha1.Subscription, error) {
	subscription := &operatorv1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name:      data.Name,
			Namespace: namespace,
		},
		Spec: &operatorv1alpha1.SubscriptionSpec{
			CatalogSource:          data.CatalogSource,
			CatalogSourceNamespace: data.CatalogSourceNamespace,
			Channel:                data.Channel,
			InstallPlanApproval:    data.InstallPlanApproval,
			Package:                data.Package,
		},
	}
	err := c.Client.Create(ctx, subscription)
	return subscription, err
}

func (c operatorClient) DeleteSubscription(ctx context.Context, name string, namespace string) error {
	subscription := &operatorv1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	return c.Client.Delete(ctx, subscription)
}

func (c operatorClient) ListPackageManifests(ctx context.Context, list *pkgserverv1.PackageManifestList) error {
	return c.Client.List(ctx, list)
}
