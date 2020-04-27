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
	"io/ioutil"
	"os"
	"strings"

	fwebhook "github.com/fairwindsops/polaris/pkg/webhook"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	batchv2alpha1 "k8s.io/api/batch/v2alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	apitypes "k8s.io/apimachinery/pkg/types"
	k8sConfig "sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var supportedVersions = map[string]runtime.Object{
	"appsv1/Deployment":      &appsv1.Deployment{},
	"appsv1beta1/Deployment": &appsv1beta1.Deployment{},
	"appsv1beta2/Deployment": &appsv1beta2.Deployment{},

	"appsv1/StatefulSet":      &appsv1.StatefulSet{},
	"appsv1beta1/StatefulSet": &appsv1beta1.StatefulSet{},
	"appsv1beta2/StatefulSet": &appsv1beta2.StatefulSet{},

	"appsv1/DaemonSet":      &appsv1.DaemonSet{},
	"appsv1beta2/DaemonSet": &appsv1beta2.DaemonSet{},

	"batchv1/Job": &batchv1.Job{},

	"batchv1beta1/CronJob":  &batchv1beta1.CronJob{},
	"batchv2alpha1/CronJob": &batchv2alpha1.CronJob{},

	"corev1/ReplicationController": &corev1.ReplicationController{},

	"corev1/Pod": &corev1.Pod{},
}

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
		mgr, err := manager.New(k8sConfig.GetConfigOrDie(), manager.Options{})
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
		for name, supportedAPIType := range supportedVersions {
			webhookName := strings.ToLower(name)
			webhookName = strings.ReplaceAll(webhookName, "/", "-")
			hook, err := fwebhook.NewWebhook(webhookName, mgr, fwebhook.Validator{Config: config}, supportedAPIType)
			if err != nil {
				logrus.Warningf("Couldn't build webhook %s: %v", webhookName, err)
				continue
			}
			webhooks = append(webhooks, hook)
			logrus.Infof("%s webhook started", webhookName)
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
