package cmd

import (
	. "github.com/onsi/ginkgo/v2/dsl/core"
	. "github.com/onsi/gomega"
)

var _ = Describe("List Bundles Cmd", func() {
	When("Calling opcap list bundles", func() {
		It("should succeed", func() {
			_, err := executeCommand(listBundlesCmd(), []string{}...)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
