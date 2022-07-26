package operator

import (
	"context"
	"fmt"
	"time"

	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
)

// ApproveInstallPlan waits and approves installPlans
func (c operatorClient) ApproveInstallPlan(ctx context.Context, sub *operatorv1alpha1.Subscription) error {
	subKey := types.NamespacedName{
		Namespace: sub.GetNamespace(),
		Name:      sub.GetName(),
	}

	// Wait for an installPlan to be associated to the subscription.
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

	namespace := sub.GetNamespace()

	// Get installPlan referenced by sub.Status.InstallPlanRef
	installPlan := operatorv1alpha1.InstallPlan{}
	installPlanKey := types.NamespacedName{Name: sub.Status.InstallPlanRef.Name, Namespace: namespace}
	err := c.Client.Get(ctx, installPlanKey, &installPlan)
	if err != nil {
		logger.Errorf("installPlan %s not found in namespace %s: %w", sub.Status.InstallPlanRef.Name, namespace, err)
		return err
	}

	// Approve installPlan if necessary
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
