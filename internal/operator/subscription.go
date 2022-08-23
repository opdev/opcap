package operator

import (
	"context"
	"fmt"
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
func (c operatorClient) GetSubscriptionData(options OperatorCheckOptions) ([]SubscriptionData, error) {
	var packageManifests pkgserverv1.PackageManifestList
	err := c.ListPackageManifests(context.Background(), &packageManifests, options)
	if err != nil {
		logger.Errorf("Error while listing new PackageManifest Objects: %w", err)
		return nil, err
	}

	SubscriptionList := []SubscriptionData{}

	for _, pkgm := range packageManifests.Items {
		// iterate through the channels and set value
		// for defaultChannel variable
		var defaultChannel pkgserverv1.PackageChannel
		for _, pkgch := range pkgm.Status.Channels {
			if pkgch.IsDefaultChannel(pkgm) {
				defaultChannel = pkgch
			}
		}

		// create a new subscription for each of the supported install mode types
		// on the default channel
		for _, installMode := range defaultChannel.CurrentCSVDesc.InstallModes {
			if installMode.Supported {
				SubscriptionList = append(SubscriptionList,
					SubscriptionData{
						Name:                   strings.Join([]string{defaultChannel.Name, pkgm.Name, strings.ToLower(string(installMode.Type)), "subscription"}, "-"),
						Channel:                defaultChannel.Name,
						CatalogSource:          options.CatalogSource,
						CatalogSourceNamespace: options.CatalogSourceNamespace,
						Package:                pkgm.Name,
						InstallModeType:        installMode.Type,
						InstallPlanApproval:    operatorv1alpha1.ApprovalAutomatic,
					},
				)
			}

			// Return subscriptions for all install modes
			// only when AllInstallModes is true
			if !options.AllInstallModes {
				break
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

func (c operatorClient) ListPackageManifests(ctx context.Context, list *pkgserverv1.PackageManifestList, options OperatorCheckOptions) error {
	var pkgManifestsList pkgserverv1.PackageManifestList
	if err := c.Client.List(ctx, &pkgManifestsList); err != nil {
		return err
	}

	pkgs := filterPackageManifests(pkgManifestsList.Items, options.CatalogSource, options.FilterPackages)

	if err := checkFilteredResults(pkgs, options.FilterPackages); err != nil {
		return err
	}
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

func checkFilteredResults(pkgs []pkgserverv1.PackageManifest, filter []string) error {
	if len(filter) > 0 && len(pkgs) != len(filter) {
		var missingPackages []string
		for _, f := range filter {
			notFound := true
			for _, pkg := range pkgs {
				if f == pkg.Name {
					notFound = false
				}
			}
			if notFound {
				missingPackages = append(missingPackages, f)
			}
		}
		joinedMissingPackages := strings.Join(missingPackages, ", ")
		return fmt.Errorf("Could not find the following requested package filters:\n%#v", joinedMissingPackages)
	}
	return nil
}

// ListCRDs returns the list of CRDs present in the cluster
func (c operatorClient) ListCRDs(ctx context.Context, list *apiextensionsv1.CustomResourceDefinitionList) error {
	return c.Client.List(ctx, list)
}
