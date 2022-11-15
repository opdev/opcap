package report

const (
	operandTextReportTemplate = `
{{ with $dot := . }}
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

	operandJsonReportTemplate = `{{with $dot := .}}{{range $index, $value := .CustomResources }}{"package":"{{ $dot.Subscription.Package }}","Operand Kind":"{{ kind $value }}","Operand Name":"{{ name $value }}","message":"{{ if gt $dot.OperandCount 0 }}created{{ else }}failed{{ end }}"}{{ end }}{{ end }}`
)
