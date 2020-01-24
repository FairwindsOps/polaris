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
	"os"
	"io/ioutil"
	"strings"

	"github.com/spf13/cobra"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	fwebhook "github.com/fairwindsops/polaris/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
	apitypes "k8s.io/apimachinery/pkg/types"
)

var webhookPort int
var disableWebhookConfigInstaller bool

func init() {
	rootCmd.AddCommand(webhookCmd)
	webhookCmd.PersistentFlags().IntVarP(&webhookPort, "port", "p", 9876, "Port for the dashboard webserver.")
	webhookCmd.PersistentFlags().BoolVar(&disableWebhookConfigInstaller, "disable-webhook-config-installer", false, "disable the installer in the webhook server, so it won't install webhook configuration resources during bootstrapping.")
}

var webhookCmd = &cobra.Command{
	Use:   "webhook",
	Short: "Runs the webhook webserver.",
	Long:  `Runs the webhook webserver.`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Debug("Setting up controller manager")
		mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{})
		if err != nil {
			logrus.Errorf("Unable to set up overall controller manager: %v", err)
			os.Exit(1)
		}

		polarisAppName := "polaris"
		polarisResourceName := "polaris-webhook"
		polarisNamespaceBytes, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")

		if err != nil {
			// Not exiting here as we have fallback options
			logrus.Debugf("Error reading namespace information: %v", err)
		}

		polarisNamespace := string(polarisNamespaceBytes)
		if polarisNamespace == "" {
			polarisNamespace = polarisResourceName
			logrus.Debugf("Could not determine current namespace, creating resources in %s namespace", polarisNamespace)
		}

		logrus.Info("Setting up webhook server")
		as, err := webhook.NewServer(polarisResourceName, mgr, webhook.ServerOptions{
			Port:                          int32(webhookPort),
			CertDir:                       "/opt/cert",
			DisableWebhookConfigInstaller: &disableWebhookConfigInstaller,
			BootstrapOptions: &webhook.BootstrapOptions{
				ValidatingWebhookConfigName: polarisResourceName,
				Secret: &apitypes.NamespacedName{
					Namespace: polarisNamespace,
					Name:      polarisResourceName,
				},

				Service: &webhook.Service{
					Namespace: polarisNamespace,
					Name:      polarisResourceName,

					// Selectors should select the pods that runs this webhook server.
					Selectors: map[string]string{
						"app":       polarisAppName,
						"component": "webhook",
					},
				},
			},
		})

		if err != nil {
			logrus.Errorf("Error setting up webhook server: %v", err)
			os.Exit(1)
		}

		logrus.Infof("Polaris webhook server listening on port %d", webhookPort)

		// Iterate all the configurations supported controllers to scan and register them for webhooks
		// Should only register controllers that are configured to be scanned
		logrus.Debug("Registering webhooks to the webhook server")
		var webhooks []webhook.Webhook
		for index, controllerToScan := range c.ControllersToScan {
			for innerIndex, supportedAPIType := range controllerToScan.ListSupportedAPIVersions() {
				webhookName := strings.ToLower(fmt.Sprintf("%s-%d-%d", controllerToScan, index, innerIndex))
				hook := fwebhook.NewWebhook(webhookName, mgr, fwebhook.Validator{Config: c}, supportedAPIType)
				webhooks = append(webhooks, hook)
			}
		}

		if err = as.Register(webhooks...); err != nil {
			logrus.Debugf("Unable to register webhooks in the admission server: %v", err)
			os.Exit(1)
		}

		logrus.Debug("Starting webhook manager")
		if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
			logrus.Errorf("Error starting manager: %v", err)
			os.Exit(1)
		}
	},
}
