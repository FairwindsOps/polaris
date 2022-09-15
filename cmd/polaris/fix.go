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
	"strings"

	"github.com/fairwindsops/polaris/pkg/kube"
	"github.com/fairwindsops/polaris/pkg/mutation"
	"github.com/fairwindsops/polaris/pkg/validator"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const templateLineMarker = "# POLARIS_FIX_TMPL"
const templateOpenMarker = "POLARIS_OPEN_TMPL"
const templateCloseMarker = "POLARIS_CLOSE_TMPL"

var (
	filesPath   string
	checksToFix []string
	fixAll      bool
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

		if filesPath == "" {
			logrus.Error("Please specify a files-path flag")
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

		if len(checksToFix) > 0 {
			if len(checksToFix) == 1 && checksToFix[0] == "all" {
				allchecks := []string{}
				for key := range config.Checks {
					allchecks = append(allchecks, key)
				}
				config.Mutations = allchecks
			} else if len(checksToFix) == 0 && checksToFix[0] == "none" {
				config.Mutations = nil
			} else {
				config.Mutations = checksToFix
			}
		}

		for _, fullFilePath := range yamlFiles {
			yamlContent, err := ioutil.ReadFile(fullFilePath)
			if err != nil {
				logrus.Fatalf("Error reading file with file path %s: %v", fullFilePath, err)
			}

			if err != nil {
				logrus.Fatalf("Error marshalling %s: %v", fullFilePath, err)
			}

			if isTemplate {
				yamlContent = []byte(detemplate(string(yamlContent)))
			}
			kubeResources := kube.CreateResourceProviderFromYaml(string(yamlContent))
			results, err := validator.ApplyAllSchemaChecksToResourceProvider(&config, kubeResources)
			if err != nil {
				logrus.Fatalf("Error applying schema check to the resources %s: %v", fullFilePath, err)
			}
			allMutations := mutation.GetMutationsFromResults(results)

			updatedYamlContent := ""
			if len(allMutations) > 0 {
				for _, resources := range kubeResources.Resources {
					for _, resource := range resources {
						key := fmt.Sprintf("%s/%s/%s", resource.Kind, resource.Resource.GetName(), resource.Resource.GetNamespace())
						mutations := allMutations[key]
						mutatedYamlContent, err := mutation.ApplyAllMutations(string(resource.OriginalObjectYAML), mutations)
						if err != nil {
							logrus.Errorf("Error applying schema mutations to the resource %s: %v", key, err)
							os.Exit(1)
						}
						if updatedYamlContent != "" {
							updatedYamlContent += "\n---\n"
						}
						updatedYamlContent += mutatedYamlContent
					}
				}
			}

			if isTemplate {
				updatedYamlContent = retemplate(updatedYamlContent)
			}

			if updatedYamlContent != "" {
				err = ioutil.WriteFile(fullFilePath, []byte(updatedYamlContent), 0644)
				if err != nil {
					logrus.Fatalf("Error writing output to file: %v", err)
				}
			}
		}

	},
}

func detemplate(content string) string {
	lines := strings.Split(content, "\n")
	for idx, line := range lines {
		lines[idx] = detemplateLine(line)
	}
	return strings.Join(lines, "\n")
}

func retemplate(content string) string {
	lines := strings.Split(content, "\n")
	for idx, line := range lines {
		lines[idx] = retemplateLine(line)
	}
	return strings.Join(lines, "\n")
}

func detemplateLine(line string) string {
	if !strings.HasPrefix(strings.TrimSpace(line), "{{") {
		line = strings.ReplaceAll(line, "{", templateOpenMarker)
		line = strings.ReplaceAll(line, "}", templateCloseMarker)
		return line
	}
	tmplStart := strings.Index(line, "{{")
	newLine := line[:tmplStart] + templateLineMarker + line[tmplStart:]
	return newLine
}

func retemplateLine(line string) string {
	if !strings.Contains(line, templateLineMarker) {
		line = strings.ReplaceAll(line, templateOpenMarker, "{")
		line = strings.ReplaceAll(line, templateCloseMarker, "}")
		return line
	}
	return strings.Replace(line, templateLineMarker, "", 1)
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
