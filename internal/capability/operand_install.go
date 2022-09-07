package capability

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/opdev/opcap/internal/operator"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func (ca *capAudit) getAlmExamples() error {
	ctx := context.Background()

	olmClientset, err := operator.NewOlmClientset()
	if err != nil {
		return err
	}

	opts := v1.ListOptions{}

	// gets the list of CSVs present in a particular namespace
	CSVList, err := olmClientset.OperatorsV1alpha1().ClusterServiceVersions(ca.namespace).List(ctx, opts)
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

	ca.customResources = almList

	return nil
}

// OperandInstall installs the operand from the ALMExamples in the ca.namespace
func (ca *capAudit) OperandInstall() error {
	logger.Debugw("installing operand for operator", "package", ca.subscription.Package, "channel", ca.subscription.Channel, "installmode", ca.subscription.InstallModeType)

	ca.getAlmExamples()

	if len(ca.customResources) == 0 {
		logger.Debug("exiting OperandInstall since no ALM_Examples found in CSV")
		return nil
	}

	csv, _ := ca.client.GetCompletedCsvWithTimeout(ca.namespace, time.Minute)

	if strings.ToLower(string(csv.Status.Phase)) != "succeeded" {
		logger.Debug("exiting OperandInstall since CSV has failed")
		return nil
	}

	for _, cr := range ca.customResources {
		obj := &unstructured.Unstructured{Object: cr}
		// using dynamic client to create Unstructured objects in k8s
		client, err := operator.NewDynamicClient()
		if err != nil {
			return err
		}

		// set the namespace of CR to the namespace of the subscription
		obj.SetNamespace(ca.namespace)

		var crdList apiextensionsv1.CustomResourceDefinitionList
		err = ca.client.ListCRDs(context.TODO(), &crdList)
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

		csv, _ := ca.client.GetCompletedCsvWithTimeout(ca.namespace, time.Minute)

		if strings.ToLower(string(csv.Status.Phase)) != "succeeded" {
			logger.Debug("exiting OperandInstall since CSV has failed")
			return nil
		}

		// create the resource using the dynamic client and log the error if it occurs in stdout.json
		unstructuredCR, err := client.Resource(gvr).Namespace(ca.namespace).Create(context.TODO(), obj, v1.CreateOptions{})
		if err != nil {
			return err
		}
		ca.operands = append(ca.operands, *unstructuredCR)
	}

	file, err := os.OpenFile("operand_install_report.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		file.Close()
		return err
	}
	defer file.Close()

	ca.OperandInstallJsonReport(file)
	ca.OperandTextReport(os.Stdout)

	return nil
}
