package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Those variables are populated at build time by ldflags.
// If you're running from a local debugger they will show empty fields.

var (
	Version   string
	GoVersion string
	BuildTime string
	GitUser   string
	GitCommit string
)

func versionCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "version",
		Short: "Prints opcap's version information",
		Long:  `opcap, go and git commit information for this particular binary build are included at build time and can be accessed by this command`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "Version:\t%s\n", Version)
			fmt.Fprintf(cmd.OutOrStdout(), "Go Version:\t%s\n", GoVersion)
			fmt.Fprintf(cmd.OutOrStdout(), "Build Time:\t%s\n", BuildTime)
			fmt.Fprintf(cmd.OutOrStdout(), "Git User:\t%s\n", GitUser)
			fmt.Fprintf(cmd.OutOrStdout(), "Git Commit:\t%s\n", GitCommit)
		},
	}

	return &cmd
}
