package cmd

import (
	. "github.com/onsi/ginkgo/v2/dsl/core"
	. "github.com/onsi/gomega"
	pkgserverv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("PackageList Cmd", func() {
	var scheme *runtime.Scheme

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		Expect(pkgserverv1.AddToScheme(scheme)).To(Succeed())
	})
	When("Calling the command", func() {
		When("catalogsource is given", func() {
			It("should succeed", func() {
				packageManifest := &pkgserverv1.PackageManifest{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "packages.operators.coreos.com",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test-catalogsourcenamespace",
					},
					Status: pkgserverv1.PackageManifestStatus{
						CatalogSource:          "test-catalogsource",
						CatalogSourceNamespace: "test-catalogsourcenamespace",
					},
				}
				fakeClientBuilder := fake.NewClientBuilder().WithScheme(scheme)
				fakeClient := fakeClientBuilder.WithObjects([]client.Object{packageManifest}...).Build()
				output, err := executeCommand(listPackagesCmd(fakeClient), []string{"--catalogsource=test-catalogsource"}...)
				Expect(err).ToNot(HaveOccurred())
				Expect(output).To(MatchRegexp("test\\s+test-catalogsource\\s+test-catalogsourcename"))
			})
		})
		When("the list fails", func() {
			It("should error", func() {
				By("removing the scheme", func() {
					builder := fake.NewClientBuilder()
					fakeClient := builder.WithLists().Build()
					_, err := executeCommand(listPackagesCmd(fakeClient))
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})
})
