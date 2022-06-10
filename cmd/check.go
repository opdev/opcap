/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"opcap/internal/capability"

	"github.com/spf13/cobra"
)

type CheckCommandFlags struct {
	CatalogSource          string `json:"catalogsource"`
	CatalogSourceNamespace string `json:"catalogsourcenamespace"`
}

var checkflags CheckCommandFlags

// TODO: provide godoc compatible comment for checkCmd
var checkCmd = &cobra.Command{
	Use: "check",
	// TODO: provide Short description for check command
	Short: "",
	// TODO: provide Long description for check command
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("check called")
		capability.OperatorInstallAllFromCatalog(checkflags.CatalogSource, checkflags.CatalogSourceNamespace)
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
	flags := checkCmd.Flags()

	flags.StringVar(&checkflags.CatalogSource, "catalogsource", "certified-operators",
		"the catalog source to use for audit")
	flags.StringVar(&checkflags.CatalogSourceNamespace, "catalogsourcenamespace", "openshift-marketplace",
		"the namespace/project the catalog source can be found in")
}
