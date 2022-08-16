package capability

import (
	"context"
	"strings"
	"time"

	"github.com/opdev/opcap/internal/operator"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func (ca *CapAudit) getAlmExamples() error {
	ctx := context.Background()

	olmClientset, err := operator.NewOlmClientset()
	if err != nil {
		return err
	}

	opts := v1.ListOptions{}

	// gets the list of CSVs present in a particular namespace
	CSVList, err := olmClientset.OperatorsV1alpha1().ClusterServiceVersions(ca.Namespace).List(ctx, opts)
	if err != nil {
		return err
	}

	almExamples := ""
	// map of string interface which consist of ALM examples from the CSVList
	if len(CSVList.Items) > 0 {
		almExamples = CSVList.Items[0].ObjectMeta.Annotations["alm-examples"]
	}
	var almList []map[string]interface{}

	err = yaml.Unmarshal([]byte(almExamples), &almList)
	if err != nil {
		return err
	}

	ca.CustomResources = almList

	return nil
}

// OperandInstall installs the operand from the ALMExamples in the ca.namespace
func (ca *CapAudit) OperandInstall() error {
	logger.Debugw("installing operand for operator", "package", ca.Subscription.Package, "channel", ca.Subscription.Channel, "installmode", ca.Subscription.InstallModeType)

	ca.getAlmExamples()

	// TODO: we need a stratergy to select which CR to select from ALMExamplesList
	if len(ca.CustomResources) > 0 {
		obj := &unstructured.Unstructured{Object: ca.CustomResources[0]}
		// using dynamic client to create Unstructured objests in k8s
		client, err := operator.NewDynamicClient()
		if err != nil {
			return err
		}

		// set the namespace of CR to the namespace of the subscription
		obj.SetNamespace(ca.Namespace)

		var crdList apiextensionsv1.CustomResourceDefinitionList
		err = ca.Client.ListCRDs(context.TODO(), &crdList)
		if err != nil {
			return err
		}

		var Resource string

		for _, crd := range crdList.Items {
			if crd.Spec.Group == obj.GroupVersionKind().Group && crd.Spec.Names.Kind == obj.GroupVersionKind().Kind {
				Resource = crd.Spec.Names.Plural
			}
		}

		gvr := schema.GroupVersionResource{
			Group:    obj.GroupVersionKind().Group,
			Version:  obj.GroupVersionKind().Version,
			Resource: Resource,
		}

		csv, _ := ca.Client.GetCompletedCsvWithTimeout(ca.Namespace, time.Minute)
		if strings.ToLower(string(csv.Status.Phase)) == "succeeded" {
			// create the resource using the dynamic client and log the error if it occurs in stdout.json
			unstructuredCR, err := client.Resource(gvr).Namespace(ca.Namespace).Create(context.TODO(), obj, v1.CreateOptions{})
			if err != nil {

				return err
			} else {
				ca.Operands = append(ca.Operands, *unstructuredCR)
			}
		} else {
			logger.Debug("exiting OperandInstall since CSV has failed")
		}
	} else {
		logger.Debug("exiting OperandInstall since no ALM_Examples found in CSV")
	}

	ca.Report(OperandInstallRptOptionFile{FilePath: "operand_install_report.json"}, OperandInstallRptOptionPrint{})

	return nil
}
