package cmd

import (
	"fmt"
	"os"

	"github.com/opdev/opcap/internal/capability"
	"github.com/opdev/opcap/internal/operator"

	pkgserverv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"

	"github.com/spf13/cobra"
)

var checkflags operator.OperatorCheckOptions

// TODO: provide godoc compatible comment for checkCmd
func checkCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "check",
		Short: "Checks if operator meets minimum capability requirement.",
		Long: `The 'check' command checks if OpenShift operators meet minimum
requirements for Operator Capabilities Level to attest operator
advanced features by running custom resources provided by CSVs
and/or users.`,
		Example: `opcap check --catalogsource=certified-operators --catalogsourcenamespace=openshift-marketplace --list-packages=false'`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if checkflags.ListPackages {
				psc, err := operator.NewOpCapClient()
				if err != nil {
					return fmt.Errorf("unable to create OpCap client: %v", err)
				}
				var packageManifestList pkgserverv1.PackageManifestList
				err = psc.ListPackageManifests(cmd.Parent().Context(), &packageManifestList, checkflags)
				if err != nil {
					return fmt.Errorf("unable to list PackageManifests: %v", err)
				}

				if len(packageManifestList.Items) == 0 {
					return fmt.Errorf("no PackageManifests returned from PackageServer")
				}

				for _, packageManifest := range packageManifestList.Items {
					fmt.Println(packageManifest.Name)
				}
				os.Exit(0)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			psc, err := operator.NewOpCapClient()
			if err != nil {
				return fmt.Errorf("unable to create OpCap client: %v", err)
			}
			var packageManifestList pkgserverv1.PackageManifestList
			err = psc.ListPackageManifests(cmd.Parent().Context(), &packageManifestList, checkflags)
			if err != nil {
				return fmt.Errorf("unable to list PackageManifests: %v", err)
			}

			if len(packageManifestList.Items) == 0 {
				return fmt.Errorf("no PackageManifests returned from PackageServer")
			}
			// Build auditor by catalog
			auditor, err := capability.BuildAuditorByCatalog(checkflags)
			if err != nil {
				return err
			}
			// run all dynamically built audits in the auditor workqueue
			return auditor.RunAudits()
		},
	}

	defaultAuditPlan := []string{"OperatorInstall", "OperatorCleanUp"}

	flags := cmd.Flags()

	flags.StringVar(&checkflags.CatalogSource, "catalogsource", "certified-operators",
		"specifies the catalogsource to test against")
	flags.StringVar(&checkflags.CatalogSourceNamespace, "catalogsourcenamespace", "openshift-marketplace",
		"specifies the namespace where the catalogsource exists")
	flags.StringSliceVar(&checkflags.AuditPlan, "auditplan", defaultAuditPlan, "audit plan is the ordered list of operator test functions to be called during a capability audit.")
	flags.BoolVar(&checkflags.ListPackages, "list-packages", false, "list packages in the catalog")
	flags.StringSliceVar(&checkflags.FilterPackages, "filter-packages", []string{}, "a list of package(s) which limits audits and/or other flag(s) output")
	flags.BoolVar(&checkflags.AllInstallModes, "all-installmodes", false, "when set, all install modes supported by an operator will be tested")

	return &cmd
}
