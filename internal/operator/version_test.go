package operator

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	configv1 "github.com/openshift/api/config/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Version", func() {
	var operatorClient operatorClient
	Context("GetVersion", func() {
		BeforeEach(func() {
			scheme := runtime.NewScheme()
			configv1.AddToScheme(scheme)
			// test data
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
			client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&version).Build()
			operatorClient.Client = client
		})
		When("getting openshift version", func() {
			It("should succeed", func() {
				version, err := operatorClient.GetOpenShiftVersion(context.TODO())
				Expect(err).ToNot(HaveOccurred())
				Expect(version).To(Equal("4.10.34"))
			})
		})
	})
	Context("VersionNotFound", func() {
		BeforeEach(func() {
			scheme := runtime.NewScheme()
			configv1.AddToScheme(scheme)
			client := fake.NewClientBuilder().WithScheme(scheme).WithObjects().Build()
			operatorClient.Client = client
		})
		When("there is no version obj", func() {
			It("should error", func() {
				_, err := operatorClient.GetOpenShiftVersion(context.TODO())
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
