package capability

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"opcap/internal/operator"
	"os"
	"strings"

	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

func (ca *capAudit) OperandInstall() error {
	logger.Debugw("installing operand for operator", "package", ca.subscription.Package, "channel", ca.subscription.Channel, "installmode", ca.subscription.InstallModeType)

	var csv *operatorv1alpha1.ClusterServiceVersion
	ctx := context.Background()

	olmClientset, err := operator.NewOlmClientset()
	if err != nil {
		return err
	}

	opts := v1.ListOptions{}
	var watch watch.Interface
	var ok bool

	watch, err = olmClientset.OperatorsV1alpha1().ClusterServiceVersions(ca.namespace).Watch(ctx, opts)
	if err != nil {
		return err
	}

	for event := range watch.ResultChan() {
		csv, ok = event.Object.(*operatorv1alpha1.ClusterServiceVersion)
		if !ok {
			return fmt.Errorf("received unexpected object type from watch: object-type %T", event.Object)
		}
		alm := csv.ObjectMeta.Annotations["alm-examples"]
		var almList []map[string]interface{}
		json.Unmarshal([]byte(alm), &almList)
		obj := &unstructured.Unstructured{Object: almList[0]}

		config, _ := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
		client, _ := dynamic.NewForConfig(config)

		groupVersion := strings.Split(obj.GetAPIVersion(), "/")
		gvr := schema.GroupVersionResource{
			Group:    groupVersion[0],
			Version:  groupVersion[1],
			Resource: obj.GetKind(),
		}

		_, err := client.Resource(gvr).Namespace(ca.namespace).Create(context.TODO(), obj, v1.CreateOptions{})
		if err != nil {
			log.Println(err)
		}

	}
	return nil
}
