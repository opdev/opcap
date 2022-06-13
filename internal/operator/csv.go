package operator

import (
	"context"
	"fmt"
	"time"

	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
)

func (c operatorClient) WaitForCsvOnNamespace(namespace string) (string, error) {
	ctx := context.Background()
	var watcher watch.Interface

	err := wait.ExponentialBackoff(wait.Backoff{Steps: 3, Duration: 2 * time.Second, Factor: 5, Cap: 90 * time.Second},
		func() (bool, error) {

			var err error

			olmClientset, err := NewOlmClientset()
			if err != nil {
				return false, err
			}

			opts := v1.ListOptions{}

			watcher, err = olmClientset.OperatorsV1alpha1().ClusterServiceVersions(namespace).Watch(ctx, opts)
			if err != nil {
				return false, err
			}

			return true, nil

		})
	if err != nil {
		logger.Error("Failed to create csv.")
		return "", err
	}

	var csv *operatorv1alpha1.ClusterServiceVersion
	var ok bool

	for event := range watcher.ResultChan() {
		csv, ok = event.Object.(*operatorv1alpha1.ClusterServiceVersion)
		if !ok {
			return "", fmt.Errorf("received unexpected object type from watch: object-type %T", event.Object)
		}
		if csv.Status.Phase == operatorv1alpha1.CSVPhaseSucceeded ||
			csv.Status.Phase == operatorv1alpha1.CSVPhaseFailed {

			break
		}

	}

	return string(csv.Status.Phase), nil
}
