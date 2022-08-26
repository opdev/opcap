package cmd

import (
	"github.com/opdev/opcap/internal/upload"

	"github.com/gobuffalo/envy"
	"github.com/spf13/cobra"
)

var osversion string

var uploadflags upload.UploadOptions

// uploadCmd is used to upload objects to an S3 compatible backend using the MinIO client
func uploadCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "upload",
		Short: "Upload audit logs to an S3 compatible storage service.",
		Long:  `Upload audit logs to an S3 compatible storage service.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return upload.Upload(cmd.Parent().Context(), uploadflags, checkflags.CatalogSource, checkflags.CatalogSourceNamespace)
		},
	}

	flags := cmd.Flags()

	flags.StringVar(&uploadflags.Bucket, "bucket", envy.Get("S3_BUCKET", ""),
		"s3 bucket where result will be stored")
	flags.StringVar(&uploadflags.Path, "path", envy.Get("S3_PATH", ""),
		"s3 path where result will be stored")
	flags.StringVar(&uploadflags.Endpoint, "endpoint", envy.Get("S3_ENDPOINT", ""),
		"s3 endpoint where bucket will be created")
	flags.StringVar(&uploadflags.AccessKeyID, "accesskeyid", envy.Get("S3_ACCESS_KEY_ID", ""),
		"s3 access key id for authentication")
	flags.StringVar(&uploadflags.SecretAccessKey, "secretaccesskey", envy.Get("S3_SECRET_ACCESS_KEY", ""),
		"s3 secret access key for authentication")
	flags.StringVar(&uploadflags.UseSSL, "usessl", envy.Get("S3_USESSL", "false"),
		"when used s3 backend is expected to be accessible via https; false by default")
	flags.StringVar(&uploadflags.Trace, "trace", envy.Get("TRACE", "false"),
		"enable tracing; false by default")

	return &cmd
}
