package operator

import (
	"context"
	"fmt"
	"time"

	"github.com/opdev/opcap/internal/logger"
	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
)

// Gets completed CSVs, Succeeded or Failed, with timeout on delay duration
func (c operatorClient) GetCompletedCsvWithTimeout(ctx context.Context, namespace string, delay time.Duration) (operatorv1alpha1.ClusterServiceVersion, error) {
	// csv will catch CSVs from watch events
	csv := &operatorv1alpha1.ClusterServiceVersion{}
	var ok bool

	// get watcher for csv
	watcher, err := c.csvWatcher(ctx, namespace)
	if err != nil {
		return *csv, err
	}

	// eventChan receives all events from CSVs on the selected namespace
	// If a CSV changes we verify if it succeeded or failed
	eventChan := watcher.ResultChan()

	// delay and timeout to control how long we should wait for a CSV
	// to fail or succeed
	timeout := time.After(delay)

	// ticker will make the for loop below behave better for our purposes
	// we don't need a super fast sub second looping here
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		select {

		// case catches CSV events
		case event := <-eventChan:
			csv, ok = event.Object.(*operatorv1alpha1.ClusterServiceVersion)
			// fail on wrong objects
			if !ok {
				return *csv, fmt.Errorf("received unexpected object type from watch: object-type %T", event.Object)
			}
			// check for succeed or failed
			if csv.Status.Phase == operatorv1alpha1.CSVPhaseSucceeded ||
				csv.Status.Phase == operatorv1alpha1.CSVPhaseFailed {

				return *csv, nil
			}

		// if it takes more than delay return with error
		case <-timeout:
			return *csv, fmt.Errorf("operator install timeout")

		default:
			continue

		}
	}

	return operatorv1alpha1.ClusterServiceVersion{}, fmt.Errorf("unexpected error while waiting for csv")
}

// waits for CSV on namespace and gets a watcher for CSV events
func (c operatorClient) csvWatcher(ctx context.Context, namespace string) (watch.Interface, error) {
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
		logger.Errorf("Failed to create csv.")
		return nil, err
	}

	return watcher, nil
}
