package capability

import (
	"fmt"
	"io"

	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func (ca capAudit) OperatorInstallTextReport(w io.Writer) error {

	fmt.Fprint(w, "\n")
	fmt.Fprint(w, "Operator Install Report:\n")
	fmt.Fprint(w, "-----------------------------------------\n")
	fmt.Fprintf(w, "Report Date: %s\n", time.Now())
	fmt.Fprintf(w, "OpenShift Version: %s\n", ca.ocpVersion)
	fmt.Fprintf(w, "Package Name: %s\n", ca.subscription.Package)
	fmt.Fprintf(w, "Channel: %s\n", ca.subscription.Channel)
	fmt.Fprintf(w, "Catalog Source: %s\n", ca.subscription.CatalogSource)
	fmt.Fprintf(w, "Install Mode: %s\n", ca.subscription.InstallModeType)

	if !ca.csvTimeout {
		fmt.Fprintf(w, "Result: %s\n", ca.csv.Status.Phase)
	} else {
		fmt.Fprint(w, "Result: timeout\n")
	}

	fmt.Fprintf(w, "Message: %s\n", ca.csv.Status.Message)
	fmt.Fprintf(w, "Reason: %s\n", ca.csv.Status.Reason)
	fmt.Fprint(w, "-----------------------------------------\n")

	return nil
}

func (ca capAudit) OperatorInstallJsonReport(w io.Writer) error {

	if !ca.csvTimeout {
		fmt.Fprintf(w, "{\"level\":\"info\",\"message\":\""+string(ca.csv.Status.Phase)+"\",\"package\":\""+ca.subscription.Package+
			"\",\"channel\":\""+ca.subscription.Channel+"\",\"installmode\":\""+string(ca.subscription.InstallModeType)+"\"}\n")
	} else {
		fmt.Fprintf(w, "{\"level\":\"info\",\"message\":\""+"timeout"+"\",\"package\":\""+ca.subscription.Package+"\",\"channel\":\""+
			ca.subscription.Channel+"\",\"installmode\":\""+string(ca.subscription.InstallModeType)+"\"}\n")
	}

	return nil
}

func (ca capAudit) OperandTextReport(w io.Writer) error {

	for _, cr := range ca.customResources {
		operand := &unstructured.Unstructured{Object: cr}

		fmt.Fprint(w, "\n")
		fmt.Fprintf(w, "Operand Install Report:\n")
		fmt.Fprintf(w, "-----------------------------------------\n")
		fmt.Fprintf(w, "Report Date: %s\n", time.Now())
		fmt.Fprintf(w, "OpenShift Version: %s\n", ca.ocpVersion)
		fmt.Fprintf(w, "Package Name: %s\n", ca.subscription.Package)
		fmt.Fprintf(w, "Operand Kind: %s\n", operand.GetKind())
		fmt.Fprintf(w, "Operand Name: %s\n", operand.GetName())

		if len(ca.operands) > 0 {
			fmt.Fprint(w, "Operand Creation: Succeeded\n")
		} else {
			fmt.Fprint(w, "Operand Creation: Failed\n")
		}
		fmt.Fprint(w, "-----------------------------------------\n")
	}
	return nil
}

func (ca capAudit) OperandInstallJsonReport(w io.Writer) error {

	for _, cr := range ca.customResources {
		operand := &unstructured.Unstructured{Object: cr}

		if len(ca.operands) > 0 {

			fmt.Fprintf(w, "{\"package\":\""+ca.subscription.Package+"\", \"Operand Kind\": \""+operand.GetKind()+"\", \"Operand Name\": \""+operand.GetName()+
				"\",\"message\":\""+"created"+"\"}\n")
		} else {

			fmt.Fprintf(w, "{\"package\":\""+ca.subscription.Package+"\", \"Operand Kind\": \""+operand.GetKind()+"\", \"Operand Name\": \""+operand.GetName()+
				"\",\"message\":\""+"failed"+"\"}\n")
		}
	}

	return nil
}
