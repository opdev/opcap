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
	var operatorClient operatorClient
	var operatorGroupData OperatorGroupData

	BeforeEach(func() {
		logger.InitLogger("debug")
		scheme := runtime.NewScheme()
		operatorv1.AddToScheme(scheme)
		client := fake.NewClientBuilder().WithScheme(scheme).Build()
		operatorClient.Client = client
		operatorGroupData = OperatorGroupData{
			Name:             "testog",
			TargetNamespaces: []string{"default", "testns"},
		}
	})

	Context("CreateOperatorGroup", func() {
		When("creating an operator group", func() {
			It("should succeed", func() {
				og, err := operatorClient.CreateOperatorGroup(context.TODO(), operatorGroupData, "testns")
				Expect(err).ToNot(HaveOccurred())
				Expect(og).ToNot(BeNil())
			})
		})
		When("creating an operator group with an existing name", func() {
			JustBeforeEach(func() {
				_, err := operatorClient.CreateOperatorGroup(context.TODO(), operatorGroupData, "testns")
				Expect(err).ToNot(HaveOccurred())
			})
			It("should error", func() {
				og, err := operatorClient.CreateOperatorGroup(context.TODO(), operatorGroupData, "testns")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("could not create operatorgroup: testog: operatorgroups.operators.coreos.com \"testog\" already exists"))
				Expect(og).To(BeNil())
			})
		})
	})
	Context("DeleteOperatorGroup", func() {
		When("deleting the existing operator group", func() {
			JustBeforeEach(func() {
				_, err := operatorClient.CreateOperatorGroup(context.TODO(), operatorGroupData, "testns")
				Expect(err).ToNot(HaveOccurred())
			})
			It("should delete the existing operator group", func() {
				Expect(operatorClient.DeleteOperatorGroup(context.TODO(), "testog", "testns")).To(Succeed())
			})
		})
		When("deleting an operator group that does not exist", func() {
			It("should throw an error", func() {
				Expect(operatorClient.DeleteOperatorGroup(context.TODO(), "testog", "testns")).ToNot(Succeed())
			})
		})
	})
})
