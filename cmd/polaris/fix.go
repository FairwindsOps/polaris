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
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/fairwindsops/polaris/pkg/kube"
	"github.com/fairwindsops/polaris/pkg/mutation"
	"github.com/fairwindsops/polaris/pkg/validator"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	yamlV3 "gopkg.in/yaml.v3"
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
			logrus.Error("Please specify a file-path flag")
			cmd.Help()
			os.Exit(1)
		}
		var yamlFiles []string
		fileInfo, err := os.Stat(filesPath)
		if err != nil {
			logrus.Error(err)
			os.Exit(1)
		}
		if fileInfo.IsDir() {
			baseDir := filesPath
			if !strings.HasSuffix(filesPath, "/") {
				baseDir = baseDir + "/"
			}
			yamlFiles, err = getYamlFiles(baseDir)
			if err != nil {
				logrus.Error(err)
				os.Exit(1)
			}
		} else {
			yamlFiles = append(yamlFiles, filesPath)
		}
		var contentStr string
		isFirstResource := true
		for _, fullFilePath := range yamlFiles {

			yamlFile, err := ioutil.ReadFile(fullFilePath)
			if err != nil {
				logrus.Errorf("Error reading file with file path %s: %v", fullFilePath, err)
				os.Exit(1)
			}

			dec := yamlV3.NewDecoder(bytes.NewReader(yamlFile))

			for {
				data := map[string]interface{}{}
				err := dec.Decode(&data)
				// check it was parsed
				if data == nil {
					continue
				}
				// break the loop in case of EOF
				if errors.Is(err, io.EOF) {
					break
				}
				if err != nil {
					logrus.Errorf("Error decoding data for file with file path %s: %v", fullFilePath, err)
					os.Exit(1)
				}
				yamlContent, err := yamlV3.Marshal(data)
				if err != nil {
					logrus.Errorf("Error marshalling %s: %v", fullFilePath, err)
					os.Exit(1)
				}
				kubeResources := kube.CreateResourceProviderFromYaml(string(yamlContent))
				results, err := validator.ApplyAllSchemaChecksToResourceProvider(&config, kubeResources)
				if err != nil {
					logrus.Errorf("Error applying schema check to the resources %s: %v", fullFilePath, err)
					os.Exit(1)
				}
				comments, allMutations := mutation.GetMutationsAndCommentsFromResults(results)
				updatedYamlContent := string(yamlContent)
				if len(allMutations) > 0 {
					for _, resources := range kubeResources.Resources {
						key := fmt.Sprintf("%s/%s/%s", resources[0].Kind, resources[0].Resource.GetName(), resources[0].Resource.GetNamespace())
						mutations := allMutations[key]
						mutated, err := mutation.ApplyAllSchemaMutations(&config, kubeResources, resources[0], mutations)
						if err != nil {
							logrus.Errorf("Error applying schema mutations to the resources: %v", err)
							os.Exit(1)
						}
						mutatedYamlContent, err := yaml.JSONToYAML(mutated.OriginalObjectJSON)
						if err != nil {
							logrus.Errorf("Error converting JSON to Yaml : %v", err)
							os.Exit(1)
						}
						updatedYamlContent = mutation.UpdateMutatedContentWithComments(string(mutatedYamlContent), comments)
					}
				}
				if isFirstResource {
					contentStr = updatedYamlContent
					isFirstResource = false
				} else {
					contentStr += "\n"
					contentStr += "---"
					contentStr += "\n"
					contentStr += updatedYamlContent
				}
			}

			if contentStr != "" {
				err = ioutil.WriteFile(fullFilePath, []byte(contentStr), 0644)
				if err != nil {
					logrus.Errorf("Error writing output to file: %v", err)
					os.Exit(1)
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
		if filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml" {
			list = append(list, path)
		}
		return nil
	})
	return list, err
}
