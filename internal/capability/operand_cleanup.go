package capability

import (
	"context"
	"fmt"
	"time"

	"github.com/opdev/opcap/internal/logger"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// OperandCleanup removes the operand from the OCP cluster in the ca.namespace
func operandCleanup(ctx context.Context, options auditOptions) auditCleanupFn {
	return func(ctx context.Context) error {
		logger.Debugw("cleaningUp operand for operator", "package", options.subscription.Package, "channel", options.subscription.Channel, "installmode",
			options.subscription.InstallModeType)

		if len(options.customResources) > 0 {
			for _, cr := range options.customResources {
				obj := &unstructured.Unstructured{Object: cr}

				// extract name from CustomResource object and delete it
				name := obj.Object["metadata"].(map[string]interface{})["name"].(string)

				// check if CR exists, only then cleanup the operand
				err := options.client.GetUnstructured(ctx, options.namespace, name, obj)
				if err != nil && !apierrors.IsNotFound(err) {
					// Actual error. Return it
					return fmt.Errorf("could not get operand: %v", err)
				}
				if obj == nil || apierrors.IsNotFound(err) {
					// Did not find it. Somehow already gone.
					// Not an error condition, but no point in
					// continuing.
					continue
				}

				// delete the resource using the dynamic client
				if err := options.client.DeleteUnstructured(ctx, obj); err != nil {
					logger.Debugf("failed operandCleanUp: package: %s error: %s\n", options.subscription.Package, err.Error())
					continue
				}
				time.Sleep(3 * time.Second)
				// Forcing cleanup of finalizers
				err = options.client.GetUnstructured(ctx, options.namespace, name, obj)
				if apierrors.IsNotFound(err) {
					continue
				}
				if obj.GetFinalizers() != nil {
					obj.SetFinalizers([]string{})

					if err := options.client.UpdateUnstructured(ctx, obj); err != nil {
						logger.Debugf("Couldn't update finalizer: %s", err.Error())
						continue
					}
				}
				return nil
			}
		}

		return nil
	}
}
