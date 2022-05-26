package operator

import (
	"context"

	operatorv1 "github.com/operator-framework/api/pkg/operators/v1"
	corev1 "k8s.io/api/core/v1"
)

type OperatorGroupData struct {
	Name             string
	TargetNamespaces []string
}

type Client interface {
	CreateOperatorGroup(ctx context.Context, data OperatorGroupData, namespace string) (*operatorv1.OperatorGroup, error)
	DeleteOperatorGroup(ctx context.Context, name string, namespace string) error
	CreateSecret(ctx context.Context, name string, content map[string]string, secretType corev1.SecretType, namespace string) (*corev1.Secret, error)
	DeleteSecret(ctx context.Context, name string, namespace string) error
}
