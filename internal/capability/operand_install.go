package capability

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/opdev/opcap/internal/logger"
	"github.com/opdev/opcap/internal/operator"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func extractAlmExamples(ctx context.Context, options *options) error {
	olmClientset, err := operator.NewOlmClientset()
	if err != nil {
		return err
	}

	opts := v1.ListOptions{}

	// gets the list of CSVs present in a particular namespace
	CSVList, err := olmClientset.OperatorsV1alpha1().ClusterServiceVersions(options.namespace).List(ctx, opts)
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

	options.customResources = almList

	return nil
}

// OperandInstall installs the operand from the ALMExamples in the ca.namespace
func operandInstall(ctx context.Context, opts ...auditOption) auditFn {
	var options options
	for _, opt := range opts {
		err := opt(&options)
		if err != nil {
			return func(_ context.Context) error {
				return fmt.Errorf("option failed: %v", err)
			}
		}
	}

	return func(ctx context.Context) error {
		logger.Debugw("installing operand for operator", "package", options.Subscription.Package, "channel", options.Subscription.Channel, "installmode", options.Subscription.InstallModeType)

		if err := extractAlmExamples(ctx, &options); err != nil {
			logger.Errorf("Failed getting ALM Examples")
		}

		if len(options.customResources) == 0 {
			logger.Debugf("exiting OperandInstall since no ALM_Examples found in CSV")
			return nil
		}

		csv, _ := options.client.GetCompletedCsvWithTimeout(ctx, options.namespace, time.Minute)

		if strings.ToLower(string(csv.Status.Phase)) != "succeeded" {
			return fmt.Errorf("exiting OperandInstall since CSV install has failed")
		}

		for _, cr := range options.customResources {
			obj := &unstructured.Unstructured{Object: cr}
			// using dynamic client to create Unstructured objects in k8s
			client, err := operator.NewDynamicClient()
			if err != nil {
				return err
			}

			// set the namespace of CR to the namespace of the subscription
			obj.SetNamespace(options.namespace)

			var crdList apiextensionsv1.CustomResourceDefinitionList
			err = options.client.ListCRDs(ctx, &crdList)
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

			csv, _ := options.client.GetCompletedCsvWithTimeout(ctx, options.namespace, time.Minute)

			if strings.ToLower(string(csv.Status.Phase)) != "succeeded" {
				logger.Debugf("exiting OperandInstall since CSV has failed")
				return nil
			}

			// create the resource using the dynamic client and log the error if it occurs in stdout.json
			unstructuredCR, err := client.Resource(gvr).Namespace(options.namespace).Create(ctx, obj, v1.CreateOptions{})
			if err != nil {
				return err
			}
			options.operands = append(options.operands, *unstructuredCR)
		}

		file, err := os.OpenFile("operand_install_report.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		defer file.Close()

		if err := operandInstallJsonReport(file, options); err != nil {
			return fmt.Errorf("could not generate operand install JSON report: %v", err)
		}

		if err := operandInstallTextReport(os.Stdout, options); err != nil {
			return fmt.Errorf("could not generate operand install text report: %v", err)
		}

		return nil
	}
}
