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
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	fwebhook "github.com/fairwindsops/polaris/pkg/webhook"
	k8sConfig "sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

var webhookPort int
var disableWebhookConfigInstaller bool
var enableMutations bool
var enableValidations bool

func init() {
	rootCmd.AddCommand(webhookCmd)
	webhookCmd.PersistentFlags().IntVarP(&webhookPort, "port", "p", 9876, "Port for the dashboard webserver.")
	webhookCmd.PersistentFlags().BoolVar(&disableWebhookConfigInstaller, "disable-webhook-config-installer", false, "Disable the installer in the webhook server, so it won't install webhook configuration resources during bootstrapping.")
	webhookCmd.PersistentFlags().BoolVar(&enableValidations, "validate", true, "Enable the validating webhook to reject workloads with issues")
	webhookCmd.PersistentFlags().BoolVar(&enableMutations, "mutate", false, "Enable the mutating webhook to modify workloads with issues")
}

var webhookCmd = &cobra.Command{
	Use:   "webhook",
	Short: "Runs the webhook webserver.",
	Long:  `Runs the webhook webserver.`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Debug("Setting up controller manager")

		mgr, err := manager.New(k8sConfig.GetConfigOrDie(), manager.Options{
			CertDir: "/opt/cert",
			Port:    webhookPort,
		})
		if err != nil {
			logrus.Errorf("Unable to set up overall controller manager: %v", err)
			os.Exit(1)
		}

		_, err = os.Stat("/opt/cert/tls.crt")
		if os.IsNotExist(err) {
			time.Sleep(time.Second * 10)
			panic("Cert does not exist")
		}
		server := mgr.GetWebhookServer()
		server.CertName = "tls.crt"
		server.KeyName = "tls.key"

		if !enableMutations && !enableValidations {
			logrus.Errorf("One of --mutate or --validate must be set to true")
			os.Exit(1)
		}

		if enableValidations {
			fwebhook.NewValidateWebhook(mgr, fwebhook.Validator{Config: config, Client: mgr.GetClient()})
		}
		if enableMutations {
			fwebhook.NewMutateWebhook(mgr, fwebhook.Mutator{Config: config, Client: mgr.GetClient()})
		}
		logrus.Infof("Polaris webhook server listening on port %d", webhookPort)
		if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
			logrus.Errorf("Error starting manager: %v", err)
			os.Exit(1)
		}
	},
}
