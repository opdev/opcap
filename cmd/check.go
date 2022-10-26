package cmd

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/opdev/opcap/internal/capability"
	"github.com/opdev/opcap/internal/operator"
	"k8s.io/client-go/rest"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

type checkCommandFlags struct {
	AuditPlan              []string `json:"auditPlan"`
	CatalogSource          string   `json:"catalogsource"`
	CatalogSourceNamespace string   `json:"catalogsourcenamespace"`
	Packages               []string `json:"packages"`
	AllInstallModes        bool     `json:"allInstallModes"`
	ExtraCRDirectory       string   `json:"extraCRDirectory"`
}

var checkflags checkCommandFlags

// TODO: provide godoc compatible comment for checkCmd
func checkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Checks if operator meets minimum capability requirement.",
		Long: `The 'check' command checks if OpenShift operators meet minimum
requirements for Operator Capabilities Level to attest operator
advanced features by running custom resources provided by CSVs
and/or users.`,
		Example: "opcap check --catalogsource=certified-operators --catalogsourcenamespace=openshift-marketplace",
		RunE:    checkRunE,
	}

	defaultAuditPlan := []string{"OperatorInstall"}

	flags := cmd.Flags()

	flags.StringVar(&checkflags.CatalogSource, "catalogsource", "certified-operators",
		"specifies the catalogsource to test against")
	flags.StringVar(&checkflags.CatalogSourceNamespace, "catalogsourcenamespace", "openshift-marketplace",
		"specifies the namespace where the catalogsource exists")
	flags.StringSliceVar(&checkflags.AuditPlan, "audit-plan", defaultAuditPlan, "audit plan is the ordered list of operator test functions to be called during a capability audit.")
	flags.StringSliceVar(&checkflags.Packages, "packages", []string{}, "a list of package(s) which limits audits and/or other flag(s) output")
	flags.BoolVar(&checkflags.AllInstallModes, "all-installmodes", false, "when set, all install modes supported by an operator will be tested")
	flags.StringVar(&checkflags.ExtraCRDirectory, "extra-cr-directory", "",
		"directory containing the additional Custom Resources to be deployed by the OperandInstall audit. The manifest files should be located in subdirectories named after the packages they are corresponding to.")

	return cmd
}

func checkRunE(cmd *cobra.Command, args []string) error {
	kubeconfig, err := kubeConfig()
	if err != nil {
		return fmt.Errorf("could not get kubeconfig: %v", err)
	}

	client, err := operator.NewOpCapClient(kubeconfig)
	if err != nil {
		return fmt.Errorf("could not create client: %v", err)
	}

	fs := afero.NewOsFs()

	return runAudits(cmd.Context(), kubeconfig, client, fs, cmd.OutOrStdout())
}

func runAudits(ctx context.Context, kubeconfig *rest.Config, client operator.Client, fs afero.Fs, reportWriter io.Writer) error {
	// run all dynamically built audits in the auditor workqueue
	if err := capability.RunAudits(ctx,
		capability.WithAuditPlan(checkflags.AuditPlan),
		capability.WithCatalogSource(checkflags.CatalogSource),
		capability.WithCatalogSourceNamespace(checkflags.CatalogSourceNamespace),
		capability.WithPackages(checkflags.Packages),
		capability.WithAllInstallModes(checkflags.AllInstallModes),
		capability.WithClient(client),
		capability.WithExtraCRDirectory(checkflags.ExtraCRDirectory),
		capability.WithFilesystem(fs),
		capability.WithTimeout(time.Minute),
		capability.WithReportWriter(reportWriter),
	); err != nil {
		return err
	}

	return nil
}
