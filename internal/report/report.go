package report

import (
	"fmt"
	"os"
	"time"

	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
)

// operatorInstallReport holds all data for all operator install
// reports, options will dictate where to send that data

// TODO: include other relevant pieces of data such as
// Csv spec and events, operator pod and deployment status
// and events as fields of operatorInstallReport for other
// complex reports
type operatorInstallReport struct {
	ocpVersion    string
	packageName   string
	channel       string
	catalogSource string
	installMode   string
	reportDate    time.Time
	csvStatus     operatorv1alpha1.ClusterServiceVersionStatus
	reportOpts    []OperatorInstallReportOption
}

func (r *operatorInstallReport) Report() error {

	for _, opt := range r.reportOpts {

		err := opt.report(r)
		if err != nil {
			return err
		}
	}

	return nil
}

func NewOperatorInstallReport() *operatorInstallReport {
	return &operatorInstallReport{}
}

func (o *operatorInstallReport) Init(ocpVersion string, packageName string, channel string, catalogSource string, InstallMode string, csvStatus operatorv1alpha1.ClusterServiceVersionStatus, opts ...OperatorInstallReportOption) operatorInstallReport {

	opInstallRpt := operatorInstallReport{
		ocpVersion:    ocpVersion,
		packageName:   packageName,
		channel:       channel,
		catalogSource: catalogSource,
		installMode:   InstallMode,
		reportDate:    time.Now(),
		csvStatus:     csvStatus,
		reportOpts:    opts,
	}

	return opInstallRpt
}

// Report option is a family of reporting strategies
// This is specific to opertator install reports
// Ex: print to screen, write to file, write to bucket etc.
type OperatorInstallReportOption interface {

	// report here is supposed to be called by the Report object
	// therefore it's a private method and shouldn't be called
	// outside the reporting package
	report(r *operatorInstallReport) error
}

// Simple Print option implementation
type OpInstallRptOptPrint struct{}

func (opt OpInstallRptOptPrint) report(r *operatorInstallReport) error {

	fmt.Println()
	fmt.Println("opcap report:")
	fmt.Println("-----------------------------------------")
	fmt.Printf("Report Date: %s\n", r.reportDate)
	fmt.Printf("OpenShift Version: %s\n", r.ocpVersion)
	fmt.Printf("Package Name: %s\n", r.packageName)
	fmt.Printf("Channel: %s\n", r.channel)
	fmt.Printf("Catalog Source: %s\n", r.catalogSource)
	fmt.Printf("Install Mode: %s\n", r.installMode)
	fmt.Printf("Result:: %s\n", r.csvStatus.Phase)
	fmt.Printf("Message: %s\n", r.csvStatus.Message)
	fmt.Printf("Reason: %s\n", r.csvStatus.Reason)
	fmt.Println("-----------------------------------------")

	return nil
}

// Initial quick and dirty file implementation to replace
// stdout.json
type OpInstallRptOptFile struct {
	FilePath string
}

func (opt OpInstallRptOptFile) report(r *operatorInstallReport) error {

	file, err := os.OpenFile(opt.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		file.Close()
		return err
	}
	defer file.Close()

	file.WriteString("{\"level\":\"info\",\"message\":\"" + string(r.csvStatus.Phase) + "\",\"package\":\"" + r.packageName + "\",\"channel\":\"" + r.channel + "\",\"installmode\":\"" + r.installMode + "\"}\n")

	return nil
}
