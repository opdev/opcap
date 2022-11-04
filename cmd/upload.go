package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/opdev/opcap/internal/operator"

	"github.com/gobuffalo/envy"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var osversion string

type uploadCommandFlags struct {
	Bucket          string `json:"bucket"`
	Path            string `json:"path"`
	Endpoint        string `json:"endpoint"`
	AccessKeyID     string `json:"accesskeyid"`
	SecretAccessKey string `json:"secretaccesskey"`
	UseSSL          string `json:"usessl"`
	Trace           string `json:"trace"`
	LogLevel        string `json:"loglevel"`
}

type Report struct {
	Catalog          string  `json:"catalog"`
	CatalogNamespace string  `json:"catalognamespace"`
	OpenShiftVersion string  `json:"osversion"`
	Audits           []Audit `json:"audits"`
}

type Audit struct {
	Message     string `json:"message"`
	Package     string `json:"package,omitempty"`
	Channel     string `json:"channel,omitempty"`
	InstallMode string `json:"installmode,omitempty"`
}

// constants for time/date formatting
const (
	// YYYYMMDD: 20220323
	YYYYMMDD = "20060102"
	// 24h hhmmss: 142320
	HHMMSS24h = "150405"
)

var uploadflags uploadCommandFlags

// uploadCmd is used to upload objects to an S3 compatible backend using the MinIO client
func uploadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "upload",
		Short:   "Upload audit logs to an S3 compatible storage service.",
		Long:    `Upload audit logs to an S3 compatible storage service.`,
		PreRunE: uploadPreRunE,
		RunE:    uploadRunE,
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

	return cmd
}

func uploadPreRunE(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	kubeconfig, err := kubeConfig()
	if err != nil {
		return fmt.Errorf("could not get kubeconfig: %v", err)
	}

	opClient, err := operator.NewOpCapClient(kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to initialize OpenShift client: %v", err)
	}

	osversion, err = opClient.GetOpenShiftVersion(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to connect to OpenShift: %v", err)
	}

	return nil
}

func uploadRunE(cmd *cobra.Command, args []string) error {
	// Convert uploadflags.UseSSL and Trace to bool
	usessl, _ := strconv.ParseBool(uploadflags.UseSSL)
	trace, _ := strconv.ParseBool(uploadflags.Trace)

	// Create a minio client to interact with minio object store
	minioClient, err := minio.New(uploadflags.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(uploadflags.AccessKeyID, uploadflags.SecretAccessKey, ""),
		Secure: usessl,
	})
	if err != nil {
		return err
	}
	if trace {
		minioClient.TraceOn(os.Stdout)
	}

	fs := afero.NewOsFs()

	return upload(cmd.Context(), uploadflags, minioClient, fs, osversion)
}

type minioClient interface {
	BucketExists(ctx context.Context, bucket string) (bool, error)
	MakeBucket(ctx context.Context, bucket string, opts minio.MakeBucketOptions) error
	FPutObject(ctx context.Context, bucket, path, file string, opts minio.PutObjectOptions) (minio.UploadInfo, error)
}

func loadAudits(ctx context.Context, fs afero.Fs, filename string) ([]Audit, error) {
	// Initialize audits array with a capacity of 10 and a length of 0
	audits := make([]Audit, 0, 10)
	f, err := fs.Open(filename)
	if err != nil {
		return audits, err
	}
	s := bufio.NewScanner(f)
	for s.Scan() {
		var audit Audit
		if err := json.Unmarshal(s.Bytes(), &audit); err != nil {
			return audits, err
		}
		if audit.Message == "Succeeded" || audit.Message == "failed" || audit.Message == "timeout" {
			audits = append(audits, audit)
		}
	}

	return audits, nil
}

func upload(ctx context.Context, uploadFlags uploadCommandFlags, minioClient minioClient, fs afero.Fs, osversion string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// check for bucket, create if it does not exist
	now := time.Now()
	if uploadFlags.Bucket == "" {
		uploadFlags.Bucket = now.Format(YYYYMMDD)
	}
	prefix := now.Format(HHMMSS24h)

	if ok, _ := minioClient.BucketExists(ctx, uploadFlags.Bucket); !ok {
		minioClient.MakeBucket(ctx, uploadFlags.Bucket, minio.MakeBucketOptions{})
	}

	var report Report

	report.OpenShiftVersion = osversion
	report.Catalog = checkflags.CatalogSource
	report.CatalogNamespace = checkflags.CatalogSourceNamespace

	// for each line is stdout.json which is provided by opcap create an Audit object and add to the rawreport Audits field.
	audits, err := loadAudits(ctx, fs, "operator_install_report.json")
	if err != nil {
		return err
	}
	report.Audits = audits

	data, err := json.Marshal(report)
	if err != nil {
		return err
	}

	if err = afero.WriteFile(fs, "report.json", data, 0o644); err != nil {
		return err
	}

	if uploadFlags.Path == "" {
		uploadFlags.Path = osversion + "/" + prefix + "_report.json"
	}
	_, err = minioClient.FPutObject(ctx, uploadFlags.Bucket, uploadFlags.Path, "report.json", minio.PutObjectOptions{ContentType: "application/json"})
	if err != nil {
		return err
	}

	return nil
}
