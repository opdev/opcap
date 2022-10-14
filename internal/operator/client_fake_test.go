package operator

import (
	"github.com/opdev/opcap/internal/logger"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimefake "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type fakeOpClient struct {
	Client client.WithWatch
}

func NewFakeOpClient(initObjs ...runtime.Object) Client {
	scheme := runtime.NewScheme()
	err := addSchemes(scheme)
	if err != nil {
		logger.Errorf("could not create scheme: %v", err)
		return nil
	}

	return &operatorClient{
		Client: runtimefake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(initObjs...).Build(),
	}
}
