package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/opdev/opcap/internal/logger"
	"github.com/spf13/cobra"
)

var logLevel string

// rootCmd represents the base command when called without any subcommands
func rootCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "opcap",
		Short: "A CLI that assesses the capability levels of operators from a specified catalog source",
		Long: `The OpenShift Operator Capabilities Tool (opcap) is a command line interface that assesses
	the capability levels of operators from a specified catalog source`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return logger.InitLogger(logLevel)
		},
	}

	cmd.PersistentFlags().StringVar(&logLevel, "log-level", "", "specifies the one of the log levels in order of decreasing verbosity: debug, error, info, warn")

	cmd.AddCommand(uploadCmd())
	cmd.AddCommand(versionCmd())
	cmd.AddCommand(checkCmd())

	return &cmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cmd := rootCmd()
	err := cmd.ExecuteContext(context.Background())
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Opcap tool execution failed: %v\n", err)
		os.Exit(1)
	}
}
