package cmd

import (
	"bytes"
	"context"

	. "github.com/onsi/ginkgo/v2/dsl/core"
	. "github.com/onsi/gomega"
	"github.com/opdev/opcap/internal/operator"
	configv1 "github.com/openshift/api/config/v1"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	pkgserverv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"
	"github.com/spf13/afero"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

var _ = Describe("Check command tests", func() {
	When("creating a checkCmd", func() {
		It("should contain flags", func() {
			cmd := checkCmd()
			Expect(cmd.HasFlags()).To(BeTrue())
		})
	})

	When("executing the checkCmd", func() {
		It("should error because of no KUBECONFIG", func() {
			_, err := executeCommand(checkCmd(), "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("please provide kubeconfig"))
		})
	})

	When("running audits", func() {
		It("should succeed", func() {
			checkflags.AuditPlan = []string{"fakeplan"}
			checkflags.CatalogSource = "test-cs"
			fakekubeconfig := &rest.Config{}
			pkg := pkgserverv1.PackageManifest{
				TypeMeta: metav1.TypeMeta{
					Kind:       "PackageManifest",
					APIVersion: "packages.operators.coreos.com/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-package",
				},
				Status: pkgserverv1.PackageManifestStatus{
					CatalogSource: "test-cs",
					Channels: []pkgserverv1.PackageChannel{
						{
							Name: "test",
							CurrentCSVDesc: pkgserverv1.CSVDescription{
								InstallModes: []v1alpha1.InstallMode{
									{
										Type:      v1alpha1.InstallModeTypeOwnNamespace,
										Supported: true,
									},
								},
							},
						},
					},
					DefaultChannel: "test",
				},
			}
			version := configv1.ClusterVersion{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "config.openshift.io/v1",
					Kind:       "ClusterVersion",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "version",
				},
				Status: configv1.ClusterVersionStatus{
					History: []configv1.UpdateHistory{
						{Version: "4.10.34"},
						{Version: "4.9.8"},
					},
				},
			}
			output := bytes.NewBufferString("")
			Expect(runAudits(context.TODO(), fakekubeconfig, operator.NewFakeOpClient(&pkg, &version), afero.NewMemMapFs(), output)).To(Succeed())
			// This is empty since no audits should actually run here.
			Expect(output.String()).To(BeEmpty())
		})
	})
})
