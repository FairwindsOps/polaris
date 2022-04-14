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

	"github.com/fairwindsops/polaris/pkg/kube"
	"github.com/fairwindsops/polaris/pkg/mutation"
	"github.com/fairwindsops/polaris/pkg/validator"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

var filesPath string

func init() {
	rootCmd.AddCommand(fixCommand)
	fixCommand.PersistentFlags().StringVar(&filesPath, "files-path", "", "mutate and fix one or more YAML files in a specified folder")
}

var fixCommand = &cobra.Command{
	Use:   "fix",
	Short: "Fix Infrastructure as code files.",
	Long:  `Fix Infrastructure as code files.`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Debug("Setting up controller manager")

		if filesPath == "" {
			logrus.Error("Please specify a sub-command.")
			err := cmd.Help()
			panic(err)
		}

		baseDir := filesPath + "/"
		yamlFiles, err := getYamlFiles(baseDir)
		if err != nil {
			panic(err)
		}

		for _, fullFilePath := range yamlFiles {
			kubeResources, err := kube.CreateResourceProviderFromPath(fullFilePath)
			if err != nil {
				panic(err)
			}
			results, err := validator.ApplyAllSchemaChecksToResourceProvider(&config, kubeResources)
			if err != nil {
				panic(err)
			}
			comments, allMutations := mutation.GetMutationsAndCommentsFromResults(results)
			if len(allMutations) > 0 {
				for _, resources := range kubeResources.Resources {
					key := fmt.Sprintf("%s/%s/%s", resources[0].Kind, resources[0].Resource.GetName(), resources[0].Resource.GetNamespace())
					mutations := allMutations[key]
					mutated, err := mutation.ApplyAllSchemaMutations(&config, kubeResources, resources[0], mutations)
					if err != nil {
						panic(err)
					}
					yamlContent, err := yaml.JSONToYAML(mutated.OriginalObjectJSON)
					if err != nil {
						panic(err)
					}
					byteContent := mutation.UpdateMutatedContentWithComments(fullFilePath, string(yamlContent), comments)
					err = ioutil.WriteFile(fullFilePath, byteContent, 0644)
					if err != nil {
						logrus.Errorf("Error writing output to file: %v", err)
						os.Exit(1)
					}
				}
			}
		}

	},
}

func getYamlFiles(rootpath string) ([]string, error) {
	var list []string
	err := filepath.Walk(rootpath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".yaml" {
			list = append(list, path)
		}
		return nil
	})
	return list, err
}
