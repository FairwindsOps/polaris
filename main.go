// Copyright 2019 ReactiveOps
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

	conf "github.com/reactiveops/fairwinds/pkg/config"
	"github.com/reactiveops/fairwinds/pkg/dashboard"
	"github.com/reactiveops/fairwinds/pkg/kube"
	"github.com/reactiveops/fairwinds/pkg/validator"
	fwebhook "github.com/reactiveops/fairwinds/pkg/webhook"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	apitypes "k8s.io/apimachinery/pkg/types"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

func main() {
	dashboard := flag.Bool("dashboard", false, "Runs the webserver for Fairwinds dashboard.")
	webhook := flag.Bool("webhook", false, "Runs the webhook webserver.")
	audit := flag.Bool("audit", false, "Runs a one-time audit.")
	auditPath := flag.String("audit-path", "", "If specified, audits one or more YAML files instead of a cluster")
	dashboardPort := flag.Int("dashboard-port", 8080, "Port for the dashboard webserver")
	webhookPort := flag.Int("webhook-port", 9876, "Port for the webhook webserver")
	auditOutputURL := flag.String("output-url", "", "Destination URL to send audit results")
	auditOutputFile := flag.String("output-file", "", "Destination file for audit results")
	configPath := flag.String("config", "config.yaml", "Location of Fairwinds configuration file")
	logLevel := flag.String("log-level", logrus.InfoLevel.String(), "Logrus log level")
	disableWebhookConfigInstaller := flag.Bool("disable-webhook-config-installer", false,
		"disable the installer in the webhook server, so it won't install webhook configuration resources during bootstrapping")

	flag.Parse()

	parsedLevel, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		logrus.Errorf("log-level flag has invalid value %s", *logLevel)
	} else {
		logrus.SetLevel(parsedLevel)
	}

	c, err := conf.ParseFile(*configPath)
	if err != nil {
		logrus.Errorf("Error parsing config at %s: %v", *configPath, err)
		os.Exit(1)
	}

	if !*dashboard && !*webhook && !*audit {
		*audit = true
	}

	if *webhook {
		startWebhookServer(c, *disableWebhookConfigInstaller, *webhookPort)
	} else if *dashboard {
		k, _ := kube.CreateResourceProvider(*auditPath)
		startDashboardServer(c, k, *dashboardPort)
	} else if *audit {
		k, _ := kube.CreateResourceProvider(*auditPath)
		runAudit(c, k, *auditOutputFile, *auditOutputURL)
	}
}

func startDashboardServer(c conf.Configuration, k *kube.ResourceProvider, port int) {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
	http.HandleFunc("/results.json", func(w http.ResponseWriter, r *http.Request) {
		dashboard.EndpointHandler(w, r, c, k)
	})
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "public/favicon.ico")
	})
	fileServer := http.FileServer(dashboard.GetAssetBox())
	http.Handle("/static/", http.StripPrefix("/static/", fileServer))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		auditData, err := validator.RunAudit(c, k)
		if err != nil {
			logrus.Errorf("Error getting audit data: %v", err)
			http.Error(w, "Error running audit", 500)
			return
		}
		dashboard.MainHandler(w, r, auditData)
	})

	logrus.Infof("Starting Fairwinds dashboard server on port %d", port)
	logrus.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func startWebhookServer(c conf.Configuration, disableWebhookConfigInstaller bool, port int) {
	logrus.Debug("Setting up controller manager")
	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{})
	if err != nil {
		logrus.Errorf("Unable to set up overall controller manager: %v", err)
		os.Exit(1)
	}

	fairwindsResourceName := "fairwinds"
	fairwindsNamespaceBytes, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")

	if err != nil {
		// Not exiting here as we have fallback options
		logrus.Debugf("Error reading namespace information: %v", err)
	}

	fairwindsNamespace := string(fairwindsNamespaceBytes)
	if fairwindsNamespace == "" {
		fairwindsNamespace = fairwindsResourceName
		logrus.Debugf("Could not determine current namespace, creating resources in %s namespace", fairwindsNamespace)
	}

	logrus.Info("Setting up webhook server")
	as, err := webhook.NewServer(fairwindsResourceName, mgr, webhook.ServerOptions{
		Port:                          int32(port),
		CertDir:                       "/tmp/cert",
		DisableWebhookConfigInstaller: &disableWebhookConfigInstaller,
		BootstrapOptions: &webhook.BootstrapOptions{
			ValidatingWebhookConfigName: fairwindsResourceName,
			Secret: &apitypes.NamespacedName{
				Namespace: fairwindsNamespace,
				Name:      fairwindsResourceName,
			},

			Service: &webhook.Service{
				Namespace: fairwindsNamespace,
				Name:      fairwindsResourceName,

				// Selectors should select the pods that runs this webhook server.
				Selectors: map[string]string{
					"app": fairwindsResourceName,
				},
			},
		},
	})

	if err != nil {
		logrus.Errorf("Error setting up webhook server: %v", err)
		os.Exit(1)
	}

	logrus.Infof("Fairwinds webhook server listening on port %d", port)

	d := fwebhook.NewWebhook("deploy", mgr, fwebhook.Validator{Config: c}, &appsv1.Deployment{})
	logrus.Debug("Registering webhooks to the webhook server")
	if err = as.Register(d); err != nil {
		logrus.Debugf("Unable to register webhooks in the admission server: %v", err)
		os.Exit(1)
	}

	logrus.Debug("Starting webhook manager")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		logrus.Errorf("Error starting manager: %v", err)
		os.Exit(1)
	}
}

func runAudit(c conf.Configuration, k *kube.ResourceProvider, outputFile string, outputURL string) {
	auditData, err := validator.RunAudit(c, k)

	if err != nil {
		panic(err)
	}

	if outputURL == "" && outputFile == "" {
		yamlBytes, err := yaml.Marshal(auditData)

		if err != nil {
			logrus.Errorf("Error marshalling YAML: %v", err)
			os.Exit(1)
		}

		os.Stdout.Write(yamlBytes)

	} else {
		jsonData, err := json.MarshalIndent(auditData, "", "  ")

		if err != nil {
			logrus.Errorf("Error marshalling JSON: %v", err)
			os.Exit(1)
		}

		if outputURL != "" {
			req, err := http.NewRequest("POST", outputURL, bytes.NewBuffer(jsonData))

			if err != nil {
				logrus.Errorf("Error building request for output: %v", err)
				os.Exit(1)
			}

			req.Header.Set("Content-Type", "application/json")
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
			err := ioutil.WriteFile(outputFile, []byte(jsonData), 0644)
			if err != nil {
				logrus.Errorf("Error writing output to file: %v", err)
				os.Exit(1)
			}
		}
	}
}
