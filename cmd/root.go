package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/opdev/opcap/internal/logger"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
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
	cmd.AddCommand(packageCmd())

	return &cmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(ctx context.Context) error {
	cmd := rootCmd()

	err := cmd.ExecuteContext(ctx)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Opcap tool execution failed: %v\n", err)
		return ctx.Err()
	}
	return nil
}

// kubeConfig return kubernetes cluster config
func kubeConfig() (*rest.Config, error) {
	config, err := ctrl.GetConfig()
	if err != nil {
		// returned when there is no kubeconfig
		if errors.Is(err, clientcmd.ErrEmptyConfig) {
			return nil, fmt.Errorf("please provide kubeconfig before retrying: %v", err)
		}

		// returned when the kubeconfig has no servers
		if errors.Is(err, clientcmd.ErrEmptyCluster) {
			return nil, fmt.Errorf("malformed kubeconfig. Please check before retrying: %v", err)
		}

		// any other errors getting kubeconfig would be caught here
		return nil, fmt.Errorf("error getting kubeocnfig. Please check before retrying: %v", err)
	}
	return config, nil
}
