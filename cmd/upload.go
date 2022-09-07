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
	"github.com/spf13/cobra"
)

var osversion string

type UploadCommandFlags struct {
	Bucket          string `json:"bucket"`
	Path            string `json:"path"`
	Endpoint        string `json:"endpoint"`
	AccessKeyID     string `json:"accesskeyid"`
	SecretAccessKey string `json:"secretaccesskey"`
	UseSSL          string `json:"usessl"`
	Trace           string `json:"trace"`
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

var uploadflags UploadCommandFlags

// uploadCmd is used to upload objects to an S3 compatible backend using the MinIO client
var uploadCmd = &cobra.Command{
	Use:     "upload",
	Short:   "Upload audit logs to an S3 compatible storage service.",
	Long:    `Upload audit logs to an S3 compatible storage service.`,
	PreRunE: uploadPreRunE,
	RunE:    uploadRunE,
}

func init() {
	rootCmd.AddCommand(uploadCmd)
	flags := uploadCmd.Flags()

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
}

func uploadPreRunE(cmd *cobra.Command, args []string) error {
	opClient, err := operator.NewOpCapClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to initialize OpenShift client: ", err)
		os.Exit(1)
	}

	osversion, err = opClient.GetOpenShiftVersion(cmd.Context())
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to OpenShift: ", err)
		os.Exit(1)
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

	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	// check for bucket, create if it does not exist
	now := time.Now()
	if uploadflags.Bucket == "" {
		uploadflags.Bucket = now.Format(YYYYMMDD)
	}
	prefix := now.Format(HHMMSS24h)

	if ok, _ := minioClient.BucketExists(ctx, uploadflags.Bucket); !ok {
		minioClient.MakeBucket(ctx, uploadflags.Bucket, minio.MakeBucketOptions{})
	}

	var report Report

	report.OpenShiftVersion = osversion
	report.Catalog = checkflags.CatalogSource
	report.CatalogNamespace = checkflags.CatalogSourceNamespace

	// for each line is stdout.json which is provided by opcap create an Audit object and add to the rawreport Audits field.
	f, err := os.Open("operator_install_report.json")
	if err != nil {
		return err
	}
	s := bufio.NewScanner(f)
	for s.Scan() {
		var audit Audit
		if err := json.Unmarshal(s.Bytes(), &audit); err != nil {
			return err
		}
		if audit.Message == "Succeeded" || audit.Message == "failed" || audit.Message == "timeout" {
			report.Audits = append(report.Audits, audit)
		}
	}
	data, err := json.Marshal(report)
	if err != nil {
		return err
	}

	if err = os.WriteFile("report.json", data, 0o644); err != nil {
		return err
	}

	if uploadflags.Path == "" {
		uploadflags.Path = osversion + "/" + prefix + "_report.json"
	}
	_, err = minioClient.FPutObject(ctx, uploadflags.Bucket, uploadflags.Path, "report.json", minio.PutObjectOptions{ContentType: "application/json"})
	if err != nil {
		return err
	}

	return nil
}
