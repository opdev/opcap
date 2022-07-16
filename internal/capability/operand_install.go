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

func OperandInstallForOperator(namespace string) error {
	var csv *operatorv1alpha1.ClusterServiceVersion
	ctx := context.Background()

	olmClientset, err := operator.NewOlmClientset()
	if err != nil {
		return err
	}

	opts := v1.ListOptions{}
	var watch watch.Interface
	var ok bool

	watch, err = olmClientset.OperatorsV1alpha1().ClusterServiceVersions(namespace).Watch(ctx, opts)
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
		obj.SetNamespace(namespace)

		groupVersion := strings.Split(obj.GetAPIVersion(), "/")
		gvr := schema.GroupVersionResource{
			Group:    groupVersion[0],
			Version:  groupVersion[1],
			Resource: strings.ToLower(obj.GetKind()),
		}

		_, err := client.Resource(gvr).Namespace(obj.GetNamespace()).Create(context.TODO(), obj, v1.CreateOptions{})
		if err != nil {
			log.Println(err)
		}

	}

	return nil
}
