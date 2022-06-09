package operator

import (
	"context"
	configv1 "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

// getConfigClient returns a OpenShift Config clientset used to get the ClusterVersion resource
// in order to determine the version of the OpenShift cluster used during opcap run
func getConfigClient() *configv1.ConfigV1Client {
	// create openshift config clientset
	cfg, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	if err != nil {
		logger.Errorf("Unable to build config from flags: %s", err)
	}

	clientset, err := configv1.NewForConfig(cfg)
	if err != nil {
		logger.Errorf("Unable to create new OpenShift Config clientset: %s", err)
	}

	return clientset
}

// GetOpenShiftVersion uses the OpenShift Config clientset to get a ClusterVersion resource which has the
// version of an OpenShift cluster
func (c operatorClient) GetOpenShiftVersion() (string, error) {
	// version is the OpenShift version of the cluster
	var version string

	// OpenShift Config client used to talk with the OpenShift API
	configClient := getConfigClient()

	cversions, err := configClient.ClusterVersions().List(context.TODO(), metav1.ListOptions{})
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
