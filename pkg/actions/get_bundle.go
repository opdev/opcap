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

package actions

import (
	"os/exec"

	// "strings"

	log "github.com/sirupsen/logrus"

	"opcap/pkg"
)

// Manifest define the manifest.json which is  required to read the bundle
type Manifest struct {
	Config string
	Layers []string
}

// GetDataFromBundleImage returns the bundle from the image

func DownloadImage(image string, containerEngine string) error {
	log.Infof("Downloading image %s to audit...", image)
	cmd := exec.Command(containerEngine, "pull", image)
	_, err := pkg.RunCommand(cmd)
	// if found an error try again
	// Sometimes it faces issues to download the image
	if err != nil {
		log.Warnf("error %s faced to downlad the image. Let's try more one time.", err)
		cmd := exec.Command(containerEngine, "pull", image)
		_, err = pkg.RunCommand(cmd)
	}
	return err
}
