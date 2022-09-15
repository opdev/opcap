package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "opcap",
	Short: "A CLI that assesses the capability levels of operators from a specified catalog source",
	Long: `The OpenShift Operator Capabilities Tool (opcap) is a command line interface that assesses
the capability levels of operators from a specified catalog source`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd.AddCommand(versionCmd())

	err := rootCmd.ExecuteContext(context.Background())
	if err != nil {
		fmt.Fprintf(rootCmd.ErrOrStderr(), "Opcap tool execution failed: %v\n", err)
		os.Exit(1)
	}
}
