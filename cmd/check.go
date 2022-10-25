package cmd

import (
	"fmt"

	"github.com/opdev/opcap/internal/capability"
	"github.com/opdev/opcap/internal/operator"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

type CheckCommandFlags struct {
	AuditPlan              []string `json:"auditPlan"`
	CatalogSource          string   `json:"catalogsource"`
	CatalogSourceNamespace string   `json:"catalogsourcenamespace"`
	Packages               []string `json:"packages"`
	AllInstallModes        bool     `json:"allInstallModes"`
	ExtraCRDirectory       string   `json:"extraCRDirectory"`
}

var checkflags CheckCommandFlags

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
		RunE: func(cmd *cobra.Command, args []string) error {
			kubeconfig, err := kubeConfig()
			if err != nil {
				return fmt.Errorf("could not get kubeconfig: %v", err)
			}

			client, err := operator.NewOpCapClient(kubeconfig)
			if err != nil {
				return fmt.Errorf("could not create client: %v", err)
			}

			capAuditor := &capability.CapAuditor{
				AuditPlan:              checkflags.AuditPlan,
				CatalogSource:          checkflags.CatalogSource,
				CatalogSourceNamespace: checkflags.CatalogSourceNamespace,
				Packages:               checkflags.Packages,
				AllInstallModes:        checkflags.AllInstallModes,
				OpCapClient:            client,
			}

			if checkflags.ExtraCRDirectory != "" {
				if err := capAuditor.ExtraCRDirectory(checkflags.ExtraCRDirectory); err != nil {
					return err
				}
			}

			fs := afero.NewOsFs()

			// run all dynamically built audits in the auditor workqueue
			if err := capAuditor.RunAudits(cmd.Context(), fs, cmd.OutOrStdout()); err != nil {
				return err
			}

			return nil
		},
	}

	defaultAuditPlan := []string{"OperatorInstall", "OperatorCleanUp"}

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
