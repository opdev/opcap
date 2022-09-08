package operator

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Namespace", func() {
	var oc Client
	var err error

	BeforeEach(func() {
		oc, err = NewOpCapClient()
		Expect(err).ToNot(HaveOccurred())
	})
	Context("Namespaces", func() {
		It("should exercise Namespaces", func() {
			By("creating a Namespace", func() {
				ns, err := oc.CreateNamespace(context.TODO(), "testns")
				Expect(err).ToNot(HaveOccurred())
				Expect(ns).ToNot(BeNil())
			})
			By("creating it again should error", func() {
				ns, err := oc.CreateNamespace(context.TODO(), "testns")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("namespaces \"testns\" already exists"))
				Expect(ns).To(BeNil())
			})
			By("deleting that Namespace", func() {
				err := oc.DeleteNamespace(context.TODO(), "testns")
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
