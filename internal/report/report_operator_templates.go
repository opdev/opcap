package report

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
	operatorJsonReportTemplate = `{"level":"info","message":"{{ if .CsvTimeout }}timeout{{ else }}{{ .Csv.Status.Phase }}{{ end }}","package":"{{ .Subscription.Package }}","channel":"{{ .Subscription.Channel }}","installmode":"{{ .Subscription.InstallModeType }}"}{{"\n"}}`
)
