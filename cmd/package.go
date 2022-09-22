package cmd

import (
	"fmt"

	pkgserverv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func errorPreRunE(message string, err error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err != nil {
			return fmt.Errorf("%s: %v", message, err)
		}
		return fmt.Errorf("%s", message)
	}
}

func packageCmd() *cobra.Command {
	// Run is empty. Otherwise, on an error, it would not be marked
	// as Runnable, which would not print out the usage/help.
	// TODO: Can we add the subcommand before the client is established?
	cmd := cobra.Command{
		Use:   "package",
		Short: "Package commands",
		Long:  "Commands that will operate on package manifests",
		Run:   func(cmd *cobra.Command, args []string) {},
	}

	scheme := runtime.NewScheme()
	if err := pkgserverv1.AddToScheme(scheme); err != nil {
		cmd.PreRunE = errorPreRunE("unable to add scheme", err)
		return &cmd
	}

	k8sconfig, err := config.GetConfig()
	if err != nil {
		cmd.PreRunE = errorPreRunE("unable to establish kubeconfig", nil)
		return &cmd
	}

	c, err := client.New(k8sconfig, client.Options{Scheme: scheme})
	if err != nil {
		cmd.PreRunE = errorPreRunE("unable to create controller-runtime client: %v", err)
		return &cmd
	}

	cmd.AddCommand(packageListCmd(c))

	// We have the subcommand now, so make Run nil to trigger the usage/help
	// properly when no other subcommands are present on the CLI.
	cmd.Run = nil

	return &cmd
}
