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

package one

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gobuffalo/envy"

	"opcap/pkg"
	"opcap/pkg/models"
	index "opcap/pkg/reports/capabilities"

	_ "github.com/mattn/go-sqlite3"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var flags = index.BindFlags{}

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "one",
		Short:   "Checks for Operator Capability level 1, i.e Basic Install",
		Long:    "",
		PreRunE: validation,
		RunE:    run,
	}

	currentPath, err := os.Getwd()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	cmd.Flags().StringVar(&flags.PackageName, "package-name", "",
		"filter by the Package names which are like *package-name*. Required for Operator Clean-up")
	cmd.Flags().StringVar(&flags.BundleName, "bundle-name", "",
		"filter by the Bundle names which are like *bundle-name*")
	cmd.Flags().StringVar(&flags.FilterBundle, "bundle-image", "",
		"filter by the Bundle names which are like *bundle-image*")
	cmd.Flags().StringVar(&flags.OutputFormat, "output", pkg.JSON,
		fmt.Sprintf("inform the output format. [Options: %s]", pkg.JSON))
	cmd.Flags().StringVar(&flags.OutputPath, "output-path", currentPath,
		"inform the path of the directory to output the report. (Default: current directory)")
	cmd.Flags().StringVar(&flags.S3Bucket, "bucket-name", "",
		"minio bucket name where result will be stored")
	cmd.Flags().StringVar(&flags.Endpoint, "endpoint", envy.Get("MINIO_ENDPOINT", ""), ""+
		"minio endpoint where bucket will be created")
	cmd.Flags().StringVar(&flags.ContainerEngine, "container-engine", pkg.Docker,
		fmt.Sprintf("specifies the container tool to use. If not set, the default value is docker. "+
			"Note that you can use the environment variable CONTAINER_ENGINE to inform this option. "+
			"[Options: %s and %s]", pkg.Docker, pkg.Podman))
	cmd.Flags().StringVar(&flags.PullSecretName, "pull-secret-name", "registry-pull-secret",
		"Name of Kubernetes Secret to use for pulling registry images")
	cmd.Flags().StringVar(&flags.ServiceAccount, "service-account", "default",
		"Name of Kubernetes Service Account to use")
	cmd.Flags().StringVar(&flags.InstallMode, "install-mode", "AllNamespaces",
		"Install mode for installing the operator. Accepts following strings as input `MultiNamespace=ns1,ns2 | AllNamespace | OwnNamespace | SingleNamespace=ns1`")

	return cmd
}

func validation(cmd *cobra.Command, args []string) error {

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
	log.Info("Running operator capabilities level 1 checks")

	reportData := index.Data{}
	reportData.Flags = flags
	pkg.GenerateTemporaryDirs()

	var Bundle models.AuditCapabilities

	log.Info("Deploying operator with operator-sdk...")
	operatorsdk := exec.Command("operator-sdk", "run", "bundle", flags.FilterBundle, "--pull-secret-name", flags.PullSecretName, "--timeout", "5m", "--install-mode", flags.InstallMode)
	runCommand, err := pkg.RunCommand(operatorsdk)

	if err != nil {
		log.Errorf("Unable to run operator-sdk run bundle: %v\n", err)
	}

	RBLogs := string(runCommand[:])
	Bundle.InstallLogs = append(Bundle.InstallLogs, RBLogs)
	Bundle.OperatorBundleImagePath = flags.FilterBundle
	Bundle.OperatorBundleName = flags.BundleName

	reportData.AuditCapabilities = append(reportData.AuditCapabilities, Bundle)
	reportData.AuditCapabilities[0].Capabilities = false

	if strings.Contains(RBLogs, "OLM has successfully installed") {
		log.Info("Operator Installed Successfully")
		reportData.AuditCapabilities[0].Capabilities = true
	}

	if flags.PackageName != "" {
		log.Info("Cleaning up installed Operator:", flags.PackageName)
		Bundle.PackageName = flags.PackageName
		cleanup := exec.Command("operator-sdk", "cleanup", flags.PackageName)
		runCleanup, err := pkg.RunCommand(cleanup)
		if err != nil {
			log.Errorf("Unable to run operator-sdk cleanup: %v\n", err)
		}
		CLogs := string(runCleanup)
		reportData.AuditCapabilities[0].CleanUpLogs = append(reportData.AuditCapabilities[0].CleanUpLogs, CLogs)
	}

	log.Info("Generating output...")
	if err := reportData.OutputReport(); err != nil {
		return err
	}

	log.Info("Uploading result to S3")
	filename := pkg.GetReportName(reportData.Flags.BundleName, "cap_level_1", "json")
	path := filepath.Join(reportData.Flags.OutputPath, filename)
	if err := pkg.WriteDataToS3(path, filename, flags.S3Bucket, flags.Endpoint); err != nil {
		return err
	}

	log.Info("Task Completed!!!!!")

	return nil
}
