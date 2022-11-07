package cmd

import (
	"bytes"
	"context"
	"os"

	. "github.com/onsi/ginkgo/v2/dsl/core"
	. "github.com/onsi/gomega"
	pkgserverv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("ListPackage Cmd", func() {
	var scheme *runtime.Scheme

	BeforeEach(func() {
		DeferCleanup(os.Setenv, "KUBECONFIG", os.Getenv("KUBECONFIG"))
		os.Unsetenv("KUBECONFIG")
		scheme = runtime.NewScheme()
		Expect(pkgserverv1.AddToScheme(scheme)).To(Succeed())
	})
	When("Executing the command", func() {
		When("no KUBECONFIG is set", func() {
			It("should fail", func() {
				_, err := executeCommand(listPackagesCmd(), []string{"--catalogsource=test-catalogsource"}...)
				Expect(err).To(HaveOccurred())
			})
		})
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
				out := bytes.NewBufferString("")
				err := listPackages(context.TODO(), out, fakeClient)
				Expect(err).ToNot(HaveOccurred())
			})
		})
		When("the list fails", func() {
			It("should error", func() {
				builder := fake.NewClientBuilder()
				fakeClient := builder.WithLists().Build()
				out := bytes.NewBufferString("")
				err := listPackages(context.TODO(), out, fakeClient)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
