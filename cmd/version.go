package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

// Those variables are populated at build time by ldflags.
// If you're running from a local debugger they will show empty fields.

var (
	Version   string
	GoVersion string
	BuildTime string
	GitUser   string
	GitCommit string
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints opcap's version information",
	Long:  `opcap, go and git commit information for this particular binary build are included at build time and can be accessed by this command`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Version:\t", Version)
		fmt.Println("Go Version:\t", GoVersion)
		fmt.Println("Build Time:\t", BuildTime)
		fmt.Println("Git User:\t", GitUser)
		fmt.Println("Git Commit:\t", GitCommit)
	},
}
