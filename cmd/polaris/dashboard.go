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
	"net/http"

	"github.com/spf13/cobra"
	"github.com/sirupsen/logrus"
	"github.com/fairwindsops/polaris/pkg/dashboard"
	"github.com/fairwindsops/polaris/pkg/validator"
)

var serverPort int
var basePath string
var loadAuditFile string


func init() {
	rootCmd.AddCommand(dashboardCmd)
	dashboardCmd.PersistentFlags().IntVarP(&serverPort, "port", "p", 8080, "Port for the dashboard webserver.")
	dashboardCmd.PersistentFlags().StringVar(&basePath, "base-path", "/", "Path on which the dashboard is served.")
	dashboardCmd.PersistentFlags().StringVar(&loadAuditFile, "load-audit-file", "", "Runs the dashboard with data saved from a past audit.")
	dashboardCmd.PersistentFlags().StringVar(&auditPath, "audit-path", "", "If specified, audits one or more YAML files instead of a cluster.")
	dashboardCmd.PersistentFlags().StringVar(&displayName, "display-name", "", "An optional identifier for the audit.")

}

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Runs the webserver for Polaris dashboard.",
	Long:  `Runs the webserver for Polaris dashboard.`,
	Run: func(cmd *cobra.Command, args []string) {
		if displayName != "" {
			config.DisplayName = displayName
		}
	
		var auditDataPtr *validator.AuditData
		if loadAuditFile != "" {
			auditData := validator.ReadAuditFromFile(loadAuditFile)
			auditDataPtr = &auditData
		}
		router := dashboard.GetRouter(c, auditPath, serverPort, basePath, auditDataPtr)
		router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("OK"))
		})
		http.Handle("/", router)
	
		logrus.Infof("Starting Polaris dashboard server on port %d", serverPort)
		logrus.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", serverPort), nil))
	},
}
