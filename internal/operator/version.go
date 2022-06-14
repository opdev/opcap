package operator

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetOpenShiftVersion uses the OpenShift Config clientset to get a ClusterVersion resource which has the
// version of an OpenShift cluster
func (c operatorClient) GetOpenShiftVersion() (string, error) {
	// version is the OpenShift version of the cluster
	var version string

	cversions, err := c.KubeClientSet.ClusterVersions().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
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
