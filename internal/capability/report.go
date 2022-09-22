package capability

import (
	"fmt"
	"io"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func operatorInstallTextReport(w io.Writer, ca options) error {
	fmt.Fprint(w, "\n")
	fmt.Fprint(w, "Operator Install Report:\n")
	fmt.Fprint(w, "-----------------------------------------\n")
	fmt.Fprintf(w, "Report Date: %s\n", time.Now())
	fmt.Fprintf(w, "OpenShift Version: %s\n", ca.OcpVersion)
	fmt.Fprintf(w, "Package Name: %s\n", ca.Subscription.Package)
	fmt.Fprintf(w, "Channel: %s\n", ca.Subscription.Channel)
	fmt.Fprintf(w, "Catalog Source: %s\n", ca.Subscription.CatalogSource)
	fmt.Fprintf(w, "Install Mode: %s\n", ca.Subscription.InstallModeType)

	if !ca.CsvTimeout {
		fmt.Fprintf(w, "Result: %s\n", ca.Csv.Status.Phase)
	} else {
		fmt.Fprint(w, "Result: timeout\n")
	}

	fmt.Fprintf(w, "Message: %s\n", ca.Csv.Status.Message)
	fmt.Fprintf(w, "Reason: %s\n", ca.Csv.Status.Reason)
	fmt.Fprint(w, "-----------------------------------------\n")

	return nil
}

func operatorInstallJsonReport(w io.Writer, ca options) error {
	if !ca.CsvTimeout {
		fmt.Fprintf(w, "{\"level\":\"info\",\"message\":\""+string(ca.Csv.Status.Phase)+"\",\"package\":\""+ca.Subscription.Package+
			"\",\"channel\":\""+ca.Subscription.Channel+"\",\"installmode\":\""+string(ca.Subscription.InstallModeType)+"\"}\n")
	} else {
		fmt.Fprintf(w, "{\"level\":\"info\",\"message\":\""+"timeout"+"\",\"package\":\""+ca.Subscription.Package+"\",\"channel\":\""+
			ca.Subscription.Channel+"\",\"installmode\":\""+string(ca.Subscription.InstallModeType)+"\"}\n")
	}

	return nil
}

func operandTextReport(w io.Writer, ca options) error {
	for _, cr := range ca.customResources {
		operand := &unstructured.Unstructured{Object: cr}

		fmt.Fprint(w, "\n")
		fmt.Fprintf(w, "Operand Install Report:\n")
		fmt.Fprintf(w, "-----------------------------------------\n")
		fmt.Fprintf(w, "Report Date: %s\n", time.Now())
		fmt.Fprintf(w, "OpenShift Version: %s\n", ca.OcpVersion)
		fmt.Fprintf(w, "Package Name: %s\n", ca.Subscription.Package)
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

func operandInstallJsonReport(w io.Writer, ca options) error {
	for _, cr := range ca.customResources {
		operand := &unstructured.Unstructured{Object: cr}

		if len(ca.operands) > 0 {
			fmt.Fprintf(w, "{\"package\":\""+ca.Subscription.Package+"\", \"Operand Kind\": \""+operand.GetKind()+"\", \"Operand Name\": \""+operand.GetName()+
				"\",\"message\":\""+"created"+"\"}\n")
		} else {
			fmt.Fprintf(w, "{\"package\":\""+ca.Subscription.Package+"\", \"Operand Kind\": \""+operand.GetKind()+"\", \"Operand Name\": \""+operand.GetName()+
				"\",\"message\":\""+"failed"+"\"}\n")
		}
	}

	return nil
}
