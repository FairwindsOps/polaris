// Copyright 2019 FairwindsOps Inc
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

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	
	conf "github.com/fairwindsops/polaris/pkg/config"
	"github.com/fairwindsops/polaris/pkg/kube"
	"github.com/fairwindsops/polaris/pkg/validator"
	"github.com/fairwindsops/polaris/cmd/polaris"
	"github.com/sirupsen/logrus"
	_ "k8s.io/client-go/plugin/pkg/client/auth" // Required for other auth providers like GKE.
	"sigs.k8s.io/yaml"

)

const (
	// Version represents the current release version of Polaris
	Version = "0.6.0"
	// TODO make sure this gets set.
	commit = "n/a"
)

func main() {
	// Load CLI Flags
	// TODO: Split up global flags vs dashboard/webhook/audit specific flags
	audit := flag.Bool("audit", false, "Runs a one-time audit.")
	auditPath := flag.String("audit-path", "", "If specified, audits one or more YAML files instead of a cluster")
	setExitCode := flag.Bool("set-exit-code-on-error", false, "When running with --audit, set an exit code of 3 when the audit contains error-level issues.")
	minScore := flag.Int("set-exit-code-below-score", 0, "When running with --audit, set an exit code of 4 when the score is below this threshold (1-100)")
	auditOutputURL := flag.String("output-url", "", "Destination URL to send audit results")
	auditOutputFile := flag.String("output-file", "", "Destination file for audit results")
	auditOutputFormat := flag.String("output-format", "json", "Output format for results - json, yaml, or score")
	displayName := flag.String("display-name", "", "An optional identifier for the audit")
	
	
	
	
	//configPath := flag.String("config", "", "Location of Polaris configuration file")
	//disallowExemptions := flag.Bool("disallow-exemptions", false, "Location of Polaris configuration file")
	//logLevel := flag.String("log-level", logrus.InfoLevel.String(), "Logrus log level")
	
	flag.Parse()

	c := conf.Configuration{}

	if *displayName != "" {
		c.DisplayName = *displayName
	}

	if *audit {
		auditData := runAndReportAudit(c, *auditPath, *auditOutputFile, *auditOutputURL, *auditOutputFormat)

		summary := auditData.GetSummary()
		score := summary.GetScore()
		if *setExitCode && summary.Errors > 0 {
			logrus.Infof("%d errors found in audit", summary.Errors)
			os.Exit(3)
		} else if *minScore != 0 && score < uint(*minScore) {
			logrus.Infof("Audit score of %d is less than the provided minimum of %d", score, *minScore)
			os.Exit(4)
		}
	} else {
		cmd.Execute(Version, commit)
	}
}

func startDashboardServer(c conf.Configuration, auditPath string, loadAuditFile string, port int, basePath string) {
}

func startWebhookServer(c conf.Configuration, disableWebhookConfigInstaller bool, port int) {
	
}

func runAndReportAudit(c conf.Configuration, auditPath string, outputFile string, outputURL string, outputFormat string) validator.AuditData {
	// Create a kubernetes client resource provider
	k, err := kube.CreateResourceProvider(auditPath)
	if err != nil {
		logrus.Errorf("Error fetching Kubernetes resources %v", err)
		os.Exit(1)
	}
	auditData, err := validator.RunAudit(c, k)

	if err != nil {
		logrus.Errorf("Error while running audit on resources: %v", err)
		os.Exit(1)
	}

	var outputBytes []byte
	if outputFormat == "score" {
		outputBytes = []byte(fmt.Sprintf("%d\n", auditData.GetSummary().GetScore()))
	} else if outputFormat == "yaml" {
		jsonBytes, err := json.Marshal(auditData)
		if err == nil {
			outputBytes, err = yaml.JSONToYAML(jsonBytes)
		}
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
	return auditData
}
