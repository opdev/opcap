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
	Short: "A cli tool to gauge capabilities of operators available on OpenShift's operator hub.",
	Long:  `Opcap is a cli tool that interacts with all operators available on OpenShift's operator hub in a live cluster, to assess how smart and cloud native they are.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.ExecuteContext(context.Background())
	if err != nil {
		fmt.Fprintf(rootCmd.ErrOrStderr(), "Opcap tool execution failed: %v\n", err)
		os.Exit(1)
	}
}
