package operator

import (
	"context"
	"strings"

	pkgserverv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"

	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
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
func (c operatorClient) GetSubscriptionData(catalogSource string, catalogSourceNamespace string, filter []string) ([]SubscriptionData, error) {
	var packageManifests pkgserverv1.PackageManifestList
	err := c.ListPackageManifests(context.Background(), &packageManifests, catalogSource, filter)
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

func (c operatorClient) ListPackageManifests(ctx context.Context, list *pkgserverv1.PackageManifestList, catalogSource string, filter []string) error {
	var pkgManifestsList pkgserverv1.PackageManifestList
	if err := c.Client.List(ctx, &pkgManifestsList); err != nil {
		return err
	}

	pkgs := filterPackageManifests(pkgManifestsList.Items, catalogSource, filter)
	list.Items = append(list.Items, pkgs...)

	return nil
}

func filterPackageManifests(manifests []pkgserverv1.PackageManifest, catalogSource string, filter []string) []pkgserverv1.PackageManifest {
	var result = []pkgserverv1.PackageManifest{}

	if len(filter) == 0 && catalogSource == "" {
		return manifests
	}

	matchCatalogSource := func(manifest pkgserverv1.PackageManifest) bool {
		if catalogSource == "" {
			return true
		}
		return manifest.Status.CatalogSource == catalogSource
	}

	matchFilters := func(manifest pkgserverv1.PackageManifest) bool {
		if len(filter) == 0 {
			return true
		}

		for _, f := range filter {
			if f == manifest.Name {
				return true
			}
		}

		return false
	}

	for _, pkg := range manifests {
		if matchCatalogSource(pkg) && matchFilters(pkg) {
			result = append(result, pkg)
		}
	}

	return result
}

// ListCRDs returns the list of CRDs present in the cluster
func (c operatorClient) ListCRDs(ctx context.Context, list *apiextensionsv1.CustomResourceDefinitionList) error {
	return c.Client.List(ctx, list)
}
