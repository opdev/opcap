package operator

import (
	"testing"

	pkgserverv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var manifests = []pkgserverv1.PackageManifest{
	{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mongodb-atlas-kubernetes",
			Namespace: "openshift-marketplace",
		},
		Spec: pkgserverv1.PackageManifestSpec{},
		Status: pkgserverv1.PackageManifestStatus{
			PackageName:   "mongodb-atlas-kubernetes",
			CatalogSource: "certified-operators",
		},
	},
	{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mongodb-atlas-kubernetes",
			Namespace: "openshift-marketplace",
		},
		Spec: pkgserverv1.PackageManifestSpec{},
		Status: pkgserverv1.PackageManifestStatus{
			PackageName:   "mongodb-atlas-kubernetes",
			CatalogSource: "community-operators",
		},
	},
	{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "argocd-operator",
			Namespace: "openshift-marketplace",
		},
		Spec: pkgserverv1.PackageManifestSpec{},
		Status: pkgserverv1.PackageManifestStatus{
			PackageName:   "argocd-operator",
			CatalogSource: "certified-operators",
		},
	},
	{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pachyderm-operator",
			Namespace: "openshift-marketplace",
		},
		Spec: pkgserverv1.PackageManifestSpec{},
		Status: pkgserverv1.PackageManifestStatus{
			PackageName:   "pachyderm-operator",
			CatalogSource: "certified-operators",
		},
	},
}

func TestFilterPackageManifests(t *testing.T) {
	var cases = []struct {
		catalog string
		filters []string
		expect  int
	}{
		{
			catalog: "",
			filters: []string{},
			expect:  4,
		},
		{
			catalog: "certified-operators",
			filters: []string{},
			expect:  3,
		},
		{
			catalog: "",
			filters: []string{"mongodb-atlas-kubernetes"},
			expect:  2,
		},
		{
			catalog: "certified-operators",
			filters: []string{"mongodb-atlas-kubernetes"},
			expect:  1,
		},
	}

	for i, testcase := range cases {
		t.Logf("Test %d => catalog sources: \"%s\", filters: %+v", i, testcase.catalog, testcase.filters)
		results := filterPackageManifests(manifests, testcase.catalog, testcase.filters)
		if len(results) != testcase.expect {
			t.Errorf("expected %d, received %d results", testcase.expect, len(results))
		}
		t.Logf("matched %d out of %d package manifests", len(results), len(manifests))
	}
}

func TestCheckFilteredResults(t *testing.T) {
	err := checkFilteredResults(manifests, []string{"nonexistent-operator", "mongodb-atlas-kubernetes"})
	if err == nil {
		t.Error("expecting 1 missing package. None found.")
	}
	t.Logf("packages not found in catalog sources: %v\n", err)
}
