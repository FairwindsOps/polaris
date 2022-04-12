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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"github.com/fairwindsops/polaris/pkg/kube"
	"github.com/fairwindsops/polaris/pkg/mutation"
	"github.com/fairwindsops/polaris/pkg/validator"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var filesPath string

func init() {
	rootCmd.AddCommand(fixCommand)
	fixCommand.PersistentFlags().StringVar(&filesPath, "files-path", "", "If specified, mutate and fix one or more YAML files.")
}

var fixCommand = &cobra.Command{
	Use:   "fix",
	Short: "Fix Infrastructure as code files.",
	Long:  `Fix Infrastructure as code files.`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Debug("Setting up controller manager")

		if filesPath == "" {
			_, filesPath, _, _ = runtime.Caller(0)
		}

		baseDir := filepath.Dir(filesPath)
		dirs, err := ioutil.ReadDir(baseDir)
		if err != nil {
			panic(err)
		}

		for _, dir := range dirs {
			check := dir.Name()
			checkDir := baseDir + "/" + check
			caseFiles, err := ioutil.ReadDir(checkDir)
			if err != nil {
				panic(err)
			}

			for _, caseFile := range caseFiles {
				fullFilePath := checkDir + "/" + caseFile.Name()
				kubeResources, err := kube.CreateResourceProviderFromPath(fullFilePath)
				if err != nil {
					panic(err)
				}
				results, err := validator.ApplyAllSchemaChecksToResourceProvider(&config, kubeResources)
				if err != nil {
					panic(err)
				}
				allMutations := mutation.GetMutationsFromResults(&config, results)

				for _, resources := range kubeResources.Resources {
					key := fmt.Sprintf("%s/%s/%s", resources[0].Kind, resources[0].Resource.GetName(), resources[0].Resource.GetNamespace())
					mutations := allMutations[key]
					mutated, err := mutation.ApplyAllSchemaMutations(&config, kubeResources, resources[0], mutations)
					if err != nil {
						panic(err)
					}
					err = ioutil.WriteFile(fullFilePath, mutated.OriginalObjectJSON, 0644)
					if err != nil {
						logrus.Errorf("Error writing output to file: %v", err)
						os.Exit(1)
					}
				}
			}
		}

	},
}
