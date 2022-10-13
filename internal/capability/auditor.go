package capability

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/opdev/opcap/internal/logger"
	"github.com/opdev/opcap/internal/operator"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// capAuditor implements Auditor
type CapAuditor struct {
	// AuditPlan holds the tests that should be run during an audit
	AuditPlan []string

	// CatalogSource may be built-in OLM or custom
	CatalogSource string
	// CatalogSourceNamespace will be openshift-marketplace or custom
	CatalogSourceNamespace string

	// Packages is a subset of packages to be tested from a catalogSource
	Packages []string

	// WorkQueue holds capAudits in a buffered channel in order to execute them
	WorkQueue chan capAudit

	// AllInstallModes will test all install modes supported by an operator
	AllInstallModes bool

	// extraCustomResources associates packages to a list of Custom Resources (in addition to ALMExamples)
	// to be audited by the OperandInstall AuditPlan.
	extraCustomResources map[string]interface{}

	// OpCapClient is the main OpenShift client interface
	OpCapClient operator.Client
}

// ExtraCRDirectory scans the provided directory and populates the extraCustomResources field.
// Is is expected that the extraCRDirectory posesses subdirectories. Manifest files are present in each subdirectory.
// The name of the subdirectory is used to determine which package the manifest files are corresponding to.
// The resulting structure would be:
// custom_resources_directory/
// ├── package_name1
// │   ├── manifest_file1.json
// │   └── manifest_file2.yaml
// └── package_name2
//
//	   ├── manifest_file1.json
//
//     └── manifest_file2.yaml
func (capAuditor *CapAuditor) ExtraCRDirectory(extraCRDirectory string) error {
	logger.Debugw("scaning for extra Custom Resources", "extra CR directory", extraCRDirectory)
	extraCustomResources := map[string]interface{}{} // maps packages to a list of CR

	extraCRDirectoryAbsolutePath, err := filepath.Abs(extraCRDirectory)
	if err != nil {
		return fmt.Errorf("could not get absolute path from %s: %v", extraCRDirectory, err)
	}

	err = filepath.WalkDir(extraCRDirectoryAbsolutePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil && path == extraCRDirectoryAbsolutePath {
			// Error reading the root directory, exit and return the error
			return err
		}

		if !d.IsDir() { // Act on files only
			manifestFilePath := path
			// Checking that the manifest file is placed in a subdirectory
			if len(strings.Split(manifestFilePath, "/")) != len(strings.Split(extraCRDirectoryAbsolutePath, "/"))+2 {
				logger.Errorf("Error handling manifest file %s. File should be placed in a subdirectory of %s", manifestFilePath, extraCRDirectory)
				return nil // continue
			}

			// Get the name of the subdirectory containing the manifest file. This corresponds to the package name.
			packageName := filepath.Base(filepath.Dir(manifestFilePath))

			logger.Debugw("adding Custom Resource", "source manifest file", manifestFilePath, "package", packageName)

			// Get manifest file content
			manifestBytes, err := os.ReadFile(manifestFilePath)
			if err != nil {
				logger.Errorf("Error reading file %s: %v", manifestFilePath, err)
				return nil // continue
			}

			var manifest map[string]interface{}
			err = yaml.Unmarshal(manifestBytes, &manifest)
			if err != nil {
				logger.Errorf("Error unmarshalling file %s: %v", manifestFilePath, err)
				return nil // continue
			}

			if manifest == nil {
				logger.Errorf("Empty manifest file %s", manifestFilePath)
				return nil // continue
			}

			// Add the Custom Resource to the list of extra Custom Resource for this package
			var customResourceManifests []map[string]interface{}
			if _, packageKeyPresent := extraCustomResources[packageName]; packageKeyPresent {
				customResourceManifests = extraCustomResources[packageName].([]map[string]interface{})
			}
			customResourceManifests = append(customResourceManifests, manifest)

			extraCustomResources[packageName] = customResourceManifests
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("could not read directory %s: %v", extraCRDirectory, err)
	}

	capAuditor.extraCustomResources = extraCustomResources

	return nil
}

// BuildWorkQueueByCatalog fills in the auditor workqueue with all package information found in a specific catalog
func (capAuditor *CapAuditor) buildWorkQueueByCatalog(ctx context.Context, c operator.Client) error {
	// Getting subscription data form the package manifests available in the selected catalog
	subscriptions, err := c.GetSubscriptionData(ctx, capAuditor.CatalogSource, capAuditor.CatalogSourceNamespace, capAuditor.Packages)
	if err != nil {
		return fmt.Errorf("could not get bundles from CatalogSource: %s: %v", capAuditor.CatalogSource, err)
	}

	// build workqueue as buffered channel based subscriptionData list size
	capAuditor.WorkQueue = make(chan capAudit, len(subscriptions))
	defer close(capAuditor.WorkQueue)

	// packagesToBeAudited is a subset of packages to be tested from a catalogSource
	var packagesToBeAudited []operator.SubscriptionData

	// get all install modes for all operators in the catalog
	// and add them to the packagesToBeAudited list
	if capAuditor.AllInstallModes {
		packagesToBeAudited = subscriptions
	} else {
		packages := make(map[string]bool)
		for _, subscription := range subscriptions {
			if _, exists := packages[subscription.Package]; !exists {
				packages[subscription.Package] = true
				packagesToBeAudited = append(packagesToBeAudited, subscription)
			}
		}
	}

	// add capAudits to the workqueue
	for _, subscription := range packagesToBeAudited {
		// Get extra Custom Resources for this subscription, if any
		mapExtraCustomResources := []map[string]interface{}{}
		extraCustomResources, ok := capAuditor.extraCustomResources[subscription.Package]
		if ok {
			mapExtraCustomResources = extraCustomResources.([]map[string]interface{})
		}

		capAudit, err := newCapAudit(ctx, c, subscription, capAuditor.AuditPlan, mapExtraCustomResources)
		if err != nil {
			return fmt.Errorf("could not build configuration for subscription: %s: %v", subscription.Name, err)
		}

		// load workqueue with capAudit
		capAuditor.WorkQueue <- *capAudit
	}

	return nil
}

// RunAudits executes all selected functions in order for a given audit at a time
func (capAuditor *CapAuditor) RunAudits(ctx context.Context) error {
	err := capAuditor.buildWorkQueueByCatalog(ctx, capAuditor.OpCapClient)
	if err != nil {
		return fmt.Errorf("unable to build workqueue: %v", err)
	}

	// read workqueue for audits
	for audit := range capAuditor.WorkQueue {
		// read a particular audit's auditPlan for functions
		// to be executed against operator
		for _, function := range audit.auditPlan {
			// run function/method by name
			// NOTE: The signature for this method MUST be:
			// func Fn(context.Context) error
			auditFn := newAudit(ctx, function,
				withClient(audit.client),
				withNamespace(audit.namespace),
				withOperatorGroupData(&audit.operatorGroupData),
				withSubscription(&audit.subscription),
				withTimeout(int(audit.csvWaitTime)),
				withCustomResources(audit.customResources),
			)
			if auditFn == nil {
				logger.Errorf("invalid audit plan specified: %s", function)
				continue
			}
			err := auditFn(ctx)
			if err != nil {
				logger.Errorf("error in audit: %v", err)
				break
			}
		}
	}
	return nil
}
