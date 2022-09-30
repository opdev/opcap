package operator

import (
	"context"

	v1 "github.com/openshift/api/config/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetOpenShiftVersion uses the OpenShift Config clientset to get a ClusterVersion resource which has the
// version of an OpenShift cluster
func (c operatorClient) GetOpenShiftVersion(ctx context.Context) (string, error) {
	// version is the OpenShift version of the cluster
	var version string

	var cversions v1.ClusterVersionList
	if err := c.Client.List(ctx, &cversions, &client.ListOptions{}); err != nil {
		version = "0.0.0"
		return version, err
	}

	for _, cversion := range cversions.Items {
		histories := cversion.Status.History
		for _, history := range histories {
			version = history.Version
		}
	}

	return version, nil
}
