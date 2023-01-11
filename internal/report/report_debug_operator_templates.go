package report

const (
	debugTextDataTemplate = `
Debug Operator Install Report
-----------------------------------------
Report Date: {{ now }}
OpenShift Version: {{ .OcpVersion }}
Package Name: {{ .Subscription.Package }}
Channel: {{ .Subscription.Channel }}
Catalog Source: {{ .Subscription.CatalogSource }}
Install Mode: {{ .Subscription.InstallModeType }}
Timeout: {{ .CsvTimeout }}
Phase: {{ .Csv.Status.Phase }}
Message: {{ .Csv.Status.Message }}
Reason: {{ .Csv.Status.Reason }}
Conditions: {{ .Csv.Status.Conditions }}
RequirementStatus: {{ .Csv.Status.RequirementStatus }}
-----------------------------------------
`
	debugJSONDataTemplate = `{"Package Name":"{{ .Subscription.Package }}",
"Channel":"{{ .Subscription.Channel }}",
"Catalog Source":"{{ .Subscription.CatalogSource }}",
"Install Mode":"{{ .Subscription.InstallModeType }}",
"Timeout":"{{ .CsvTimeout }}",
{{ if .Csv}}"Phase":"{{ .Csv.Status.Phase }}",
"Reason":"{{ .Csv.Status.Reason }}",
"Conditions":"{{ .Csv.Status.Phase }}",
"CSV Events": "{{ .CsvEvents }}",
"Pod Events": "{{ .PodEvents }}",
"Pod Logs": "{{ range .PodLogs }}{PodName: {{ .PodName }}, ContainerName: {{.ContainerName}}, PodLog: {{ .PodLogs }}}{{ end }}"}{{"\n"}}{{ end }}}`
)
