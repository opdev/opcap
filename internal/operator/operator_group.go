package operator

import (
	"context"
	"fmt"

	apiruntime "k8s.io/apimachinery/pkg/runtime"

	operatorv1 "github.com/operator-framework/api/pkg/operators/v1"
	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

type operatorClient struct {
	Client runtimeClient.Client
}

func NewClient(client runtimeClient.Client) Client {
	var osclient Client = &operatorClient{
		Client: client,
	}
	return osclient
}

func AddSchemes(scheme *apiruntime.Scheme) error {
	if err := operatorv1.AddToScheme(scheme); err != nil {
		return err
	}
	if err := operatorv1alpha1.AddToScheme(scheme); err != nil {
		return err
	}
	return nil
}

func (oe *operatorClient) CreateOperatorGroup(ctx context.Context, data OperatorGroupData, namespace string) (*operatorv1.OperatorGroup, error) {
	logger.Infof("Creating OperatorGroup %s in namespace %s", data.Name, namespace)
	operatorGroup := &operatorv1.OperatorGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      data.Name,
			Namespace: namespace,
		},
		Spec: operatorv1.OperatorGroupSpec{
			TargetNamespaces: data.TargetNamespaces,
		},
	}
	err := oe.Client.Create(ctx, operatorGroup)
	if err != nil {
		logger.Error(fmt.Errorf("%w: error while creating OperatorGroup: %s", err, data.Name))
		return nil, err
	}

	logger.Infof("OperatorGroup %s is created successfully in namespace %s", data.Name, namespace)
	return operatorGroup, nil
}

func (oe *operatorClient) DeleteOperatorGroup(ctx context.Context, name string, namespace string) error {
	logger.Infof("Deleting OperatorGroup %s in namespace %s", name, namespace)
	operatorGroup := operatorv1.OperatorGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	err := oe.Client.Delete(ctx, &operatorGroup)
	if err != nil {
		logger.Error(fmt.Errorf("%w: error while deleting OperatorGroup: %s in namespace: %s", err, name, namespace))
		return err
	}

	logger.Infof("OperatorGroup %s is deleted successfully from namespace %s", name, namespace)
	return nil
}
