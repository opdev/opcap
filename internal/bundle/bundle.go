package bundle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	operatorv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/util/yaml"

	git "gopkg.in/src-d/go-git.v4"
)

func GitCloneOrPullBundles(URL string, outputDir string) error {
	// Clone bundle repository, main branch only (should be the default for certified-operators)
	if _, err := git.PlainClone(outputDir, false, &git.CloneOptions{
		URL:      URL,
		Progress: os.Stdout,
	}); err != nil && err != git.ErrRepositoryAlreadyExists {
		return fmt.Errorf("failed cloning repository: %s", err)
	}
	// Pull data if repo already exists
	repo, err := git.PlainOpen(outputDir)
	if err != nil {
		return fmt.Errorf("failed opening existent repository: %s", err)
	}
	workTree, err := repo.Worktree()
	if err != nil {
		return err
	}
	err = workTree.Pull(&git.PullOptions{RemoteName: "origin"})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed pulling data from repository: %s", err)
	}
	return nil
}

func ReadBundlesFromDir(bundlesDir string) ([]Bundle, error) {
	bundles := []Bundle{}

	operators, err := os.ReadDir(filepath.Join(bundlesDir, "operators"))
	if err != nil {
		return nil, fmt.Errorf("failed to extract operators from repo: %s", err)
	}

	for _, operator := range operators {

		if !operator.IsDir() {
			continue
		}
		path := filepath.Join(bundlesDir, "operators", operator.Name())

		versions, err := os.ReadDir(path)
		if err != nil {
			return nil, fmt.Errorf("failed to extract versions for operator: %s", err)
		}

		// looking into each version folder
		for _, version := range versions {

			if !version.IsDir() {
				continue
			}
			bundle := Bundle{}
			bundle.Version = version.Name()
			versionDir := filepath.Join(path, version.Name())

			data, err := os.ReadDir(versionDir)
			if err != nil {
				return nil, fmt.Errorf("failed to extract bundle for version: %s", err)
			}

			for _, d := range data {

				// checking the manifests folder for the csv file
				if d.IsDir() && d.Name() == "manifests" {

					manifestPath := filepath.Join(versionDir, "manifests")

					if err := filepath.Walk(manifestPath, func(path string, f os.FileInfo, err error) error {
						if err == nil && strings.Contains(f.Name(), "clusterserviceversion") {
							bundle.StartingCSV, err = getStartingCsv(filepath.Join(manifestPath, f.Name()))
							if err != nil {
								return fmt.Errorf("failed to get StartingCSV for bundle: %s", err)
							}
						}

						return nil
					}); err != nil {
						return nil, err
					}
				}

				// Checking the metadata folder for the annotations file
				if d.IsDir() && d.Name() == "metadata" {
					annotationsPath := filepath.Join(versionDir, "metadata", "annotations.yaml")
					annotations, err := getAnnotations(annotationsPath)
					if err != nil {
						return nil, fmt.Errorf("failed to get metadata annotations for bundle: %s", err)
					}

					bundle.PackageName = annotations["operators.operatorframework.io.bundle.package.v1"]
					bundle.Channel = annotations["operators.operatorframework.io.bundle.channel.default.v1"]
					bundle.OcpVersions = annotations["com.redhat.openshift.versions"]
				}
			}
			bundles = append(bundles, bundle)
		}
	}
	return bundles, nil
}

func getStartingCsv(filePath string) (string, error) {
	fileReader, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("unable to load file %s: %s", filePath, err)
	}
	defer fileReader.Close()

	decoder := yaml.NewYAMLOrJSONDecoder(fileReader, 30)
	csv := operatorv1alpha1.ClusterServiceVersion{}

	if err = decoder.Decode(&csv); err != nil {
		return "", err
	}

	return csv.Name, nil
}

func getAnnotations(filePath string) (map[string]string, error) {
	f, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	annotationsFile := make(map[string]map[string]string)

	err = yaml.Unmarshal(f, &annotationsFile)
	if err != nil {
		return nil, err
	}

	return annotationsFile["annotations"], nil
}
