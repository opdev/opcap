// Copyright 2021 The Audit Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package capabilities

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"capabilities-tool/pkg"
	"capabilities-tool/pkg/actions"
	"capabilities-tool/pkg/models"
	index "capabilities-tool/pkg/reports/capabilities"

	_ "github.com/mattn/go-sqlite3"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var flags = index.BindFlags{}

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "capabilities",
		Short:   "",
		Long:    "",
		PreRunE: validation,
		RunE:    run,
	}

	currentPath, err := os.Getwd()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	cmd.Flags().StringVar(&flags.IndexImage, "index-image", "",
		"index image and tag which will be audit")
	if err := cmd.MarkFlagRequired("index-image"); err != nil {
		log.Fatalf("Failed to mark `index-image` flag for `index` sub-command as required")
	}

	cmd.Flags().StringVar(&flags.Filter, "filter", "",
		"filter by the packages names which are like *filter*")
	cmd.Flags().StringVar(&flags.FilterBundle, "filter-bundle", "",
		"filter by the Bundle names which are like *filter-bundle*")
	cmd.Flags().StringVar(&flags.OutputFormat, "output", pkg.JSON,
		fmt.Sprintf("inform the output format. [Options: %s]", pkg.JSON))
	cmd.Flags().StringVar(&flags.OutputPath, "output-path", currentPath,
		"inform the path of the directory to output the report. (Default: current directory)")
	cmd.Flags().Int32Var(&flags.Limit, "limit", 0,
		"limit the num of operator bundles to be audit")
	cmd.Flags().StringVar(&flags.S3Bucket, "s3-bucket", "", "")
	cmd.Flags().StringVar(&flags.Endpoint, "endpoint", "http://operator-audit-minio.apps.eng.opdev.io", "")
	cmd.Flags().BoolVar(&flags.HeadOnly, "head-only", false,
		"if set, will just check the operator bundle which are head of the channels")
	cmd.Flags().StringVar(&flags.ContainerEngine, "container-engine", pkg.Docker,
		fmt.Sprintf("specifies the container tool to use. If not set, the default value is docker. "+
			"Note that you can use the environment variable CONTAINER_ENGINE to inform this option. "+
			"[Options: %s and %s]", pkg.Docker, pkg.Podman))
	cmd.Flags().StringVar(&flags.PullSecretName, "pull-secret-name", "registry-pull-secret", "")
	cmd.Flags().StringVar(&flags.ServiceAccount, "service-account", "default", "")

	return cmd
}

func validation(cmd *cobra.Command, args []string) error {

	if flags.Limit < 0 {
		return fmt.Errorf("invalid value informed via the --limit flag :%v", flags.Limit)
	}

	if len(flags.OutputFormat) > 0 && flags.OutputFormat != pkg.JSON {
		return fmt.Errorf("invalid value informed via the --output flag :%v. "+
			"The available option is: %s", flags.OutputFormat, pkg.JSON)
	}

	if len(flags.OutputPath) > 0 {
		if _, err := os.Stat(flags.OutputPath); os.IsNotExist(err) {
			return err
		}
	}

	if len(flags.ContainerEngine) == 0 {
		flags.ContainerEngine = pkg.GetContainerToolFromEnvVar()
	}
	if flags.ContainerEngine != pkg.Docker && flags.ContainerEngine != pkg.Podman {
		return fmt.Errorf("invalid value for the flag --container-engine (%s)."+
			" The valid options are %s and %s", flags.ContainerEngine, pkg.Docker, pkg.Podman)
	}

	return nil
}

func run(cmd *cobra.Command, args []string) error {
	log.Info("Running capabilities run function")

	reportData := index.Data{}
	reportData.Flags = flags
	pkg.GenerateTemporaryDirs()

	reportData.Flags.Filter = strings.ReplaceAll(reportData.Flags.Filter, "”", "")

	if err := actions.DownloadImage(flags.IndexImage, flags.ContainerEngine); err != nil {
		return err
	}

	if err := actions.ExtractIndexDB(flags.IndexImage, flags.ContainerEngine); err != nil {
		return err
	}

	report, err := getDataFromIndexDB(reportData)
	if err != nil {
		log.Errorf("Unable to get data from index db: %v\n", err)
	}

	log.Info("Deploying operator with operator-sdk...")
	for idx, bundle := range report.AuditCapabilities {
		operatorsdk := exec.Command("operator-sdk", "run", "bundle", bundle.OperatorBundleImagePath, "--pull-secret-name", flags.PullSecretName, "--timeout", "5m")
		runCommand, err := pkg.RunCommand(operatorsdk)

		if err != nil {
			log.Errorf("Unable to run operator-sdk run bundle: %v\n", err)
		}

		RBLogs := string(runCommand[:])
		report.AuditCapabilities[idx].InstallLogs = append(report.AuditCapabilities[idx].InstallLogs, RBLogs)
		report.AuditCapabilities[idx].Capabilities = false

		if strings.Contains(RBLogs, "OLM has successfully installed") {
			log.Info("Operator Installed Successfully")
			report.AuditCapabilities[idx].Capabilities = true
		}

		log.Info("Cleaning up installed Operator:", bundle.PackageName)
		cleanup := exec.Command("operator-sdk", "cleanup", bundle.PackageName)
		runCleanup, err := pkg.RunCommand(cleanup)
		if err != nil {
			log.Errorf("Unable to run operator-sdk cleanup: %v\n", err)
		}
		CLogs := string(runCleanup)
		report.AuditCapabilities[idx].CleanUpLogs = append(report.AuditCapabilities[idx].CleanUpLogs, CLogs)
	}

	log.Info("Generating output...")
	if err := report.OutputReport(); err != nil {
		return err
	}

	const reportType = "capabilities"
	imageName := report.Flags.FilterBundle
	outputPath := report.Flags.OutputPath
	filename := pkg.GetReportName(imageName, reportType, "json")
	path := filepath.Join(outputPath, filename)
	log.Info(path)
	log.Info("Uploading result to S3")
	pkg.WriteDataToS3(path, filename, flags.S3Bucket, flags.Endpoint)

	log.Info("Task Completed!!!!!")

	return nil
}

func getDataFromIndexDB(report index.Data) (index.Data, error) {
	// Connect to the database
	db, err := sql.Open("sqlite3", "./output/index.db")
	if err != nil {
		return report, fmt.Errorf("unable to connect in to the database : %s", err)
	}

	sql, err := report.BuildCapabilitiesQuery()
	if err != nil {
		return report, err
	}

	row, err := db.Query(sql)
	if err != nil {
		return report, fmt.Errorf("unable to query the index db : %s", err)
	}

	defer row.Close()
	for row.Next() {
		var bundleName string
		var bundlePath string

		err = row.Scan(&bundleName, &bundlePath)
		if err != nil {
			log.Errorf("unable to scan data from index %s\n", err.Error())
		}
		log.Infof("Generating data from the bundle (%s)", bundleName)
		auditCapabilities := models.NewAuditCapabilities(bundleName, bundlePath)

		sqlString := fmt.Sprintf("SELECT c.channel_name, c.package_name FROM channel_entry c "+
			"where c.operatorbundle_name = '%s'", auditCapabilities.OperatorBundleName)
		row, err := db.Query(sqlString)
		if err != nil {
			return report, fmt.Errorf("unable to query channel entry in the index db : %s", err)
		}

		defer row.Close()
		var channelName string
		var packageName string
		for row.Next() { // Iterate and fetch the records from result cursor
			_ = row.Scan(&channelName, &packageName)
			auditCapabilities.Channels = append(auditCapabilities.Channels, channelName)
			auditCapabilities.PackageName = packageName
		}

		if len(strings.TrimSpace(auditCapabilities.PackageName)) == 0 && auditCapabilities.Bundle != nil {
			auditCapabilities.PackageName = auditCapabilities.Bundle.Package
		}

		sqlString = fmt.Sprintf("SELECT default_channel FROM package WHERE name = '%s'", auditCapabilities.PackageName)
		row, err = db.Query(sqlString)
		if err != nil {
			return report, fmt.Errorf("unable to query default channel entry in the index db : %s", err)
		}

		defer row.Close()
		var defaultChannelName string
		for row.Next() { // Iterate and fetch the records from result cursor
			_ = row.Scan(&defaultChannelName)
			auditCapabilities.DefaultChannel = defaultChannelName
		}

		sqlString = fmt.Sprintf("SELECT type, value FROM properties WHERE operatorbundle_name = '%s'",
			auditCapabilities.OperatorBundleName)
		row, err = db.Query(sqlString)
		if err != nil {
			return report, fmt.Errorf("unable to query properties entry in the index db : %s", err)
		}

		defer row.Close()
		var properType string
		var properValue string
		for row.Next() { // Iterate and fetch the records from result cursor
			_ = row.Scan(&properType, &properValue)
			auditCapabilities.PropertiesDB = append(auditCapabilities.PropertiesDB,
				pkg.PropertiesAnnotation{Type: properType, Value: properValue})
		}

		sqlString = fmt.Sprintf("select count(*) from channel where head_operatorbundle_name = '%s'",
			auditCapabilities.OperatorBundleName)
		row, err = db.Query(sqlString)
		if err != nil {
			return report, fmt.Errorf("unable to query properties entry in the index db : %s", err)
		}

		defer row.Close()
		var found int
		for row.Next() { // Iterate and fetch the records from result cursor
			_ = row.Scan(&found)
			auditCapabilities.IsHeadOfChannel = found > 0
		}

		report.AuditCapabilities = append(report.AuditCapabilities, *auditCapabilities)
	}

	return report, nil
}
