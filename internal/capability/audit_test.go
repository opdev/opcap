package capability

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/opdev/opcap/internal/operator"
	configv1 "github.com/openshift/api/config/v1"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ = Describe("Audit tests", func() {
	Context("creating a new audit", func() {
		It("should truncate the namespace if it is too long", func() {
			now := metav1.Now()
			ver := configv1.ClusterVersion{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ClusterVersion",
					APIVersion: "config.openshift.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "version",
				},
				Spec: configv1.ClusterVersionSpec{},
				Status: configv1.ClusterVersionStatus{
					History: []configv1.UpdateHistory{
						{
							Version:        "4.11",
							CompletionTime: &now,
							StartedTime:    now,
						},
					},
				},
			}
			sub := operator.SubscriptionData{
				Package:         "thisisareallylongnamethatwillneedtobetrimmedandwillotherwisecauseanerror",
				InstallModeType: v1alpha1.InstallModeTypeAllNamespaces,
				Name:            "testname",
				Channel:         "testchannel",
			}
			expectedNamespace := "opcap-thisisareallylongnamethatwillneedtobetrimme-allnamespaces"
			client := operator.NewFakeOpClient([]runtime.Object{&ver}...)
			audit, err := newCapAudit(context.Background(), client, sub, []string{}, []map[string]interface{}{})
			Expect(err).ToNot(HaveOccurred())
			Expect(audit.namespace).To(Equal(expectedNamespace))
			Expect(len(audit.namespace)).To(Equal(63))
		})
	})
})
