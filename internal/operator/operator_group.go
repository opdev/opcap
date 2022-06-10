package operator

import (
	"context"

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
		logger.Errorf("error while creating operatorgroup %s: %w", data.Name, err)
		return nil, err
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
		logger.Errorf("error while deleting OperatorGroup %s in namespace %s: %w", name, namespace, err)
		return err
	}

	logger.Debugw("operatorgroup deleted", "operatorgroup", name, "namespace", namespace)
	return nil
}
