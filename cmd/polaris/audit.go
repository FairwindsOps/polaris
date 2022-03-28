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
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"

	"github.com/fairwindsops/polaris/pkg/kube"
	"github.com/fairwindsops/polaris/pkg/validator"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

var setExitCode bool
var onlyShowFailedTests bool
var minScore int
var auditOutputURL string
var auditOutputFile string
var auditOutputFormat string
var resourceToAudit string
var useColor bool
var helmChart string
var helmValues string
var checks string

func init() {
	rootCmd.AddCommand(auditCmd)
	auditCmd.PersistentFlags().StringVar(&auditPath, "audit-path", "", "If specified, audits one or more YAML files instead of a cluster.")
	auditCmd.PersistentFlags().BoolVar(&setExitCode, "set-exit-code-on-danger", false, "Set an exit code of 3 when the audit contains danger-level issues.")
	auditCmd.PersistentFlags().BoolVar(&onlyShowFailedTests, "only-show-failed-tests", false, "If specified, audit output will only show failed tests.")
	auditCmd.PersistentFlags().IntVar(&minScore, "set-exit-code-below-score", 0, "Set an exit code of 4 when the score is below this threshold (1-100).")
	auditCmd.PersistentFlags().StringVar(&auditOutputURL, "output-url", "", "Destination URL to send audit results.")
	auditCmd.PersistentFlags().StringVar(&auditOutputFile, "output-file", "", "Destination file for audit results.")
	auditCmd.PersistentFlags().StringVarP(&auditOutputFormat, "format", "f", "json", "Output format for results - json, yaml, pretty, or score.")
	auditCmd.PersistentFlags().BoolVar(&useColor, "color", true, "Whether to use color in pretty format.")
	auditCmd.PersistentFlags().StringVar(&displayName, "display-name", "", "An optional identifier for the audit.")
	auditCmd.PersistentFlags().StringVar(&resourceToAudit, "resource", "", "Audit a specific resource, in the format namespace/kind/version/name, e.g. nginx-ingress/Deployment.apps/v1/default-backend.")
	auditCmd.PersistentFlags().StringVar(&helmChart, "helm-chart", "", "Will fill out Helm template")
	auditCmd.PersistentFlags().StringVar(&helmValues, "helm-values", "", "Optional flag to add helm values")
	auditCmd.PersistentFlags().StringVar(&checks, "checks", "", "Optional flag to specify specific checks to check")
}

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Runs a one-time audit.",
	Long:  `Runs a one-time audit.`,
	Run: func(cmd *cobra.Command, args []string) {
		if displayName != "" {
			config.DisplayName = displayName
		}
		if helmChart != "" {
			var err error
			auditPath, err = ProcessHelmTemplates(helmChart, helmValues)
			if err != nil {
				logrus.Infof("Couldn't process helm chart: %v", err)
				os.Exit(1)
			}
		}

		k, err := kube.CreateResourceProvider(context.TODO(), auditPath, resourceToAudit, config)
		if err != nil {
			logrus.Errorf("Error fetching Kubernetes resources %v", err)
			os.Exit(1)
		}

		auditData, err := validator.RunAudit(config, k)
		if err != nil {
			logrus.Errorf("Error while running audit on resources: %v", err)
			os.Exit(1)
		}

		outputAudit(auditData, auditOutputFile, auditOutputURL, auditOutputFormat, useColor, onlyShowFailedTests)

		summary := auditData.GetSummary()
		score := summary.GetScore()
		if setExitCode && summary.Dangers > 0 {
			logrus.Infof("%d danger items found in audit", summary.Dangers)
			os.Exit(3)
		} else if minScore != 0 && score < uint(minScore) {
			logrus.Infof("Audit score of %d is less than the provided minimum of %d", score, minScore)
			os.Exit(4)
		}
	},
}

// ProcessHelmTemplates turns helm into yaml to be processed by Polaris or the other tools.
func ProcessHelmTemplates(helmChart, helmValues string) (string, error) {
	cmd := exec.Command("helm", "dependency", "update", helmChart)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logrus.Error(string(output))
		return "", err
	}

	dir, err := ioutil.TempDir("", "*")
	if err != nil {
		return "", err
	}
	params := []string{
		"template", helmChart,
		helmChart,
		"--output-dir",
		dir,
	}
	if helmValues != "" {
		params = append(params, "--values", helmValues)
	}

	cmd = exec.Command("helm", params...)
	output, err = cmd.CombinedOutput()

	if err != nil {
		logrus.Error(string(output))
		return "", err
	}
	return dir, nil
}

func outputAudit(auditData validator.AuditData, outputFile, outputURL, outputFormat string, useColor bool, onlyShowFailedTests bool) {
	if onlyShowFailedTests {
		auditData = auditData.RemoveSuccessfulResults()
	}
	var outputBytes []byte
	var err error
	if outputFormat == "score" {
		outputBytes = []byte(fmt.Sprintf("%d\n", auditData.GetSummary().GetScore()))
	} else if outputFormat == "yaml" {
		var jsonBytes []byte
		jsonBytes, err = json.Marshal(auditData)
		if err == nil {
			outputBytes, err = yaml.JSONToYAML(jsonBytes)
		}
	} else if outputFormat == "pretty" {
		outputBytes = []byte(auditData.GetPrettyOutput(useColor))
	} else {
		outputBytes, err = json.MarshalIndent(auditData, "", "  ")
	}
	if err != nil {
		logrus.Errorf("Error marshalling audit: %v", err)
		os.Exit(1)
	}
	if outputURL == "" && outputFile == "" {
		os.Stdout.Write(outputBytes)
	} else {
		if outputURL != "" {
			req, err := http.NewRequest("POST", outputURL, bytes.NewBuffer(outputBytes))

			if err != nil {
				logrus.Errorf("Error building request for output: %v", err)
				os.Exit(1)
			}

			if outputFormat == "json" {
				req.Header.Set("Content-Type", "application/json")
			} else if outputFormat == "yaml" {
				req.Header.Set("Content-Type", "application/x-yaml")
			} else {
				req.Header.Set("Content-Type", "text/plain")
			}
			client := &http.Client{}
			resp, err := client.Do(req)

			if err != nil {
				logrus.Errorf("Error making request for output: %v", err)
				os.Exit(1)
			}

			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)

			if err != nil {
				logrus.Errorf("Error reading response: %v", err)
				os.Exit(1)
			}

			logrus.Infof("Received response: %v", body)
		}

		if outputFile != "" {
			err := ioutil.WriteFile(outputFile, []byte(outputBytes), 0644)
			if err != nil {
				logrus.Errorf("Error writing output to file: %v", err)
				os.Exit(1)
			}
		}
	}
}
