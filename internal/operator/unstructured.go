package operator

import (
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (c operatorClient) CreateUnstructured(ctx context.Context, namespace string, obj *unstructured.Unstructured, gvr schema.GroupVersionResource) (*unstructured.Unstructured, error) {
	return c.DynamicClient.Resource(gvr).Namespace(namespace).Create(ctx, obj, v1.CreateOptions{})
}

func (c operatorClient) GetUnstructured(ctx context.Context, namespace, name string, gvr schema.GroupVersionResource) (*unstructured.Unstructured, error) {
	return c.DynamicClient.Resource(gvr).Namespace(namespace).Get(ctx, name, v1.GetOptions{})
}

func (c operatorClient) DeleteUnstructured(ctx context.Context, namespace, name string, gvr schema.GroupVersionResource) error {
	return c.DynamicClient.Resource(gvr).Namespace(namespace).Delete(ctx, name, v1.DeleteOptions{})
}
