package capability

import (
	"context"
	"log"
	"opcap/internal/operator"
	"strings"

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

	groupVersion := strings.Split(obj.GetAPIVersion(), "/")
	gvr := schema.GroupVersionResource{
		Group:    groupVersion[0],
		Version:  groupVersion[1],
		Resource: obj.GetKind(),
	}

	_, err = client.Resource(gvr).Namespace(ca.namespace).Create(context.TODO(), obj, v1.CreateOptions{})
	if err != nil {
		log.Println(err)
	}

	return nil
}
