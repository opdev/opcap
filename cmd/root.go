/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	log "opcap/internal/logger"
	"opcap/internal/operator"
	"os"

	"github.com/spf13/cobra"
)

var osversion string
var logger = log.Sugar

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
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

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		logger.Errorf("Opcap tool execution failed: %s", err)
		os.Exit(1)
	}
}

func init() {
	opClient, err := operator.NewClient()
	if err != nil {
		logger.Errorf("Failed to initialize OpenShift client: %s", err)
		os.Exit(1)
	}

	osversion, err = opClient.GetOpenShiftVersion()
	if err != nil {
		logger.Errorf("Failed to connect to OpenShift: %s", err)
		os.Exit(1)
	}
}
