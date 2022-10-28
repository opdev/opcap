package report

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/opdev/opcap/internal/operator"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var _ = Describe("Report", func() {
	Describe("Template process test", func() {
		When("an invalid template is passed", func() {
			It("should return an error", func() {
				Expect(processTemplate(&strings.Builder{}, "{{ .Invalid }", &struct{}{})).ToNot(Succeed())
			})
		})
		When("an the writer errors", func() {
			It("should return an error", func() {
				Expect(processTemplate(errWriter(0), "this is the template", &struct{}{})).ToNot(Succeed())
			})
		})
	})
	Describe("Report template tests", func() {
		var w strings.Builder
		var data TemplateData

		BeforeEach(func() {
			DeferCleanup(w.Reset)
			data = TemplateData{
				OcpVersion: "",
				Subscription: operator.SubscriptionData{
					Name:            "testsub",
					Channel:         "test",
					CatalogSource:   "testcatalog",
					Package:         "testpackage",
					InstallModeType: "v1alpha1.InstallModeTypeAllNamespaces",
				},
				Csv: &v1alpha1.ClusterServiceVersion{
					Status: v1alpha1.ClusterServiceVersionStatus{
						Phase:   v1alpha1.CSVPhaseSucceeded,
						Message: "message",
						Reason:  v1alpha1.CSVReasonInstallSuccessful,
					},
				},
				CsvTimeout: false,
				CustomResources: []map[string]interface{}{
					{
						"kind": "testkind",
						"metadata": map[string]interface{}{
							"name": "testname",
						},
					},
				},
				Operands: []unstructured.Unstructured{
					{
						Object: map[string]interface{}{
							"metadata": map[string]interface{}{
								"kind": "testkind",
								"name": "testname",
							},
						},
					},
				},
				OperandCount: 1,
			}
		})
		Context("Operator reports", func() {
			When("generating a JSON report", func() {
				When("given successful data", func() {
					It("should create a valid JSON report", func() {
						Expect(OperatorInstallJsonReport(&w, data)).To(Succeed())
						Expect(w.String()).To(MatchJSON(`{"level":"info","message":"Succeeded","package":"testpackage","channel":"test","installmode":"AllNamespaces"}`))
					})
				})
				When("given a timeout", func() {
					BeforeEach(func() {
						data.CsvTimeout = true
					})
					It("should report a timeout", func() {
						Expect(OperatorInstallJsonReport(&w, data)).To(Succeed())
						Expect(w.String()).To(MatchJSON(`{"level":"info","message":"timeout","package":"testpackage","channel":"test","installmode":"AllNamespaces"}`))
					})
				})
			})
			When("generating a text report", func() {
				When("given successful data", func() {
					It("should print a report", func() {
						Expect(OperatorInstallTextReport(&w, data)).To(Succeed())
						Expect(w.String()).To(ContainSubstring("OpenShift Version: %s", "4.11"))
						Expect(w.String()).To(ContainSubstring("Package Name: %s", "testpackage"))
						Expect(w.String()).To(ContainSubstring("Channel: %s", "test"))
						Expect(w.String()).To(ContainSubstring("Catalog Source: %s", "testcatalog"))
						Expect(w.String()).To(ContainSubstring("Install Mode: %s", "AllNamespaces"))
						Expect(w.String()).To(ContainSubstring("Result: %s", "Succeeded"))
						Expect(w.String()).To(ContainSubstring("Message: %s", "message"))
						Expect(w.String()).To(ContainSubstring("Reason: %s", "InstallSucceeded"))
					})
				})
				When("given a timeout", func() {
					BeforeEach(func() {
						data.CsvTimeout = true
					})
					It("should report a timeout", func() {
						Expect(OperatorInstallTextReport(&w, data)).To(Succeed())
						Expect(w.String()).To(ContainSubstring("Result: %s", "timeout"))
					})
				})
			})
		})
		Context("Operand reports", func() {
			When("generating a JSON report", func() {
				When("given successful data", func() {
					It("should create a valid JSON report", func() {
						Expect(OperandInstallJsonReport(&w, data)).To(Succeed())
						Expect(w.String()).To(MatchJSON(`{"package":"testpackage","Operand Kind":"testkind","Operand Name":"testname","message":"created"}`))
					})
				})
				When("given no operands", func() {
					BeforeEach(func() {
						data.Operands = []unstructured.Unstructured{}
					})
					It("should report failed", func() {
						Expect(OperandInstallJsonReport(&w, data)).To(Succeed())
						Expect(w.String()).To(MatchJSON(`{"package":"testpackage","Operand Kind":"testkind","Operand Name":"testname","message":"failed"}`))
					})
				})
			})
			When("generating a text report", func() {
				When("given successful data", func() {
					It("should print a report", func() {
						Expect(OperandInstallTextReport(&w, data)).To(Succeed())
						Expect(w.String()).To(ContainSubstring("OpenShift Version: %s", "4.11"))
						Expect(w.String()).To(ContainSubstring("Package Name: %s", "testpackage"))
						Expect(w.String()).To(ContainSubstring("Operand Kind: %s", "testkind"))
						Expect(w.String()).To(ContainSubstring("Operand Name: %s", "testname"))
						Expect(w.String()).To(ContainSubstring("Operand Creation: %s", "Succeeded"))
					})
				})
				When("given no operands", func() {
					BeforeEach(func() {
						data.Operands = []unstructured.Unstructured{}
					})
					It("should report failed", func() {
						Expect(OperandInstallTextReport(&w, data)).To(Succeed())
						Expect(w.String()).To(ContainSubstring("Operand Creation: %s", "Failed"))
					})
				})
			})
		})
	})
})

type errWriter int

func (errWriter) Write(p []byte) (int, error) {
	return 0, fmt.Errorf("write error")
}
