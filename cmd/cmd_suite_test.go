package cmd

import (
	"bytes"
	"testing"

	. "github.com/onsi/ginkgo/v2/dsl/core"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

func TestCMD(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI Suite")
}

// executeCommand is used for cobra command testing. It is effectively what's seen here:
// https://github.com/spf13/cobra/blob/master/command_test.go#L34-L43. It should only
// be used in tests. Typically, you should pass rootCmd as the param for root, and your
// subcommand's invocation within args.
func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	err = root.Execute()

	return buf.String(), err
}
