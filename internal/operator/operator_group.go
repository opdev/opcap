package operator

import (
	"context"
	"fmt"

	"github.com/opdev/opcap/internal/logger"

	operatorv1 "github.com/operator-framework/api/pkg/operators/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type OperatorGroupData struct {
	Name             string
	TargetNamespaces []string
}

func (o *operatorClient) CreateOperatorGroup(ctx context.Context, data OperatorGroupData, namespace string) (*operatorv1.OperatorGroup, error) {
	logger.Debugw("creating OperatorGroup", "operatorgroup", data.Name, "namespace", namespace)
	operatorGroup := &operatorv1.OperatorGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      data.Name,
			Namespace: namespace,
		},
		Spec: operatorv1.OperatorGroupSpec{
			TargetNamespaces: data.TargetNamespaces,
		},
	}
	err := o.Client.Create(ctx, operatorGroup)
	if err != nil {
		return nil, fmt.Errorf("could not create operatorgroup: %s: %v", data.Name, err)
	}

	logger.Debugw("operatorgroup created", "operatorgroup", data.Name, "namespace", namespace)
	return operatorGroup, nil
}

func (o *operatorClient) DeleteOperatorGroup(ctx context.Context, name string, namespace string) error {
	logger.Debugw("deleting operatorgroup", "operatorgroup", name, "namespace", namespace)
	operatorGroup := operatorv1.OperatorGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	err := o.Client.Delete(ctx, &operatorGroup)
	if err != nil {
		return fmt.Errorf("could not delete operatorgroup: %s: namespace: %s: %v", name, namespace, err)
	}

	logger.Debugw("operatorgroup deleted", "operatorgroup", name, "namespace", namespace)
	return nil
}
