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

func (ca *capAudit) OperandInstall() error {
	logger.Debugw("installing operand for operator", "package", ca.subscription.Package, "channel", ca.subscription.Channel, "installmode", ca.subscription.InstallModeType)

	ca.GetAlmExamples()

	// TODO: we need a stratergy to select which CR to select from ALMExamplesList
	obj := &unstructured.Unstructured{Object: ca.customResources[0]}

	// using dynamic client to create Unstructured objests in k8s
	client, err := operator.NewDynamicClient()
	if err != nil {
		log.Println(err)
	}

	obj.SetNamespace(ca.namespace)

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

	_, err = client.Resource(gvr).Namespace(ca.namespace).Create(context.TODO(), obj, v1.CreateOptions{})
	if err != nil {
		logger.Infow("failed", "operand", Resource, "package", ca.subscription.Package, "error", err.Error())
	} else {
		logger.Infow("succeeded", "operand", Resource, "package", ca.subscription.Package)
	}

	return nil
}
