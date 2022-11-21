package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/opdev/opcap/internal/bundle"

	"github.com/spf13/cobra"
)

var listBundlesFlags struct {
	bundlesDir  string
	bundlesRepo string
}

func listBundlesCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "bundles",
		Short: "List all bundles and versions",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := listBundles(cmd.Context(), cmd.OutOrStdout())
			if err != nil {
				return err
			}
			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&listBundlesFlags.bundlesDir, "from-dir", "",
		"specifies the source directory with bundles under the operators directory, can't be used with --from-repo")
	flags.StringVar(&listBundlesFlags.bundlesRepo, "from-repo", "https://github.com/redhat-openshift-ecosystem/certified-operators.git",
		"Git repository URL from where to download bundles, can't be used with --from-dir")

	cmd.MarkFlagsMutuallyExclusive("from-dir", "from-repo")

	return &cmd
}

func listBundles(ctx context.Context, out io.Writer) error {
	if listBundlesFlags.bundlesDir != "" {
		bundles, err := bundle.ReadBundlesFromDir(listBundlesFlags.bundlesDir)
		if err != nil {
			return err
		}
		printBundles(bundles, out)
		return nil
	}

	dir, err := os.MkdirTemp("", "bundles-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	if err = bundle.GitCloneOrPullBundles(listBundlesFlags.bundlesRepo, dir); err != nil {
		return err
	}

	bundles, err := bundle.ReadBundlesFromDir(dir)
	if err != nil {
		return err
	}
	printBundles(bundles, out)
	return nil
}

func printBundles(bundles []bundle.Bundle, out io.Writer) {
	headings := "Package Name\tStarting CSV\tVersion\tDefault Channel\tOcpVersions"
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, headings)
	for _, bundle := range bundles {
		packageInfo := []string{bundle.PackageName, bundle.StartingCSV, bundle.Channel, bundle.Version, bundle.OcpVersions}
		fmt.Fprintln(w, strings.Join(packageInfo, "\t"))
	}
	w.Flush()
}
