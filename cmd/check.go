package cmd

import (
	"fmt"
	"go/types"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/opdev/opcap/internal/capability"
	"github.com/opdev/opcap/internal/logger"
	"github.com/opdev/opcap/internal/operator"

	pkgserverv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"

	"github.com/spf13/cobra"
)

// TODO: provide godoc compatible comment for checkCmd
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Checks if operator meets minimum capability requirement.",
	Long: `The 'check' command checks if OpenShift operators meet minimum
requirements for Operator Capabilities Level to attest operator
advanced features by running custom resources provided by CSVs
and/or users.`,
	Example: "opcap check --catalogsource=certified-operators --catalogsourcenamespace=openshift-marketplace",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		logger.InitLogger(checkflags.LogLevel)
		psc, err := operator.NewOpCapClient()
		if err != nil {
			return types.Error{Msg: "Unable to create OpCap client."}
		}
		var packageManifestList pkgserverv1.PackageManifestList
		err = psc.ListPackageManifests(cmd.Context(), &packageManifestList, checkflags.CatalogSource, checkflags.Packages)
		if err != nil {
			return types.Error{Msg: "Unable to list PackageManifests.\n" + err.Error()}
		}

		if len(packageManifestList.Items) == 0 {
			return types.Error{Msg: "No PackageManifests returned from PackageServer."}
		}

		if checkflags.ListPackages {
			headings := "Package Name\tCatalog Source\tCatalog Source Namespace"
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, headings)
			for _, packageManifest := range packageManifestList.Items {
				packageInfo := []string{packageManifest.Name, packageManifest.Status.CatalogSource, packageManifest.Status.CatalogSourceNamespace}
				fmt.Fprintln(w, strings.Join(packageInfo, "\t"))
			}
			w.Flush()
			os.Exit(0)
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		capAuditor := &capability.CapAuditor{
			AuditPlan:              checkflags.AuditPlan,
			CatalogSource:          checkflags.CatalogSource,
			CatalogSourceNamespace: checkflags.CatalogSourceNamespace,
			Packages:               checkflags.Packages,
			AllInstallModes:        checkflags.AllInstallModes,
		}

		// run all dynamically built audits in the auditor workqueue
		capAuditor.RunAudits(cmd.Context())
	},
}

type CheckCommandFlags struct {
	AuditPlan              []string `json:"auditPlan"`
	CatalogSource          string   `json:"catalogsource"`
	CatalogSourceNamespace string   `json:"catalogsourcenamespace"`
	ListPackages           bool     `json:"listPackages"`
	Packages               []string `json:"packages"`
	LogLevel               string   `json:"loglevel"`
	AllInstallModes        bool     `json:"allInstallModes"`
}

var checkflags CheckCommandFlags

func init() {
	defaultAuditPlan := []string{"OperatorInstall", "OperatorCleanUp"}

	rootCmd.AddCommand(checkCmd)
	flags := checkCmd.Flags()

	flags.StringVar(&checkflags.CatalogSource, "catalogsource", "certified-operators",
		"specifies the catalogsource to test against")
	flags.StringVar(&checkflags.CatalogSourceNamespace, "catalogsourcenamespace", "openshift-marketplace",
		"specifies the namespace where the catalogsource exists")
	flags.StringSliceVar(&checkflags.AuditPlan, "audit-plan", defaultAuditPlan, "audit plan is the ordered list of operator test functions to be called during a capability audit.")
	flags.BoolVar(&checkflags.ListPackages, "list-packages", false, "list packages in the catalog")
	flags.StringSliceVar(&checkflags.Packages, "packages", []string{}, "a list of package(s) which limits audits and/or other flag(s) output")
	flags.StringVar(&checkflags.LogLevel, "log-level", "", "specifies the one of the log levels in order of decreasing verbosity: debug, error, info, warn")
	flags.BoolVar(&checkflags.AllInstallModes, "all-installmodes", false, "when set, all install modes supported by an operator will be tested")
}
