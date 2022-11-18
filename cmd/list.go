package cmd

import (
	"github.com/spf13/cobra"
)

func listCmd() *cobra.Command {
	// Run is empty. Otherwise, on an error, it would not be marked
	// as Runnable, which would not print out the usage/help.
	cmd := cobra.Command{
		Use:   "list",
		Short: "List commands",
		Long:  "Commands that will list various object",
	}

	cmd.AddCommand(listPackagesCmd())
	cmd.AddCommand(listBundlesCmd())

	return &cmd
}
