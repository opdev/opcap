package cmd

import (
	"os"

	. "github.com/onsi/ginkgo/v2/dsl/core"
	. "github.com/onsi/gomega"
)

var _ = Describe("Package CMD", func() {
	BeforeEach(func() {
		DeferCleanup(os.Setenv, "KUBECONFIG", os.Getenv("KUBECONFIG"))
		os.Unsetenv("KUBECONFIG")
	})
	When("Initializing the command", func() {
		It("should fail", func() {
			cmd := listCmd()
			// If PreRunE && Run are not nil, there was a failure
			Expect(cmd.PreRunE).ToNot(BeNil())
			Expect(cmd.Run).ToNot(BeNil())
		})
	})
	When("Executing the command", func() {
		It("should fail", func() {
			out, err := executeCommand(listCmd())
			Expect(err).To(HaveOccurred())
			Expect(out).To(ContainSubstring("unable to establish kubeconfig"))
		})
	})
})
