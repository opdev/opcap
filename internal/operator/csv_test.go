package operator

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("CSV", func() {
	var client operatorClient
	var csv operatorv1alpha1.ClusterServiceVersion

	BeforeEach(func() {
		csv = operatorv1alpha1.ClusterServiceVersion{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "testcsv",
				Namespace: "testns",
			},
			Status: operatorv1alpha1.ClusterServiceVersionStatus{
				Phase: operatorv1alpha1.CSVPhaseInstalling,
			},
		}

		scheme := runtime.NewScheme()
		Expect(addSchemes(scheme)).To(Succeed())

		fakeClient := fake.NewClientBuilder().
			WithObjects(&csv).
			WithScheme(scheme).
			Build()

		client = operatorClient{
			Client: fakeClient,
		}
	})
	When("testing for a CSV", func() {
		When("the CSV is updated", func() {
			It("should get a completed CSV", func() {
				var resultCsv *operatorv1alpha1.ClusterServiceVersion
				var err error
				done := make(chan interface{})
				go func() {
					resultCsv, err = client.GetCompletedCsvWithTimeout(context.Background(), "testns", time.Second*30)
					close(done)
				}()

				// Allow some time for the CSV method to get going
				time.Sleep(time.Millisecond)
				updatedCsv := csv.DeepCopy()
				updatedCsv.Status.Phase = operatorv1alpha1.CSVPhaseSucceeded
				Expect(client.Client.Status().Update(
					context.Background(),
					updatedCsv,
					&runtimeClient.UpdateOptions{},
				)).To(BeNil())
				Expect(updatedCsv).ToNot(BeNil())

				Eventually(done, time.Second*60).Should(BeClosed())
				Expect(err).ToNot(HaveOccurred())
				Expect(resultCsv).ToNot(BeNil())
				Expect(resultCsv).To(Equal(updatedCsv))
			})
		})
		When("no CSV is updated", func() {
			It("should timeout", func() {
				_, err := client.GetCompletedCsvWithTimeout(context.Background(), "testns", time.Second*2)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(TimeoutError))
			})
		})
	})
	When("deleting a CSV", func() {
		When("the CSV exists", func() {
			It("should delete the CSV", func() {
				Expect(client.DeleteCSV(context.Background(), csv.ObjectMeta.Name, csv.ObjectMeta.Namespace)).To(Succeed())
			})
		})
		When("the CSV does not exist", func() {
			It("should return an error", func() {
				Expect(client.DeleteCSV(context.Background(), "notfound", csv.ObjectMeta.Namespace)).ToNot(Succeed())
			})
		})
	})
})
