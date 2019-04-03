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
	glog "log"
	"net/http"
	"os"
	"strconv"

	conf "github.com/reactiveops/fairwinds/pkg/config"
	"github.com/reactiveops/fairwinds/pkg/dashboard"
	"github.com/reactiveops/fairwinds/pkg/kube"
	"github.com/reactiveops/fairwinds/pkg/validator"
	fwebhook "github.com/reactiveops/fairwinds/pkg/webhook"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apitypes "k8s.io/apimachinery/pkg/types"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var log = logf.Log.WithName("fairwinds")

func main() {
	dashboard := flag.Bool("dashboard", false, "Runs the webserver for Fairwinds dashboard.")
	webhook := flag.Bool("webhook", false, "Runs the webhook webserver.")
	audit := flag.Bool("audit", false, "Runs a one-time audit.")
	dashboardPort := flag.Int("dashboard-port", 8080, "Port for the dashboard webserver")
	webhookPort := flag.Int("webhook-port", 9876, "Port for the webhook webserver")
	auditDestination := flag.String("audit-destination", "", "Destination URL to send audit results (prints to stdout if unspecified)")

	var disableWebhookConfigInstaller bool
	flag.BoolVar(&disableWebhookConfigInstaller, "disable-webhook-config-installer", false,
		"disable the installer in the webhook server, so it won't install webhook configuration resources during bootstrapping")

	flag.Parse()

	c, err := conf.ParseFile("config.yml")
	if err != nil {
		glog.Println("Error parsing config.yml:", err)
		os.Exit(1)
	}

	if !*dashboard && !*webhook && !*audit {
		*audit = true
	}

	if *webhook {
		startWebhookServer(c, disableWebhookConfigInstaller, *webhookPort)
	} else if *dashboard {
		startDashboardServer(c, *dashboardPort)
	} else if *audit {
		runAudit(c, *auditDestination)
	}
}

func startDashboardServer(c conf.Configuration, port int) {
	k, _ := kube.CreateKubeAPI()
	http.HandleFunc("/results.json", func(w http.ResponseWriter, r *http.Request) {
		dashboard.EndpointHandler(w, r, c, k)
	})
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("public/"))))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		dashboard.MainHandler(w, r, c, k)
	})
	portStr := strconv.Itoa(port)
	glog.Println("Starting Fairwinds dashboard server on port " + portStr)
	glog.Fatal(http.ListenAndServe(":"+portStr, nil))
}

func startWebhookServer(c conf.Configuration, disableWebhookConfigInstaller bool, port int) {
	logf.SetLogger(logf.ZapLogger(false))
	entryLog := log.WithName("entrypoint")

	// Setup a Manager
	entryLog.Info("setting up manager")
	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{})
	if err != nil {
		entryLog.Error(err, "unable to set up overall controller manager")
		os.Exit(1)
	}

	fairwindsResourceName := "fairwinds"
	fairwindsNamespaceBytes, _ := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	fairwindsNamespace := string(fairwindsNamespaceBytes)
	if fairwindsNamespace == "" {
		fmt.Printf("could not determine current namespace, creating resources in %s namespace\n", fairwindsResourceName)
		fairwindsNamespace = fairwindsResourceName
	}

	entryLog.Info("setting up webhook server")
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
		entryLog.Error(err, "unable to create a new webhook server")
		os.Exit(1)
	} else {
		glog.Println("Fairwinds webhook server listening on port " + strconv.Itoa(port))
	}

	p := fwebhook.NewWebhook("pod", mgr, fwebhook.Validator{Config: c}, &corev1.Pod{})
	d := fwebhook.NewWebhook("deploy", mgr, fwebhook.Validator{Config: c}, &appsv1.Deployment{})
	entryLog.Info("registering webhooks to the webhook server")
	if err = as.Register(p, d); err != nil {
		entryLog.Error(err, "unable to register webhooks in the admission server")
		os.Exit(1)
	}

	entryLog.Info("starting manager")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		entryLog.Error(err, "unable to run manager")
		os.Exit(1)
	}
}

func runAudit(c conf.Configuration, destination string) {
	k, _ := kube.CreateKubeAPI()
	auditData, err := validator.RunAudit(c, k)
	if err != nil {
		panic(err)
	}

	if destination != "" {
		jsonData, err := json.Marshal(auditData)
		if err != nil {
			panic(err)
		}

		req, err := http.NewRequest("POST", destination, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		glog.Println(string(body))
	} else {
		y, err := yaml.Marshal(auditData)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(y))
	}
}
