package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
func rootCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "opcap",
		Short: "A brief description of your application",
		Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		// Uncomment the following line if your bare application
		// has an action associated with it:
		// Run: func(cmd *cobra.Command, args []string) { },
	}

	cmd.AddCommand(checkCmd())
	cmd.AddCommand(uploadCmd())
	cmd.AddCommand(versionCmd())

	return &cmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd().ExecuteContext(context.Background())
}
