package report

import (
	"io"
	"strings"
	"text/template"
	"time"

	"github.com/opdev/opcap/internal/operator"
	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type TemplateData struct {
	OcpVersion      string
	Subscription    operator.SubscriptionData
	Csv             *operatorv1alpha1.ClusterServiceVersion
	CsvTimeout      bool
	CustomResources []map[string]interface{}
	OperandCount    int
	Operands        []unstructured.Unstructured
	CsvEvents       []Event
	PodEvents       []Event
	PodLogs         []PodLog
}

type Event struct {
	InvolvedObjName   string
	InvolvedObjkind   string
	CreationTimestamp metav1.Time
	Message           string
	Reason            string
}

type PodLog struct {
	PodName       string
	ContainerName string
	PodLogs       string
}

func replace(input, from, to string) string {
	return strings.Replace(input, from, to, -1)
}

func processTemplate(w io.Writer, tmpl string, data interface{}) error {
	report, err := template.New("report").
		Funcs(template.FuncMap{
			"now":     time.Now,
			"kind":    unstructuredKind,
			"name":    unstructuredName,
			"replace": replace,
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

func unstructuredKind(cr map[string]interface{}) string {
	operand := &unstructured.Unstructured{Object: cr}
	return operand.GetKind()
}

func unstructuredName(cr map[string]interface{}) string {
	operand := &unstructured.Unstructured{Object: cr}
	return operand.GetName()
}

func OperatorInstallJsonReport(w io.Writer, data TemplateData) error {
	return processTemplate(w, operatorJsonReportTemplate, data)
}

func OperatorInstallTextReport(w io.Writer, data TemplateData) error {
	return processTemplate(w, operatorTextReportTemplate, data)
}

func OperandInstallTextReport(w io.Writer, data TemplateData) error {
	return processTemplate(w, operandTextReportTemplate, data)
}

func OperandInstallJsonReport(w io.Writer, data TemplateData) error {
	return processTemplate(w, operandJsonReportTemplate, data)
}

func DebugTextReport(w io.Writer, data TemplateData) error {
	return processTemplate(w, debugTextDataTemplate, data)
}

func DebugJsonReport(w io.Writer, data TemplateData) error {
	return processTemplate(w, debugJSONDataTemplate, data)
}
