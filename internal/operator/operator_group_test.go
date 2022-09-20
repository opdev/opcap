package operator

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/opdev/opcap/internal/logger"
	operatorv1 "github.com/operator-framework/api/pkg/operators/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("OperatorGroup", func() {
	logger.InitLogger("debug")
	scheme := runtime.NewScheme()
	operatorv1.AddToScheme(scheme)
	client := fake.NewClientBuilder().WithScheme(scheme).Build()
	var operatorClient operatorClient = operatorClient{
		Client: client,
	}

	Context("OperatorGroup", func() {
		It("should exercise OperatorGroup", func() {
			operatorGroupData := OperatorGroupData{
				Name:             "testog",
				TargetNamespaces: []string{"default", "testns"},
			}
			By("creating a OperatorGroup", func() {
				og, err := operatorClient.CreateOperatorGroup(context.TODO(), operatorGroupData, "testns")
				Expect(err).ToNot(HaveOccurred())
				Expect(og).ToNot(BeNil())
			})
			By("creating it again should error", func() {
				og, err := operatorClient.CreateOperatorGroup(context.TODO(), operatorGroupData, "testns")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("could not create operatorgroup: testog: operatorgroups.operators.coreos.com \"testog\" already exists"))
				Expect(og).To(BeNil())
			})
			By("deleting that OperatorGroup", func() {
				err := operatorClient.DeleteOperatorGroup(context.TODO(), "testog", "testns")
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
