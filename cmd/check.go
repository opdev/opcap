/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"fmt"
	"go/types"
	"opcap/internal/operator"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
	PreRunE: func(cmd *cobra.Command, args []string) error {
		psc, err := operator.NewPackageServerClient()
		if err != nil {
			return types.Error{Msg: "Unable to create PackageServer client."}
		}

		pml, err := psc.OperatorsV1().PackageManifests("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return types.Error{Msg: "Unable to list PackageManifests."}
		}

		if len(pml.Items) == 0 {
			return types.Error{Msg: "No PackageManifests returned from PackageServer."}
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("check called")
		//capability.OperandInstallForOperator("opcap-dynatrace-operator")
		capability.OperatorInstallAllFromCatalog(checkflags.CatalogSource, checkflags.CatalogSourceNamespace)
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
	flags := checkCmd.Flags()

	flags.StringVar(&checkflags.CatalogSource, "catalogsource", "certified-operators",
		"")
	flags.StringVar(&checkflags.CatalogSourceNamespace, "catalogsourcenamespace", "openshift-marketplace",
		"")
}
