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

package index

import (
	"capabilities-tool/cmd/index/capabilities"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	indexCmd := &cobra.Command{
		Use:   "index",
		Short: "Install Operator Bundle using operator-sdk",
		Long:  "Install subcommand is used to index an operator bundle using operator-sdk run bundle",
	}

	indexCmd.AddCommand(
		capabilities.NewCmd(),
	)

	return indexCmd

}
