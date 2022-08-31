package cmd

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/opdev/opcap/internal/packages"

	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var packageListFlags struct {
	CatalogSource string
	Packages      []string
}

func packageListCmd(client client.Client) *cobra.Command {
	cmd := cobra.Command{
		Use:   "list",
		Short: "List the package manifests for a given CatalogSource and Namespace",
		RunE: func(cmd *cobra.Command, args []string) error {
			packageManifestList, err := packages.List(cmd.Context(), client, packageListFlags.CatalogSource, packageListFlags.Packages)
			if err != nil {
				return err
			}

			headings := "Package Name\tCatalog Source\tCatalog Source Namespace"
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, headings)
			for _, packageManifest := range packageManifestList {
				packageInfo := []string{packageManifest.Name, packageManifest.Status.CatalogSource, packageManifest.Status.CatalogSourceNamespace}
				fmt.Fprintln(w, strings.Join(packageInfo, "\t"))
			}
			w.Flush()

			return nil
		},
	}

	flags := cmd.Flags()

	flags.StringVar(&packageListFlags.CatalogSource, "catalogsource", "certified-operators",
		"specifies the catalogsource to test against")
	flags.StringSliceVar(&packageListFlags.Packages, "packages", []string{}, "a list of package(s) which limits audits and/or other flag(s) output")

	return &cmd
}
