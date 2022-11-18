package capability

import (
	"bytes"
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/opdev/opcap/internal/operator"
	configv1 "github.com/openshift/api/config/v1"
	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	pkgserverv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"
	"github.com/spf13/afero"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Auditor tests", func() {
	var options auditorOptions
	var fs afero.Fs
	var client operator.Client
	BeforeEach(func() {
		fs = afero.NewMemMapFs()
		packageManifest := &pkgserverv1.PackageManifest{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "packages.operators.coreos.com",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "testnamespace",
			},
			Status: pkgserverv1.PackageManifestStatus{
				CatalogSource:          "testsource",
				CatalogSourceNamespace: "testnamespace",
				Channels: []pkgserverv1.PackageChannel{
					{
						Name: "default",
						CurrentCSVDesc: pkgserverv1.CSVDescription{
							InstallModes: []operatorv1alpha1.InstallMode{
								{
									Type:      operatorv1alpha1.InstallModeTypeOwnNamespace,
									Supported: true,
								},
								{
									Type:      operatorv1alpha1.InstallModeTypeAllNamespaces,
									Supported: true,
								},
							},
						},
					},
				},
				DefaultChannel: "default",
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

		client = operator.NewFakeOpClient(packageManifest, &version)
		Expect(WithAuditPlan([]string{"operatorinstall"})(&options)).To(Succeed())
		Expect(WithCatalogSource("testsource")(&options)).To(Succeed())
		Expect(WithCatalogSourceNamespace("testnamespace")(&options)).To(Succeed())
		Expect(WithPackages([]string{})(&options)).To(Succeed())
		Expect(WithAllInstallModes(false)(&options)).To(Succeed())
		Expect(WithClient(client)(&options)).To(Succeed())
		Expect(WithExtraCRDirectory("")(&options)).To(Succeed())
		Expect(WithFilesystem(fs)(&options)).To(Succeed())
	})

	Context("Auditor functional options", func() {
		var options *auditorOptions

		BeforeEach(func() {
			options = &auditorOptions{}
		})

		Context("audit plan", func() {
			When("plan is supplied", func() {
				It("should set audit plan correctly", func() {
					Expect(WithAuditPlan([]string{"testplan"})(options)).To(Succeed())
					Expect(options.auditPlan).To(ContainElement("testplan"))
				})
			})
			When("no plan is supplied", func() {
				It("should throw an error", func() {
					Expect(WithAuditPlan([]string{})(options)).ToNot(Succeed())
				})
			})
			When("an empty plan is supplied", func() {
				It("should throw an error", func() {
					Expect(WithAuditPlan(([]string{""}))(options)).ToNot(Succeed())
				})
			})
		})
		Context("catalogsource", func() {
			When("catalogsource is supplied", func() {
				It("should set catalogsource correctly", func() {
					Expect(WithCatalogSource("testcs")(options)).To(Succeed())
					Expect(options.catalogSource).To(Equal("testcs"))
				})
			})
		})

		Context("catalogsource namespace", func() {
			When("catalogsource namespace is supplied", func() {
				It("should set catalogsource correctly", func() {
					Expect(WithCatalogSourceNamespace("testns")(options)).To(Succeed())
					Expect(options.catalogSourceNamespace).To(Equal("testns"))
				})
			})
		})

		Context("packages", func() {
			When("packages is supplied", func() {
				It("should set packages correctly", func() {
					Expect(WithPackages([]string{"testpackage"})(options)).To(Succeed())
					Expect(options.packages).To(ContainElement("testpackage"))
				})
			})
		})

		Context("AllInstallModes", func() {
			When("AllInstallModes is supplied", func() {
				It("should set allInstallModes correctly", func() {
					Expect(WithAllInstallModes(true)(options)).To(Succeed())
					Expect(options.allInstallModes).To(BeTrue())
				})
			})
		})

		Context("Client", func() {
			When("client is supplied", func() {
				It("should set client correctly", func() {
					client := operator.NewFakeOpClient()
					Expect(WithClient(client)(options)).To(Succeed())
					Expect(options.opCapClient).To(Equal(client))
				})
			})
			When("client is nil", func() {
				It("should throw an error", func() {
					Expect(WithClient(nil)(options)).ToNot(Succeed())
				})
			})
		})

		Context("Extra CR Directory", func() {
			When("extra CR directory is supplied", func() {
				It("should set extra CR directory correctly", func() {
					Expect(WithExtraCRDirectory("/customcrdir")(options)).To(Succeed())
					Expect(options.extraCustomResources).To(Equal("/customcrdir"))
				})
			})
		})

		Context("Filesystem", func() {
			When("a filesystem is not supplied", func() {
				It("should throw an error", func() {
					Expect(WithFilesystem(nil)(options)).ToNot(Succeed())
				})
			})
		})

		Context("Timeout", func() {
			When("a timeout is supplied", func() {
				It("should set the timeout properly", func() {
					Expect(WithTimeout(time.Second)(options)).To(Succeed())
					Expect(options.timeout).To(Equal(time.Second))
				})
			})
		})
	})

	Context("Extra CR Directory", func() {
		When("extra CR directory is not provided", func() {
			It("should still succeed", func() {
				Expect(RunAudits(context.Background(),
					WithAuditPlan([]string{"operatorinstall"}),
					WithCatalogSource("testsource"),
					WithCatalogSourceNamespace("testnamespace"),
					WithPackages([]string{}),
					WithAllInstallModes(false),
					WithClient(client),
					WithExtraCRDirectory(""),
					WithFilesystem(fs),
					WithTimeout(time.Millisecond),
				)).To(Succeed())
			})
		})
		When("extra CR is provided", func() {
			When("it is empty", func() {
				It("should still succeed", func() {
					Expect(RunAudits(context.Background(),
						WithAuditPlan([]string{"operatorinstall"}),
						WithCatalogSource("testsource"),
						WithCatalogSourceNamespace("testnamespace"),
						WithPackages([]string{}),
						WithAllInstallModes(false),
						WithClient(client),
						WithExtraCRDirectory("/"),
						WithFilesystem(fs),
						WithTimeout(time.Millisecond),
					)).To(Succeed())
				})
			})
			When("It is not empty", func() {
				It("should not throw an error", func() {
					Expect(fs.MkdirAll("/packages/mypackage", 0o755)).To(Succeed())
					Expect(afero.WriteFile(fs, "/packages/mypackage/manifest.json", bytes.NewBufferString(`{"foo":"bar"}`).Bytes(), 0o644))
					options := options
					options.extraCustomResources = "/packages"
					extraCRs, err := extraCRDirectory(context.TODO(), &options)
					Expect(err).ToNot(HaveOccurred())
					Expect(extraCRs).ToNot(BeNil())
					Expect(extraCRs["mypackage"][0]).To(ContainElement("bar"))
				})
			})
		})
	})
})
