package capability

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/opdev/opcap/internal/logger"
	"github.com/opdev/opcap/internal/report"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func extractAlmExamples(ctx context.Context, options *options) error {
	// gets the list of CSVs present in a particular namespace
	csvList, err := options.client.ListClusterServiceVersions(ctx, options.namespace)
	if err != nil {
		return err
	}
	almExamples := ""
	for _, csvVal := range csvList.Items {
		if strings.HasPrefix(csvVal.ObjectMeta.Name, options.operatorGroupData.Name) {
			// map of string interface which consist of ALM examples from the CSVList
			almExamples = csvVal.ObjectMeta.Annotations["alm-examples"]
		}
	}
	var almList []map[string]interface{}

	err = yaml.Unmarshal([]byte(almExamples), &almList)
	if err != nil {
		return err
	}

	options.customResources = append(options.customResources, almList...)

	return nil
}

// OperandInstall installs the operand from the ALMExamples in the ca.namespace
func operandInstall(ctx context.Context, opts ...auditOption) (auditFn, auditCleanupFn) {
	var options options
	for _, opt := range opts {
		err := opt(&options)
		if err != nil {
			return func(_ context.Context) error {
					return fmt.Errorf("option failed: %v", err)
				}, func(_ context.Context) error {
					return nil
				}
		}
	}

	return func(ctx context.Context) error {
		logger.Debugw("installing operand for operator", "package", options.subscription.Package, "channel", options.subscription.Channel, "installmode", options.subscription.InstallModeType)

		if err := extractAlmExamples(ctx, &options); err != nil {
			logger.Errorf("could not get ALM Examples: %v", err)
		}

		if len(options.customResources) == 0 {
			logger.Debugf("exiting OperandInstall since no ALM_Examples found in CSV")
			return nil
		}

		csv, err := options.client.GetCompletedCsvWithTimeout(ctx, options.namespace, time.Minute)
		if err != nil {
			return fmt.Errorf("could not get CSV: %v", err)
		}
		options.csv = csv

		if strings.ToLower(string(csv.Status.Phase)) != "succeeded" {
			return fmt.Errorf("exiting OperandInstall since CSV install has failed")
		}

		for _, cr := range options.customResources {
			obj := &unstructured.Unstructured{Object: cr}

			// set the namespace of CR to the namespace of the subscription
			obj.SetNamespace(options.namespace)

			// create the resource using the dynamic client and log the error if it occurs
			err := options.client.CreateUnstructured(ctx, obj)
			if err != nil {
				// If there is an error, log and continue
				logger.Errorw("could not create resource", "error", err, "namespace", options.namespace)
				continue
			}
			options.operands = append(options.operands, *obj)
		}

		file, err := os.OpenFile("operand_install_report.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		defer file.Close()

		err = report.OperandInstallJsonReport(file, report.TemplateData{
			CustomResources: options.customResources,
			OcpVersion:      options.ocpVersion,
			Subscription:    *options.subscription,
			Csv:             options.csv,
			OperandCount:    len(options.operands),
		})
		if err != nil {
			return fmt.Errorf("could not generate operand install JSON report: %v", err)
		}

		err = report.OperandInstallTextReport(os.Stdout, report.TemplateData{
			CustomResources: options.customResources,
			OcpVersion:      options.ocpVersion,
			Subscription:    *options.subscription,
			Csv:             options.csv,
			OperandCount:    len(options.operands),
		})
		if err != nil {
			return fmt.Errorf("could not generate operand install text report: %v", err)
		}

		return nil
	}, operandCleanup(ctx, opts...)
}
