package capability

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/opdev/opcap/internal/logger"
	"github.com/opdev/opcap/internal/operator"
	"github.com/spf13/afero"
	"k8s.io/apimachinery/pkg/util/yaml"
)

type customResources = map[string][]map[string]interface{}

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
func extraCRDirectory(ctx context.Context, options *auditorOptions) (customResources, error) {
	logger.Debugw("scaning for extra Custom Resources", "extra CR directory", options.extraCustomResources)
	extraCustomResources := customResources{} // maps packages to a list of CR

	extraCRDirectoryAbsolutePath, err := filepath.Abs(options.extraCustomResources)
	if err != nil {
		return nil, fmt.Errorf("could not get absolute path from %s: %v", options.extraCustomResources, err)
	}

	err = afero.Walk(options.fs, extraCRDirectoryAbsolutePath, func(path string, d fs.FileInfo, err error) error {
		if err != nil && path == extraCRDirectoryAbsolutePath {
			// Error reading the root directory, exit and return the error
			return err
		}

		if !d.IsDir() { // Act on files only
			manifestFilePath := path
			// Checking that the manifest file is placed in a subdirectory
			if len(strings.Split(manifestFilePath, "/")) != len(strings.Split(extraCRDirectoryAbsolutePath, "/"))+2 {
				logger.Errorf("Error handling manifest file %s. File should be placed in a subdirectory of %s", manifestFilePath, options.extraCustomResources)
				return nil // continue
			}

			// Get the name of the subdirectory containing the manifest file. This corresponds to the package name.
			packageName := filepath.Base(filepath.Dir(manifestFilePath))

			logger.Debugw("adding Custom Resource", "source manifest file", manifestFilePath, "package", packageName)

			// Get manifest file content
			manifestBytes, err := afero.ReadFile(options.fs, manifestFilePath)
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
				customResourceManifests = extraCustomResources[packageName]
			}
			customResourceManifests = append(customResourceManifests, manifest)

			extraCustomResources[packageName] = customResourceManifests
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("could not read directory %s: %v", options.extraCustomResources, err)
	}

	return extraCustomResources, nil
}

// BuildWorkQueueByCatalog fills in the auditor workqueue with all package information found in a specific catalog
func buildWorkQueueByCatalog(ctx context.Context, options *auditorOptions, extraCustomResources customResources) error {
	// Getting subscription data form the package manifests available in the selected catalog
	subscriptions, err := options.opCapClient.GetSubscriptionData(ctx, options.catalogSource, options.catalogSourceNamespace, options.packages)
	if err != nil {
		return fmt.Errorf("could not get bundles from CatalogSource: %s: %v", options.catalogSource, err)
	}

	// build workqueue as buffered channel based subscriptionData list size
	options.workQueue = make(chan capAudit, len(subscriptions))
	defer close(options.workQueue)

	// packagesToBeAudited is a subset of packages to be tested from a catalogSource
	var packagesToBeAudited []operator.SubscriptionData

	// get all install modes for all operators in the catalog
	// and add them to the packagesToBeAudited list
	if options.allInstallModes {
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
		extraCustomResources, ok := extraCustomResources[subscription.Package]
		if ok {
			mapExtraCustomResources = extraCustomResources
		}

		capAudit, err := newCapAudit(ctx, options.opCapClient, subscription, options.auditPlan, mapExtraCustomResources)
		if err != nil {
			return fmt.Errorf("could not build configuration for subscription: %s: %v", subscription.Name, err)
		}

		// load workqueue with capAudit
		options.workQueue <- *capAudit
	}

	return nil
}

func cleanup(ctx context.Context, stack *Stack[auditCleanupFn]) {
	if stack == nil {
		return
	}
	for !stack.Empty() {
		logger.Debugw("cleaning up...")
		cleaner, err := stack.Pop()
		if errors.Is(err, StackEmptyError) {
			break
		}
		if err := cleaner(ctx); err != nil {
			logger.Errorf("cleanup failed: %v", err)
		}
	}
	return
}

// RunAudits executes all selected functions in order for a given audit at a time
func RunAudits(ctx context.Context, opts ...auditorOption) error {
	var options auditorOptions
	for _, opt := range opts {
		err := opt(&options)
		if err != nil {
			return err
		}
	}

	cleanups := Stack[auditCleanupFn]{}
	defer cleanup(ctx, &cleanups)

	var extraCustomResources customResources
	if options.extraCustomResources != "" {
		var err error
		extraCustomResources, err = extraCRDirectory(ctx, &options)
		if err != nil {
			return fmt.Errorf("could not read extra custom resources directory: %v", err)
		}
	}

	err := buildWorkQueueByCatalog(ctx, &options, extraCustomResources)
	if err != nil {
		return fmt.Errorf("unable to build workqueue: %v", err)
	}

	// read workqueue for audits
	for audit := range options.workQueue {
		// read a particular audit's auditPlan for functions
		// to be executed against operator
		for _, function := range audit.auditPlan {
			// run function/method by name
			// NOTE: The signature for this method MUST be:
			// func Fn(context.Context) error
			auditFn, auditCleanupFn := newAudit(ctx, function,
				withClient(audit.client),
				withNamespace(audit.namespace),
				withOperatorGroupData(&audit.operatorGroupData),
				withSubscription(&audit.subscription),
				withTimeout(options.timeout),
				withCustomResources(audit.customResources),
				withFilesystem(options.fs),
				withReportWriter(options.reportWriter),
				withDetailedReports(options.detailedReports),
			)
			if auditFn == nil {
				logger.Errorf("invalid audit plan specified: %s", function)
				continue
			}
			cleanups.Push(auditCleanupFn)
			err := auditFn(ctx)
			if err != nil {
				logger.Errorf("error in audit: %v", err)
				break
			}
		}

		// Perform the cleanups now for this audit
		cleanup(ctx, &cleanups)
	}
	return nil
}

func WithAuditPlan(auditPlan []string) auditorOption {
	return func(options *auditorOptions) error {
		if len(auditPlan) == 0 {
			return fmt.Errorf("audit plan cannot be empty")
		}
		for _, plan := range auditPlan {
			if len(plan) == 0 {
				return fmt.Errorf("audit plan incorrectly specified")
			}
		}
		options.auditPlan = auditPlan
		return nil
	}
}

func WithCatalogSource(catalogSource string) auditorOption {
	return func(options *auditorOptions) error {
		options.catalogSource = catalogSource
		return nil
	}
}

func WithCatalogSourceNamespace(catalogSourceNamespace string) auditorOption {
	return func(options *auditorOptions) error {
		options.catalogSourceNamespace = catalogSourceNamespace
		return nil
	}
}

func WithPackages(packages []string) auditorOption {
	return func(options *auditorOptions) error {
		options.packages = packages
		return nil
	}
}

func WithAllInstallModes(allInstallModes bool) auditorOption {
	return func(options *auditorOptions) error {
		options.allInstallModes = allInstallModes
		return nil
	}
}

func WithClient(client operator.Client) auditorOption {
	return func(options *auditorOptions) error {
		if client == nil {
			return fmt.Errorf("client cannot be nil")
		}
		options.opCapClient = client
		return nil
	}
}

func WithExtraCRDirectory(extraCRDirectory string) auditorOption {
	return func(options *auditorOptions) error {
		options.extraCustomResources = extraCRDirectory
		return nil
	}
}

func WithFilesystem(fs afero.Fs) auditorOption {
	return func(options *auditorOptions) error {
		if fs == nil {
			return fmt.Errorf("filesystem must be specified")
		}
		options.fs = fs
		return nil
	}
}

func WithTimeout(timeout time.Duration) auditorOption {
	return func(options *auditorOptions) error {
		options.timeout = timeout
		return nil
	}
}

func WithReportWriter(w io.Writer) auditorOption {
	return func(options *auditorOptions) error {
		if w == nil {
			return fmt.Errorf("report writer cannot be nil")
		}
		options.reportWriter = w
		return nil
	}
}

func WithDetailedReports(detailedReports bool) auditorOption {
	return func(options *auditorOptions) error {
		options.detailedReports = detailedReports
		return nil
	}
}
