// Copyright 2020 FairwindsOps Inc
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

package cmd

import (
	"errors"
	"os"

	"github.com/fairwindsops/polaris/pkg/fix"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	filesPath   string
	checksToFix []string
	isTemplate  bool
)

func init() {
	rootCmd.AddCommand(fixCommand)
	fixCommand.PersistentFlags().StringVar(&filesPath, "files-path", "", "mutate and fix one or more YAML files in a specified folder")
	fixCommand.PersistentFlags().BoolVar(&isTemplate, "template", false, "set to true when modifyng a YAML template, like a Helm chart (experimental)")
	fixCommand.PersistentFlags().StringSliceVar(&checksToFix, "checks", []string{}, "Optional flag to specify specific checks to fix eg. checks=hostIPCSet,hostPIDSet and checks=all applies fix to all defined checks mutations")
}

var fixCommand = &cobra.Command{
	Use:   "fix",
	Short: "Fix Infrastructure as code files.",
	Long:  `Fix Infrastructure as code files.`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Debug("Setting up controller manager")

		err := fix.Execute(config, filesPath, isTemplate, checksToFix...)
		if err != nil {

			if errors.Is(err, fix.ErrFilesPathRequired) {
				logrus.Error("Please specify a files-path flag")
				cmd.Help()
				os.Exit(1)
			}

			logrus.Fatal(err)
		}
	},
}
