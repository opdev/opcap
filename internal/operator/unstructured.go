package operator

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (c operatorClient) CreateUnstructured(ctx context.Context, obj *unstructured.Unstructured) error {
	return c.Client.Create(ctx, obj, &client.CreateOptions{})
}

func (c operatorClient) GetUnstructured(ctx context.Context, namespace, name string, obj *unstructured.Unstructured) error {
	return c.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, obj)
}

func (c operatorClient) UpdateUnstructured(ctx context.Context, obj *unstructured.Unstructured) error {
	return c.Client.Update(ctx, obj)
}

func (c operatorClient) DeleteUnstructured(ctx context.Context, obj *unstructured.Unstructured) error {
	return c.Client.Delete(ctx, obj, &client.DeleteOptions{})
}
