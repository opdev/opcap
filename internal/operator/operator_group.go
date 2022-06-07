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
	logger.Debugf("Creating OperatorGroup %s in namespace %s", data.Name, namespace)
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
		logger.Errorf("%w: error while creating OperatorGroup: %s", err, data.Name)
		return nil, err
	}

	logger.Debugf("OperatorGroup %s is created successfully in namespace %s", data.Name, namespace)
	return operatorGroup, nil
}

func (o *operatorClient) DeleteOperatorGroup(ctx context.Context, name string, namespace string) error {
	logger.Debugf("Deleting OperatorGroup %s in namespace %s", name, namespace)
	operatorGroup := operatorv1.OperatorGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	err := o.Client.Delete(ctx, &operatorGroup)
	if err != nil {
		logger.Errorf("%w: error while deleting OperatorGroup: %s in namespace: %s", err, name, namespace)
		return err
	}

	logger.Debugf("OperatorGroup %s is deleted successfully from namespace %s", name, namespace)
	return nil
}
