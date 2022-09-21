package operator

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/opdev/opcap/internal/logger"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Namespace", func() {
	logger.InitLogger("debug")
	scheme := runtime.NewScheme()
	corev1.AddToScheme(scheme)
	client := fake.NewClientBuilder().WithScheme(scheme).Build()
	var operatorClient operatorClient = operatorClient{
		Client: client,
	}

	Context("CreateNamespace", func() {
		When("creating a namespace", func() {
			ns, err := operatorClient.CreateNamespace(context.TODO(), "testns")
			It("should succeed", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(ns).ToNot(BeNil())
			})
		})
		When("creating a namespace that already exists", func() {
			JustBeforeEach(func() {
				ns, err := operatorClient.CreateNamespace(context.TODO(), "testns")
				It("should error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("could not create namespace: testns: namespaces \"testns\" already exists"))
					Expect(ns).To(BeNil())
				})
			})
		})
	})
	Context("DeleteNamespace", func() {
		When("deleting the existing namespace", func() {
			err := operatorClient.DeleteNamespace(context.TODO(), "testns")
			It("should succeed", func() {
				Expect(err).ToNot(HaveOccurred())
			})
		})
		When("deleting a namespace that does not exist", func() {
			JustBeforeEach(func() {
				err := operatorClient.DeleteNamespace(context.TODO(), "testns")
				It("should return error", func() {
					Expect(err).To(MatchError("could not delete namespace: testns: namespaces \"testns\" not found"))
				})
			})
		})
	})
})
