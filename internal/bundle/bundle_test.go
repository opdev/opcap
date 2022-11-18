package bundle

import (
	"os"

	. "github.com/onsi/ginkgo/v2/dsl/core"
	. "github.com/onsi/gomega"
)

var _ = Describe("Package Bundle", func() {
	When("Cloning or pulling bundle repository", func() {
		// valid url to be cloned
		URL := "https://github.com/redhat-openshift-ecosystem/certified-operators.git"

		It("should succeed", func() {
			dir, err := os.MkdirTemp(".", "bundles")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(dir)

			err = GitCloneOrPullBundles(URL, dir)
			Expect(err).ToNot(HaveOccurred())
		})
	})
	When("Listing bundles with channels", func() {
		// path to cloned repo
		repoPath := "testdata"
		testData := []Bundle{
			{
				PackageName: "acc-operator",
				Channel:     "stable",
				Version:     "21.10.19",
				StartingCSV: "acc-operator.v21.10.19",
				OcpVersions: "v4.6-v4.8",
			},
			{
				PackageName: "acc-operator",
				Channel:     "alpha",
				Version:     "21.10.7",
				StartingCSV: "acc-operator.v21.10.7",
				OcpVersions: "v4.6-v4.8",
			},
			{
				PackageName: "acc-operator",
				Channel:     "stable",
				Version:     "21.12.60",
				StartingCSV: "acc-operator.v21.12.60",
				OcpVersions: "v4.6-v4.8",
			},
		}
		It("should succeed", func() {
			bundles, err := ReadBundlesFromDir(repoPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(bundles).To(ContainElements(testData))
		})
	})
	When("running getStartingCSV", func() {
		It("should succeed", func() {
			startingCSV, err := getStartingCsv("testdata/operators/acc-operator/21.10.7/manifests/acc-operator.clusterserviceversion.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(startingCSV).To(Equal("acc-operator.v21.10.7"))
		})
	})
	When("running getAnnotatinons", func() {
		It("should succeed", func() {
			annotations, err := getAnnotations("testdata/operators/acc-operator/21.10.7/metadata/annotations.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(annotations["operators.operatorframework.io.bundle.package.v1"]).To(Equal("acc-operator"))
		})
	})
})
