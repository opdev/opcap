package packages

import (
	"context"
	"fmt"
	"strings"

	pkgserverv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func List(ctx context.Context, client client.Client, catalogSource string, filter []string) ([]pkgserverv1.PackageManifest, error) {
	var packageManifestList pkgserverv1.PackageManifestList
	if err := client.List(ctx, &packageManifestList); err != nil {
		return nil, err
	}

	pkgs := filterPackageManifests(packageManifestList.Items, catalogSource, filter)

	if err := checkFilteredResults(pkgs, filter); err != nil {
		return []pkgserverv1.PackageManifest{}, nil
	}

	return pkgs, nil
}

func filterPackageManifests(manifests []pkgserverv1.PackageManifest, catalogSource string, filter []string) []pkgserverv1.PackageManifest {
	result := []pkgserverv1.PackageManifest{}

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
		return fmt.Errorf("could not find the following requested package filters:\n%#v", joinedMissingPackages)
	}
	return nil
}
