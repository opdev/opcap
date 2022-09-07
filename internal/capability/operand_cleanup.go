package capability

import (
	"context"

	"github.com/opdev/opcap/internal/operator"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// OperandCleanup removes the operand from the OCP cluster in the ca.namespace
func (ca *capAudit) OperandCleanUp(ctx context.Context) error {
	logger.Debugw("cleaningUp operand for operator", "package", ca.subscription.Package, "channel", ca.subscription.Channel, "installmode",
		ca.subscription.InstallModeType)

	if len(ca.customResources) > 0 {
		for _, cr := range ca.customResources {
			obj := &unstructured.Unstructured{Object: cr}

			// using dynamic client to create Unstructured objests in k8s
			client, err := operator.NewDynamicClient()
			if err != nil {
				return err
			}

			var crdList apiextensionsv1.CustomResourceDefinitionList
			err = ca.client.ListCRDs(ctx, &crdList)
			if err != nil {
				return err
			}

			var Resource string

			// iterate over the CRD list to find the CRD for the resource we are trying to delete
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

			// extract name from CustomResource object and delete it
			name := obj.Object["metadata"].(map[string]interface{})["name"].(string)

			// check if CR exists, only then cleanup the operand
			crInstance, _ := client.Resource(gvr).Namespace(ca.namespace).Get(ctx, name, v1.GetOptions{})
			if crInstance != nil {
				// delete the resource using the dynamic client
				err = client.Resource(gvr).Namespace(ca.namespace).Delete(ctx, name, v1.DeleteOptions{})
				if err != nil {
					logger.Debugf("failed operandCleanUp: %s package: %s error: %s\n", Resource, ca.subscription.Package, err.Error())
					return err
				}
			}
		}
	}

	return nil
}
