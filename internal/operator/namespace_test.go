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
	var operatorClient operatorClient

	BeforeEach(func() {
		logger.InitLogger("debug")
		scheme := runtime.NewScheme()
		corev1.AddToScheme(scheme)
		client := fake.NewClientBuilder().WithScheme(scheme).Build()
		operatorClient.Client = client
	})

	Context("CreateNamespace", func() {
		When("creating a namespace", func() {
			It("should succeed", func() {
				ns, err := operatorClient.CreateNamespace(context.TODO(), "testns")
				Expect(err).ToNot(HaveOccurred())
				Expect(ns).ToNot(BeNil())
			})
		})
		When("creating a namespace that already exists", func() {
			JustBeforeEach(func() {
				_, err := operatorClient.CreateNamespace(context.TODO(), "testns")
				Expect(err).ToNot(HaveOccurred())
			})
			It("should error", func() {
				ns, err := operatorClient.CreateNamespace(context.TODO(), "testns")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("could not create namespace: testns: namespaces \"testns\" already exists"))
				Expect(ns).To(BeNil())
			})
		})
	})
	Context("DeleteNamespace", func() {
		When("deleting the existing namespace", func() {
			JustBeforeEach(func() {
				_, err := operatorClient.CreateNamespace(context.TODO(), "testns")
				Expect(err).ToNot(HaveOccurred())
			})
			It("should succeed", func() {
				Expect(operatorClient.DeleteNamespace(context.TODO(), "testns")).To(Succeed())
			})
		})
		When("deleting a namespace that does not exist", func() {
			It("should return error", func() {
				Expect(operatorClient.DeleteNamespace(context.TODO(), "testns")).ToNot(Succeed())
			})
		})
	})
})
