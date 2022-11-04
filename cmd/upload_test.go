package cmd

import (
	"context"
	"strings"

	"github.com/minio/minio-go/v7"
	. "github.com/onsi/ginkgo/v2/dsl/core"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var _ minioClient = fakeMinioClient{}

type fakeMinioClient struct{}

func (f fakeMinioClient) BucketExists(ctx context.Context, bucket string) (bool, error) {
	return false, nil
}

func (f fakeMinioClient) FPutObject(ctx context.Context, bucket, path, file string, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	return minio.UploadInfo{}, nil
}

func (f fakeMinioClient) MakeBucket(ctx context.Context, bucket string, opts minio.MakeBucketOptions) error {
	return nil
}

var _ = Describe("Upload tests", func() {
	When("creating upload cmd", func() {
		It("should contain flags", func() {
			cmd := uploadCmd()
			Expect(cmd.HasFlags()).To(BeTrue())
		})
	})

	When("uploading", func() {
		It("should succeed", func() {
			report := `{"Message":"Succeeded"}`
			afs := afero.NewMemMapFs()
			afero.WriteReader(afs, "operator_install_report.json", strings.NewReader(report))
			Expect(upload(context.TODO(), uploadCommandFlags{}, fakeMinioClient{}, afs, "4.11")).To(Succeed())
		})
	})
})
