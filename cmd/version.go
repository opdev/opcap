package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

// Those variables are populated at build time by ldflags.
// If you're running from a local debugger they will show empty fields.

var (
	// Version is the current version of the binary
	Version string

	// GoVersion is the go version used to build the binary
	GoVersion string

	// BuildTime is when the binary was built
	BuildTime string

	// GitUser is the git user that built this binary
	GitUser string

	// GitCommit is the commit that this binary is based on
	GitCommit string
)

func versionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Prints opcap's version information",
		Long:  `opcap, go and git commit information for this particular binary build are included at build time and can be accessed by this command`,
		Run: func(cmd *cobra.Command, args []string) {
			printVersionInfo(cmd.OutOrStdout())
		},
	}

	return cmd
}

func printVersionInfo(w io.Writer) {
	fmt.Fprintf(w, "Version:\t%s\n", Version)
	fmt.Fprintf(w, "Go Version:\t%s\n", GoVersion)
	fmt.Fprintf(w, "Build Time:\t%s\n", BuildTime)
	fmt.Fprintf(w, "Git User:\t%s\n", GitUser)
	fmt.Fprintf(w, "Git Commit:\t%s\n", GitCommit)
}
