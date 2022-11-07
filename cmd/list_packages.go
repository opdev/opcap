package cmd

import (
	"context"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/opdev/opcap/internal/packages"
	pkgserverv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var packageListFlags struct {
	CatalogSource string
	Packages      []string
}

func listPackagesCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "packages",
		Short: "List the package manifests for a given CatalogSource and Namespace",
		RunE:  listPackagesRunE,
	}

	flags := cmd.Flags()

	flags.StringVar(&packageListFlags.CatalogSource, "catalogsource", "certified-operators",
		"specifies the catalogsource to test against")
	flags.StringSliceVar(&packageListFlags.Packages, "packages", []string{}, "a list of package(s) which limits audits and/or other flag(s) output")

	return &cmd
}

func listPackagesRunE(cmd *cobra.Command, args []string) error {
	scheme := runtime.NewScheme()
	if err := pkgserverv1.AddToScheme(scheme); err != nil {
		return err
	}

	k8sconfig, err := config.GetConfig()
	if err != nil {
		return err
	}

	c, err := client.New(k8sconfig, client.Options{Scheme: scheme})
	if err != nil {
	}

	return listPackages(cmd.Context(), cmd.OutOrStdout(), c)
}

func listPackages(ctx context.Context, out io.Writer, c client.Client) error {
	packageManifestList, err := packages.List(ctx, c, packageListFlags.CatalogSource, packageListFlags.Packages)
	if err != nil {
		return err
	}

	headings := "Package Name\tCatalog Source\tCatalog Source Namespace"
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, headings)
	for _, packageManifest := range packageManifestList {
		packageInfo := []string{packageManifest.Name, packageManifest.Status.CatalogSource, packageManifest.Status.CatalogSourceNamespace}
		fmt.Fprintln(w, strings.Join(packageInfo, "\t"))
	}
	w.Flush()

	return nil
}
