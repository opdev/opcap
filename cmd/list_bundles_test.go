package cmd

import (
	. "github.com/onsi/ginkgo/v2/dsl/core"
	. "github.com/onsi/gomega"
)

var _ = Describe("List Bundles Cmd", func() {
	When("Calling opcap list bundles", func() {
		var gitUrl string
		var err error

		// test list bundles '--from-repo' flag
		Context("repo flag", func() {
			It("should succeed", func() {
				gitUrl = "https://github.com/redhat-openshift-ecosystem/redhat-marketplace-operators.git"
				_, err = executeCommand(listBundlesCmd(), []string{"--from-repo=" + gitUrl}...)
				Expect(err).ToNot(HaveOccurred())
			})
		})
		// test list bundles cmd '--from-dir' flag
		Context("dir flag", func() {
			It("should succeed", func() {
				_, err := executeCommand(listBundlesCmd(), []string{"--from-dir=", "../internal/bundle/testdata/operators"}...)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
