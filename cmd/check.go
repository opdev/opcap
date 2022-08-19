package cmd

import (
	"context"
	"fmt"
	"go/types"
	"os"

	"github.com/opdev/opcap/internal/capability"
	"github.com/opdev/opcap/internal/logger"
	"github.com/opdev/opcap/internal/operator"

	pkgserverv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"

	"github.com/spf13/cobra"
)

// TODO: provide godoc compatible comment for checkCmd
var checkCmd = &cobra.Command{
	Use: "check",
	// TODO: provide Short description for check command
	Short: "Checks if operator meets minimum capability requirement.",
	// TODO: provide Long description for check command
	Long: `The 'check' command checks if OpenShift operators meet minimum
requirements for Operator Capabilities Level to attest operator
advanced features by running custom resources provided by CSVs
and/or users.

Usage:
opcap check [flags]

Example:
opcap check --catalogsource=certified-operators --catalogsourcenamespace=openshift-marketplace --list-packages=false'

Flags:
--catalogsource				specifies the catalogsource to test against
--catalogsourcenamespace	specifies the namespace where the catalogsource exists
--auditplan					audit plan is the ordered list of operator test functions to be called during a capability audit
--list-packages				list packages in the catalog
--filter-packages			 one or more package names (comma-separated) that limit audits and/or other flag(s) output
`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		psc, err := operator.NewOpCapClient()
		if err != nil {
			return types.Error{Msg: "Unable to create OpCap client."}
		}
		var packageManifestList pkgserverv1.PackageManifestList
		err = psc.ListPackageManifests(context.TODO(), &packageManifestList, checkflags.FilterPackages)
		if err != nil {
			return types.Error{Msg: "Unable to list PackageManifests.\n" + err.Error()}
		}

		if len(packageManifestList.Items) == 0 {
			return types.Error{Msg: "No PackageManifests returned from PackageServer."}
		}

		if checkflags.ListPackages {
			for _, packageManifest := range packageManifestList.Items {
				fmt.Println(packageManifest.Name)
			}
			os.Exit(0)
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("check called")

		// Build auditor by catalog
		auditor, err := capability.BuildAuditorByCatalog(checkflags.CatalogSource, checkflags.CatalogSourceNamespace, checkflags.AuditPlan, checkflags.FilterPackages)
		if err != nil {
			logger.Sugar.Fatal("Unable to build auditor")
		}
		// run all dynamically built audits in the auditor workqueue
		auditor.RunAudits()
	},
}

type CheckCommandFlags struct {
	AuditPlan              []string `json:"auditPlan"`
	CatalogSource          string   `json:"catalogsource"`
	CatalogSourceNamespace string   `json:"catalogsourcenamespace"`
	ListPackages           bool     `json:"listPackages"`
	FilterPackages         []string `json:"filterPackages"`
}

var checkflags CheckCommandFlags

func init() {

	var defaultAuditPlan = []string{"OperatorInstall", "OperatorCleanUp"}

	rootCmd.AddCommand(checkCmd)
	flags := checkCmd.Flags()

	flags.StringVar(&checkflags.CatalogSource, "catalogsource", "certified-operators",
		"specifies the catalogsource to test against")
	flags.StringVar(&checkflags.CatalogSourceNamespace, "catalogsourcenamespace", "openshift-marketplace",
		"specifies the namespace where the catalogsource exists")
	flags.StringSliceVar(&checkflags.AuditPlan, "auditplan", defaultAuditPlan, "audit plan is the ordered list of operator test functions to be called during a capability audit.")
	flags.BoolVar(&checkflags.ListPackages, "list-packages", false, "list packages in the catalog")
	flags.StringSliceVar(&checkflags.FilterPackages, "filter-packages", []string{}, "a list of package(s) which limits audits and/or other flag(s) output")
}
