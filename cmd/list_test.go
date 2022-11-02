package cmd

import (
	. "github.com/onsi/ginkgo/v2/dsl/core"
	. "github.com/onsi/gomega"
)

var _ = Describe("List CMD", func() {
	When("Initializing the command", func() {
		It("should not error", func() {
			cmd := listCmd()
			Expect(cmd.PreRunE).To(BeNil())
			Expect(cmd.Run).To(BeNil())
		})
	})
	When("Executing the command", func() {
		It("should not error", func() {
			_, err := executeCommand(listCmd())
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
