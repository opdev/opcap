package cmd

import (
	. "github.com/onsi/ginkgo/v2/dsl/core"
	. "github.com/onsi/gomega"
)

var _ = Describe("Version", func() {
	When("running the version command", func() {
		It("should print the things", func() {
			out, err := executeCommand(versionCmd())
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(MatchRegexp("Version:\\s+foo"))
			Expect(out).To(MatchRegexp("Git Commit:\\s+bar"))
		})
	})
})
