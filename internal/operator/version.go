package operator

import (
	"context"

	configv1 "github.com/openshift/api/config/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetOpenShiftVersion uses the OpenShift Config clientset to get a ClusterVersion resource which has the
// version of an OpenShift cluster
func (c operatorClient) GetOpenShiftVersion(ctx context.Context) (string, error) {
	// version is the OpenShift version of the cluster
	var version string
	// The current version of the cluster is represented in the obj named version
	objKey := client.ObjectKey{
		Name: "version",
	}
	clusterVersion := configv1.ClusterVersion{}
	if err := c.Client.Get(ctx, objKey, &clusterVersion); err != nil {
		version = "0.0.0"
		return version, err
	}
	// History is ordered by recency per `oc explain clusterversion.status.history`
	// The newest update is first in the list.
	version = clusterVersion.Status.History[0].Version

	return version, nil
}
