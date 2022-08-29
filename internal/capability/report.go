package capability

import (
	"fmt"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func (ca capAudit) Report(opts ...ReportOption) error {

	for _, opt := range opts {

		err := opt.report(ca)
		if err != nil {
			logger.Debugf("Unable to generate report for %T", opt, "Error: %s", err)
		}

	}
	return nil
}

type ReportOption interface {
	report(ca capAudit) error
}

// Simple print option implmentation for operator install
type OperatorInstallRptOptionPrint struct{}

func (OperatorInstallRptOptionPrint) report(ca capAudit) error {

	fmt.Println()
	fmt.Println("Operator Install Report:")
	fmt.Println("-----------------------------------------")
	fmt.Printf("Report Date: %s\n", time.Now())
	fmt.Printf("OpenShift Version: %s\n", ca.ocpVersion)
	fmt.Printf("Package Name: %s\n", ca.subscription.Package)
	fmt.Printf("Channel: %s\n", ca.subscription.Channel)
	fmt.Printf("Catalog Source: %s\n", ca.subscription.CatalogSource)
	fmt.Printf("Install Mode: %s\n", ca.subscription.InstallModeType)

	if !ca.csvTimeout {
		fmt.Printf("Result: %s\n", ca.csv.Status.Phase)
	} else {
		fmt.Println("Result: timeout")
	}

	fmt.Printf("Message: %s\n", ca.csv.Status.Message)
	fmt.Printf("Reason: %s\n", ca.csv.Status.Reason)
	fmt.Println("-----------------------------------------")

	return nil
}

// Simple file option implementation for operator install
type OperatorInstallRptOptionFile struct {
	FilePath string
}

func (opt OperatorInstallRptOptionFile) report(ca capAudit) error {

	file, err := os.OpenFile(opt.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		file.Close()
		return err
	}
	defer file.Close()

	if !ca.csvTimeout {

		file.WriteString("{\"level\":\"info\",\"message\":\"" + string(ca.csv.Status.Phase) + "\",\"package\":\"" + ca.subscription.Package + "\",\"channel\":\"" + ca.subscription.Channel + "\",\"installmode\":\"" + string(ca.subscription.InstallModeType) + "\"}\n")
	} else {

		file.WriteString("{\"level\":\"info\",\"message\":\"" + "timeout" + "\",\"package\":\"" + ca.subscription.Package + "\",\"channel\":\"" + ca.subscription.Channel + "\",\"installmode\":\"" + string(ca.subscription.InstallModeType) + "\"}\n")
	}

	return nil
}

// Simple print option implmentation for operand install
type OperandInstallRptOptionPrint struct{}

func (OperandInstallRptOptionPrint) report(ca capAudit) error {

	for _, cr := range ca.customResources {
		operand := &unstructured.Unstructured{Object: cr}

		fmt.Println()
		fmt.Println("Operand Install Report:")
		fmt.Println("-----------------------------------------")
		fmt.Printf("Report Date: %s\n", time.Now())
		fmt.Printf("OpenShift Version: %s\n", ca.ocpVersion)
		fmt.Printf("Package Name: %s\n", ca.subscription.Package)
		fmt.Printf("Operand Kind: %s\n", operand.GetKind())
		fmt.Printf("Operand Name: %s\n", operand.GetName())

		if len(ca.operands) > 0 {
			fmt.Println("Operand Creation: Succeeded")
		} else {
			fmt.Println("Operand Creation: Failed")
		}
		fmt.Println("-----------------------------------------")
	}
	return nil
}

// Simple file option implementation for operand install
type OperandInstallRptOptionFile struct {
	FilePath string
}

func (opt OperandInstallRptOptionFile) report(ca capAudit) error {

	for _, cr := range ca.customResources {
		operand := &unstructured.Unstructured{Object: cr}

		file, err := os.OpenFile(opt.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			file.Close()
			return err
		}
		defer file.Close()

		if len(ca.operands) > 0 {

			file.WriteString("{\"package\":\"" + ca.subscription.Package + "\", \"Operand Kind\": \"" + operand.GetKind() + "\", \"Operand Name\": \"" + operand.GetName() + "\",\"message\":\"" + "created" + "\"}\n")
		} else {

			file.WriteString("{\"package\":\"" + ca.subscription.Package + "\", \"Operand Kind\": \"" + operand.GetKind() + "\", \"Operand Name\": \"" + operand.GetName() + "\",\"message\":\"" + "failed" + "\"}\n")
		}
	}

	return nil
}
