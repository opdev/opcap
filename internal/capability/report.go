package capability

import (
	"io"
	"text/template"
	"time"

	"github.com/opdev/opcap/internal/operator"
	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	operatorTextReportTemplate = `
Operator Install Report
-----------------------------------------
Report Date: {{ now }}
OpenShift Version: {{ .OcpVersion }}
Package Name: {{ .Subscription.Package }}
Channel: {{ .Subscription.Channel }}
Catalog Source: {{ .Subscription.CatalogSource }}
Install Mode: {{ .Subscription.InstallModeType }}
Result: {{ if .CsvTimeout }}timeout{{ else }}{{ .Csv.Status.Phase }}
Message: {{ .Csv.Status.Message }}
Reason: {{ .Csv.Status.Reason }}
{{ end }}
-----------------------------------------
`
	operatorJsonReportTemplate = `{"level":"info","message":"{{ if .CsvTimeout }}timeout{{ else }}{{ .Csv.Status.Phase }}{{ end }}","package":"{{ .Subscription.Package }}","channel":"{{ .Subscription.Channel }}","installmode":"{{ .Subscription.InstallModeType }}"}`

	operandTextReportTemplate = `{{ with $dot := . }}
{{ range $index, $value := .CustomResources }}

Operand Install Report
-----------------------------------------
Report Date: {{ now }}
OpenShift Version: {{ $dot.OcpVersion }}
Package Name: {{ $dot.Subscription.Package }}
Operand Kind: {{ kind $value }}
Operand Name: {{ name $value }}
Operand Creation: {{ if gt $dot.OperandCount 0 }}Succeeded{{ else }}Failed{{ end }}
-----------------------------------------
{{ else }}
No custom resources
{{ end }}
{{ end }}
`

	operandJsonReportTemplate = `{{with $dot := .}}{{range $index, $value := .CustomResources }}
{"package":"{{ $dot.Subscription.Package }}","Operand Kind":"{{ kind $value }}","Operand Name":"{{ name $value }}","message":"{{ if gt $dot.OperandCount 0 }}created{{ else }}failed{{ end }}"}
{{ end }}{{ end }}`
)

type operatorTemplateData struct {
	OcpVersion   string
	Subscription operator.SubscriptionData
	Csv          *operatorv1alpha1.ClusterServiceVersion
	CsvTimeout   bool
}

type operandTemplateData struct {
	CustomResources []map[string]interface{}
	OcpVersion      string
	Subscription    operator.SubscriptionData
	Csv             *operatorv1alpha1.ClusterServiceVersion
	OperandCount    int
}

func processTemplate(w io.Writer, tmpl string, data interface{}) error {
	report, err := template.New("report").
		Funcs(template.FuncMap{
			"now":  time.Now,
			"kind": unstructuredKind,
			"name": unstructuredName,
		}).
		Parse(tmpl)
	if err != nil {
		return err
	}
	if err := report.Execute(w, data); err != nil {
		return err
	}
	return nil
}

func operatorInstallJsonReport(w io.Writer, ca options) error {
	return processTemplate(w, operatorJsonReportTemplate, operatorTemplateData{
		OcpVersion:   ca.OcpVersion,
		Subscription: *ca.Subscription,
		Csv:          ca.Csv,
		CsvTimeout:   ca.CsvTimeout,
	})
}

func operatorInstallTextReport(w io.Writer, ca options) error {
	return processTemplate(w, operatorTextReportTemplate, operatorTemplateData{
		OcpVersion:   ca.OcpVersion,
		Subscription: *ca.Subscription,
		Csv:          ca.Csv,
		CsvTimeout:   ca.CsvTimeout,
	})
}

func operandInstallTextReport(w io.Writer, ca options) error {
	return processTemplate(w, operandTextReportTemplate, operandTemplateData{
		CustomResources: ca.customResources,
		OcpVersion:      ca.OcpVersion,
		Subscription:    *ca.Subscription,
		Csv:             ca.Csv,
		OperandCount:    len(ca.operands),
	})
}

func operandInstallJsonReport(w io.Writer, ca options) error {
	return processTemplate(w, operandJsonReportTemplate, operandTemplateData{
		CustomResources: ca.customResources,
		OcpVersion:      ca.OcpVersion,
		Subscription:    *ca.Subscription,
		Csv:             ca.Csv,
		OperandCount:    len(ca.operands),
	})
}

func unstructuredKind(cr map[string]interface{}) string {
	operand := &unstructured.Unstructured{Object: cr}
	return operand.GetKind()
}

func unstructuredName(cr map[string]interface{}) string {
	operand := &unstructured.Unstructured{Object: cr}
	return operand.GetName()
}
