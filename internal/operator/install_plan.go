package operator

import (
	"context"
	"fmt"
	"time"

	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ApproveInstallPlan waits and approves installPlans
func (c operatorClient) ApproveInstallPlan(ctx context.Context, sub *operatorv1alpha1.Subscription) error {
	subKey := types.NamespacedName{
		Namespace: sub.GetNamespace(),
		Name:      sub.GetName(),
	}

	ipCheck := wait.ConditionFunc(func() (done bool, err error) {
		if err := c.Client.Get(ctx, subKey, sub); err != nil {
			return false, err
		}
		if sub.Status.InstallPlanRef != nil {
			return true, nil
		}
		return false, nil
	})

	if err := wait.PollImmediateUntil(200*time.Millisecond, ipCheck, ctx.Done()); err != nil {
		logger.Errorf("install plan is not available for the subscription %s: %w", sub.Name, err)
		return fmt.Errorf("install plan is not available for the subscription %s: %v", sub.Name, err)
	}

	installPlanList := operatorv1alpha1.InstallPlanList{}
	namespace := sub.GetNamespace()

	listOpts := runtimeClient.ListOptions{
		Namespace: namespace,
	}

	err := c.Client.List(ctx, &installPlanList, &listOpts)
	if err != nil {
		logger.Errorf("Unable to list InstallPlans in Namespace %s: %w", namespace, err)
		return err
	}

	if len(installPlanList.Items) == 0 {
		logger.Errorf("no installPlan found in namespace %s", namespace)
		return fmt.Errorf("no installPlan found in namespace %s", namespace)
	}

	installPlan := operatorv1alpha1.InstallPlan{}

	err = c.Client.Get(ctx, types.NamespacedName{Name: installPlanList.Items[0].ObjectMeta.Name, Namespace: namespace}, &installPlan)
	if err != nil {
		logger.Errorf("no installPlan found in namespace %s: %w", namespace, err)
		return err
	}

	if installPlan.Spec.Approval == operatorv1alpha1.ApprovalManual {
		installPlan.Spec.Approved = true
		err := c.Client.Update(ctx, &installPlan)
		if err != nil {
			logger.Errorf("Unable to approve installPlan %s in namespace %s: %w", installPlan.ObjectMeta.Name, namespace, err)
			return err
		}
		logger.Debugf("%s installPlan approved in Namespace %s", installPlan.ObjectMeta.Name, namespace)
	}
	return nil
}
