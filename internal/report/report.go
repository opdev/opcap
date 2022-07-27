package report

import (
	"fmt"
	"time"

	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
)

type Report interface {
	GenerateReport() error
	// PrintToFile()
	// WriteToDatabase()
	// WriteToMinioBucket()
}

// CSV status
type operatorInstallReport struct {
	ocpVersion    string
	packageName   string
	channel       string
	catalogSource string
	installMode   string
	reportDate    time.Time
	csvStatus     operatorv1alpha1.ClusterServiceVersionStatus
	reportOpts    ReportOptions
	// CsvEvents
	// PodStatus
	// PodEvents
	// PodLogs
	// DeploymentStatus
	// DeploymentEvents
}

type ReportOption string

const (
	ReportOptPrint ReportOption = "Print"
	// ReportOptWriteToBucket
	// RerportOptWriteToDatabase
)

type ReportOptions []ReportOption

func NewOperatorInstallReport(ocpVersion string, packageName string, channel string, catalogSource string, InstallMode string, csvStatus operatorv1alpha1.ClusterServiceVersionStatus, opts ReportOptions) Report {

	return operatorInstallReport{
		ocpVersion:    ocpVersion,
		packageName:   packageName,
		channel:       channel,
		catalogSource: catalogSource,
		installMode:   InstallMode,
		reportDate:    time.Now(),
		csvStatus:     csvStatus,
		reportOpts:    opts,
	}
}

func (r operatorInstallReport) GenerateReport() error {

	for _, opt := range r.reportOpts {
		switch opt {
		case ReportOptPrint:
			r.print()

		}
	}
	return nil
}

func (r operatorInstallReport) print() {
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
}
