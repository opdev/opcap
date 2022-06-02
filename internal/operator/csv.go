package operator

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	log "github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

func (c operatorClient) GetCSVPhase(namespace string) (operatorv1alpha1.ClusterServiceVersionPhase, error) {

	clusterServiceVersionList := operatorv1alpha1.ClusterServiceVersionList{}

	listOpts := runtimeClient.ListOptions{
		Namespace: namespace,
	}

	err := c.Client.List(context.Background(), &clusterServiceVersionList, &listOpts)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	// TODO: create a custom error for this
	if len(clusterServiceVersionList.Items) > 1 {
		return "", fmt.Errorf("more than one CSV found in dedicated namespace %s", fmt.Sprint(len(clusterServiceVersionList.Items)))
	}

	if len(clusterServiceVersionList.Items) == 0 {
		return "", fmt.Errorf("no CSV found in namespace %s", fmt.Sprint(len(clusterServiceVersionList.Items)))
	}

	clusterServiceVersion := operatorv1alpha1.ClusterServiceVersion{}

	err = c.Client.Get(context.Background(), types.NamespacedName{Name: clusterServiceVersionList.Items[0].ObjectMeta.Name, Namespace: namespace}, &clusterServiceVersion)

	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return clusterServiceVersion.Status.Phase, nil
}

func (c operatorClient) GetInstalledCSV(ctx context.Context, namespace string) (*operatorv1alpha1.ClusterServiceVersion, error) {

	// BUG(estroz): if namespace is not contained in targetNamespaces,
	// DoCSVWait will fail because the CSV is not deployed in namespace.
	nn := types.NamespacedName{
		Namespace: namespace,
	}
	log.Infof("Waiting for ClusterServiceVersion %q to reach 'Succeeded' phase", nn)
	if err := c.DoCSVWait(ctx, nn); err != nil {
		return nil, fmt.Errorf("error waiting for CSV to install: %w", err)
	}

	// TODO: check status of all resources in the desired bundle/package.
	csv := &operatorv1alpha1.ClusterServiceVersion{}
	if err := c.Client.Get(ctx, nn, csv); err != nil {
		return nil, fmt.Errorf("error getting installed CSV: %w", err)
	}
	return csv, nil
}

func (c operatorClient) DoCSVWait(ctx context.Context, key types.NamespacedName) error {
	var (
		curPhase operatorv1alpha1.ClusterServiceVersionPhase
		newPhase operatorv1alpha1.ClusterServiceVersionPhase
	)
	once := sync.Once{}

	csv := operatorv1alpha1.ClusterServiceVersion{}
	csvPhaseSucceeded := func() (bool, error) {
		err := c.Client.Get(ctx, key, &csv)
		if err != nil {
			if apierrors.IsNotFound(err) {
				once.Do(func() {
					log.Printf("  Waiting for ClusterServiceVersion %q to appear", key)
				})
				return false, nil
			}
			return false, err
		}
		newPhase = csv.Status.Phase
		if newPhase != curPhase {
			curPhase = newPhase
			log.Printf("  Found ClusterServiceVersion %q phase: %s", key, curPhase)
		}

		switch curPhase {
		case operatorv1alpha1.CSVPhaseFailed:
			return false, fmt.Errorf("csv failed: reason: %q, message: %q", csv.Status.Reason, csv.Status.Message)
		case operatorv1alpha1.CSVPhaseSucceeded:
			return true, nil
		default:
			return false, nil
		}
	}

	err := wait.PollImmediateUntil(time.Second, csvPhaseSucceeded, ctx.Done())
	if err != nil && errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return err
}
