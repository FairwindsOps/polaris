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
	"os"
	"strings"

	conf "github.com/fairwindsops/polaris/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	mergeConfig                  bool
	configPath                   string
	disallowExemptions           bool
	disallowConfigExemptions     bool
	disallowAnnotationExemptions bool
	fixChecks                    bool
	logLevel                     string
	auditPath                    string
	displayName                  string
	kubeContext                  string
	insightsHost                 string
)

var (
	version string
)

func init() {
	// Flags
	rootCmd.PersistentFlags().BoolVarP(&mergeConfig, "merge-config", "m", false, "If true, custom configuration will be merged with default configuration instead of replacing it.")
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Location of Polaris configuration file.")
	rootCmd.PersistentFlags().StringVarP(&kubeContext, "context", "x", "", "Set the kube context.")
	rootCmd.PersistentFlags().BoolVarP(&disallowExemptions, "disallow-exemptions", "", false, "Disallow any configured exemption.")
	rootCmd.PersistentFlags().BoolVarP(&disallowConfigExemptions, "disallow-config-exemptions", "", false, "Disallow exemptions set within the configuration file.")
	rootCmd.PersistentFlags().BoolVarP(&disallowAnnotationExemptions, "disallow-annotation-exemptions", "", false, "Disallow any exemption defined as a controller annotation.")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "", logrus.InfoLevel.String(), "Logrus log level to be output (trace, debug, info, warning, error, fatal, panic).")
	rootCmd.PersistentFlags().StringVar(&insightsHost, "insights-host", "https://insights.fairwinds.com", "Fairwinds Insights host URL")
}

var config conf.Configuration

var rootCmd = &cobra.Command{
	Use:   "polaris",
	Short: "polaris",
	Long:  `Validation of best practices in your Kubernetes clusters.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		parsedLevel, err := logrus.ParseLevel(logLevel)
		if err != nil {
			logrus.Errorf("log-level flag has invalid value %s", logLevel)
		} else {
			logrus.SetLevel(parsedLevel)
		}

		config, err = conf.MergeConfigAndParseFile(configPath, mergeConfig)
		if err != nil {
			logrus.Errorf("Error parsing config at %s: %v", configPath, err)
			os.Exit(1)
		}

		config.DisallowExemptions = disallowExemptions
		config.DisallowConfigExemptions = disallowConfigExemptions
		config.DisallowAnnotationExemptions = disallowAnnotationExemptions
		config.KubeContext = kubeContext
	},
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Error("You must specify a sub-command.")
		err := cmd.Help()
		if err != nil {
			logrus.Error(err)
		}
		os.Exit(1)
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if !strings.HasPrefix(cmd.Use, "audit") {
			os.Stderr.WriteString("\n\nWant more? Automate Polaris for free with Fairwinds Insights!\n🚀 https://fairwinds.com/insights-signup/polaris 🚀 \n")
		}
	},
}

// Execute the stuff
func Execute(VERSION string) {
	version = VERSION
	if err := rootCmd.Execute(); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
}
