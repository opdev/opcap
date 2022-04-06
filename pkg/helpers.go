// Copyright 2021 The Audit Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pkg

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gobuffalo/envy"
	log "github.com/sirupsen/logrus"
)

const JSON = "json"
const DefaultContainerTool = Docker
const Docker = "docker"
const Podman = "podman"

// PropertiesAnnotation used to Unmarshal the JSON in the CSV annotation
type PropertiesAnnotation struct {
	Type  string
	Value string
}

func (p PropertiesAnnotation) String() string {
	return fmt.Sprintf("{\"type\": \"%s\", \"value\": \"%s\"}", p.Type, p.Value)
}

// Run executes the provided command within this context
func RunCommand(cmd *exec.Cmd) ([]byte, error) {
	command := strings.Join(cmd.Args, " ")
	log.Infof("running: %s\n", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, fmt.Errorf("%s failed with error: (%v) %s", command, err, string(output))
	}
	if len(output) > 0 {
		log.Debugf("command output :%s", output)
	}
	return output, nil
}

func WriteJSON(data []byte, imageName, outputPath, typeName string) error {
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, data, "", "\t")
	if err != nil {
		return err
	}

	path := filepath.Join(outputPath, GetReportName(imageName, typeName, "json"))

	_, err = ioutil.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return ioutil.WriteFile(path, prettyJSON.Bytes(), 0644)
}

func GetReportName(imageName, typeName, typeFile string) string {
	dt := strconv.FormatInt(time.Now().Unix(), 10)
	//dt := time.Now().Format("")

	//prepare image name to use as name of the file
	name := strings.ReplaceAll(imageName, "/", "_")
	name = strings.ReplaceAll(name, ":", "_")
	name = strings.ReplaceAll(name, "-", "_")

	return fmt.Sprintf("%s_%s_%s.%s", typeName, name, dt, typeFile)
}

func GenerateTemporaryDirs() {
	command := exec.Command("rm", "-rf", "tmp")
	_, _ = RunCommand(command)

	command = exec.Command("rm", "-rf", "./output/")
	_, _ = RunCommand(command)

	command = exec.Command("mkdir", "./output/")
	_, err := RunCommand(command)
	if err != nil {
		log.Fatal(err)
	}

	command = exec.Command("mkdir", "tmp")
	_, err = RunCommand(command)
	if err != nil {
		log.Fatal(err)
	}
}

type DockerInspect struct {
	ID           string       `json:"ID"`
	RepoDigests  []string     `json:"RepoDigests"`
	Created      string       `json:"Created"`
	DockerConfig DockerConfig `json:"Config"`
}

type DockerManifestInspect struct {
	ManifestData []ManifestData `json:"manifests"`
}

type ManifestData struct {
	Platform Platform `json:"platform"`
}

type Platform struct {
	Architecture string `json:"architecture"`
	SO           string `json:"so"`
}

type DockerConfig struct {
	Labels map[string]string `json:"Labels"`
}

func WriteDataToS3(filepath string, filename string, bucketname string, endpoint string) error {
	// bucket := "audit-tool-s3-bucket"
	jsonFile, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer jsonFile.Close()
	sess, err := session.NewSession(&aws.Config{
		Endpoint:         aws.String(endpoint),
		Region:           aws.String("us-east-1"),
		Credentials:      credentials.NewStaticCredentials(envy.Get("MINIO_ACCESS_KEY_ID", ""), envy.Get("MINIO_SECRET_ACCESS_KEY", ""), ""),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	})
	uploader := s3manager.NewUploader(sess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucketname),
		Key:    aws.String(filename),
		// here you pass your reader
		// the aws sdk will manage all the memory and file reading for you
		Body: jsonFile,
	})
	return err
}

// GetContainerToolFromEnvVar retrieves the value of the environment variable and defaults to docker when not set
func GetContainerToolFromEnvVar() string {
	if value, ok := os.LookupEnv("CONTAINER_ENGINE"); ok {
		return value
	}
	return DefaultContainerTool
}
