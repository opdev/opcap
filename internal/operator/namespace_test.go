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
	Context("Namespaces", func() {
		It("should exercise Namespaces", func() {
			By("creating a Namespace", func() {
				ns, err := operatorClient.CreateNamespace(context.TODO(), "testns")
				Expect(err).ToNot(HaveOccurred())
				Expect(ns).ToNot(BeNil())
			})
			By("creating it again should error", func() {
				ns, err := operatorClient.CreateNamespace(context.TODO(), "testns")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("namespaces \"testns\" already exists"))
				Expect(ns).To(BeNil())
			})
			By("deleting that Namespace", func() {
				err := operatorClient.DeleteNamespace(context.TODO(), "testns")
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
