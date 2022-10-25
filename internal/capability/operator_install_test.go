package capability

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/opdev/opcap/internal/operator"
	v1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Operator install tests", func() {
	var csv v1alpha1.ClusterServiceVersion
	var fakeClient operator.Client
	var operatorGroupData operator.OperatorGroupData
	var subscription operator.SubscriptionData
	var customResources []map[string]interface{}
	BeforeEach(func() {
		// Set up and clean up temporary directory for CSV created in operator_install
		tmpDir, err := os.MkdirTemp("", "operator-install-*")
		Expect(err).ToNot(HaveOccurred())
		DeferCleanup(os.RemoveAll, tmpDir)
		cwd, err := os.Getwd()
		Expect(err).ToNot(HaveOccurred())
		DeferCleanup(os.Chdir, cwd)
		os.Chdir(tmpDir)

		csv = v1alpha1.ClusterServiceVersion{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "testcsv",
				Namespace: "testns",
			},
			Status: v1alpha1.ClusterServiceVersionStatus{
				Phase: v1alpha1.CSVPhaseSucceeded,
			},
		}

		fakeClient = operator.NewFakeOpClient(&csv)

		operatorGroupData = operator.OperatorGroupData{
			Name:             "testog",
			TargetNamespaces: []string{"default", "testns"},
		}

		subscription = operator.SubscriptionData{
			Name:            "testsub",
			Channel:         "test",
			InstallModeType: v1alpha1.InstallModeTypeAllNamespaces,
			Package:         "testpackage",
			CatalogSource:   "testcatalog",
		}

		customResources = []map[string]interface{}{
			{
				"kind": "testkind",
				"metadata": map[string]interface{}{
					"name": "testname",
				},
			},
		}
	})
	Context("creating a new operator install audit", func() {
		When("given valid options", func() {
			It("should return functional auditFn and auditCleanupFn", func() {
				auditFn, auditCleanupFn := operatorInstall(context.TODO(),
					withClient(fakeClient),
					withNamespace("testingThings"),
					withOperatorGroupData(&operatorGroupData),
					withSubscription(&subscription),
					withTimeout(1),
					withCustomResources(customResources),
				)
				ctx := context.TODO()
				err := auditFn(ctx)
				Expect(err).To(Not(HaveOccurred()))
				Expect(auditCleanupFn(context.TODO())).To(Succeed())
			})
		})
		When("given an invalid option to add", func() {
			It("should return an auditFn that gives an error and an auditCleanupFn that returns nil", func() {
				auditFn, auditCleanupFn := operatorInstall(context.TODO(),
					withClient(fakeClient),
					withNamespace(""),
					withOperatorGroupData(&operatorGroupData),
					withSubscription(&subscription),
					withTimeout(1),
					withCustomResources(customResources),
				)
				Expect(auditFn(context.TODO())).ToNot(Succeed())
				Expect(auditCleanupFn(context.TODO())).To(BeNil())
			})
		})
	})
})
