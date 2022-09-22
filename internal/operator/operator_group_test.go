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
	operatorGroupData := OperatorGroupData{
		Name:             "testog",
		TargetNamespaces: []string{"default", "testns"},
	}

	Context("CreateOperatorGroup", func() {
		When("creating an operator group", func() {
			og, err := operatorClient.CreateOperatorGroup(context.TODO(), operatorGroupData, "testns")
			It("should succeed", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(og).ToNot(BeNil())
			})
		})
		When("creating an operator group with an existing name", func() {
			JustBeforeEach(func() {
				og, err := operatorClient.CreateOperatorGroup(context.TODO(), operatorGroupData, "testns")
				It("should error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("could not create operatorgroup: testog: operatorgroups.operators.coreos.com \"testog\" already exists"))
					Expect(og).To(BeNil())
				})
			})
		})
	})
	Context("DeleteOperatorGroup", func() {
		When("deleting the existing operator group", func() {
			err := operatorClient.DeleteOperatorGroup(context.TODO(), "testog", "testns")
			It("should delete the existing operator group", func() {
				Expect(err).ToNot(HaveOccurred())
			})
		})
		When("deleting an operator group that does not exist", func() {
			JustBeforeEach(func() {
				err := operatorClient.DeleteOperatorGroup(context.TODO(), "testog", "testns")
				It("should throw an error", func() {
					Expect(err).To(MatchError("could not delete operatorgroup: testog: namespace: testns: operatorgroups.operators.coreos.com \"testog\" not found"))
				})
			})
		})
	})
})
