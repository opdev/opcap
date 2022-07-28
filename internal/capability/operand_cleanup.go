package capability

import (
	"context"
	"log"
	"opcap/internal/operator"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (ca *capAudit) OperandCleanUp() error {
	logger.Debugw("cleaningUp operand for operator", "package", ca.subscription.Package, "channel", ca.subscription.Channel, "installmode", ca.subscription.InstallModeType)

	ca.GetAlmExamples()

	obj := &unstructured.Unstructured{Object: ca.customResources[0]}

	// using dynamic client to create Unstructured objests in k8s
	client, err := operator.NewDynamicClient()
	if err != nil {
		log.Println(err)
	}

	var crdList apiextensionsv1.CustomResourceDefinitionList
	err = ca.client.ListCRDs(context.TODO(), &crdList)
	if err != nil {
		logger.Error(err.Error())
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

	// extract name from CustomResource object and delete it
	name := obj.Object["metadata"].(map[string]interface{})["name"].(string)

	err = client.Resource(gvr).Namespace(ca.namespace).Delete(context.TODO(), name, v1.DeleteOptions{})
	if err != nil {
		logger.Infow("failed", "operandCleanUp", Resource, "package", ca.subscription.Package, "error", err.Error())
	}

	return nil
}
