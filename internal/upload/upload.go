package upload

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/opdev/opcap/internal/operator"
)

// constants for time/date formatting
const (
	// YYYYMMDD: 20220323
	YYYYMMDD = "20060102"
	// 24h hhmmss: 142320
	HHMMSS24h = "150405"
)

type UploadOptions struct {
	Bucket          string `json:"bucket"`
	Path            string `json:"path"`
	Endpoint        string `json:"endpoint"`
	AccessKeyID     string `json:"accesskeyid"`
	SecretAccessKey string `json:"secretaccesskey"`
	UseSSL          string `json:"usessl"`
	Trace           string `json:"trace"`
}

type auditReport struct {
	Catalog          string  `json:"catalog"`
	CatalogNamespace string  `json:"catalognamespace"`
	OpenShiftVersion string  `json:"osversion"`
	Audits           []audit `json:"audits"`
}

type audit struct {
	Message     string `json:"message"`
	Package     string `json:"package,omitempty"`
	Channel     string `json:"channel,omitempty"`
	InstallMode string `json:"installmode,omitempty"`
}

func initClient() (operator.Client, error) {
	opClient, err := operator.NewOpCapClient()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OpenShift client: %v", err)
	}
	return opClient, nil
}

func Upload(ctx context.Context, options UploadOptions, catalogSource, catalogSourceNamespace string) error {
	client, err := initClient()
	if err != nil {
		return err
	}
	osversion, err := client.GetOpenShiftVersion()
	if err != nil {
		return fmt.Errorf("failed to connect to OpenShift: %v", err)
	}

	// Convert uploadflags.UseSSL and Trace to bool
	useSSL, _ := strconv.ParseBool(options.UseSSL)
	trace, _ := strconv.ParseBool(options.Trace)

	// Create a minio client to interact with minio object store
	minioClient, err := minio.New(options.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(options.AccessKeyID, options.SecretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return err
	}

	if trace {
		minioClient.TraceOn(os.Stdout)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// check for bucket, create if it does not exist
	now := time.Now()
	if options.Bucket == "" {
		options.Bucket = now.Format(YYYYMMDD)
	}
	prefix := now.Format(HHMMSS24h)

	if ok, _ := minioClient.BucketExists(ctx, options.Bucket); !ok {
		minioClient.MakeBucket(ctx, options.Bucket, minio.MakeBucketOptions{})
	}

	var report auditReport

	report.OpenShiftVersion = osversion
	report.Catalog = catalogSource
	report.CatalogNamespace = catalogSourceNamespace

	// for each line is stdout.json which is provided by opcap create an Audit object and add to the rawreport Audits field.
	f, err := os.Open("operator_install_report.json")
	if err != nil {
		return err
	}
	s := bufio.NewScanner(f)
	for s.Scan() {
		var audit audit
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

	if options.Path == "" {
		options.Path = osversion + "/" + prefix + "_report.json"
	}
	_, err = minioClient.FPutObject(ctx, options.Bucket, options.Path, "report.json", minio.PutObjectOptions{ContentType: "application/json"})
	if err != nil {
		return err
	}

	return nil
}
