// Copyright 2018 ReactiveOps
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
	"flag"
	glog "log"
	"net/http"
	"os"
	"strconv"

	conf "github.com/reactiveops/fairwinds/pkg/config"
	"github.com/reactiveops/fairwinds/pkg/dashboard"
	"github.com/reactiveops/fairwinds/pkg/kube"
	"github.com/reactiveops/fairwinds/pkg/validator"
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

// FairwindsName is used for Kubernetes resource naming
var FairwindsName = "fairwinds"
var log = logf.Log.WithName(FairwindsName)

func main() {
	dashboard := flag.Bool("dashboard", false, "Runs the webserver for Fairwinds dashboard.")
	webhook := flag.Bool("webhook", false, "Runs the webhook webserver.")
	dashboardPort := flag.Int("dashboard-port", 8080, "Port for the dashboard webserver")
	webhookPort := flag.Int("webhook-port", 9876, "Port for the webhook webserver")

	var disableWebhookConfigInstaller bool
	flag.BoolVar(&disableWebhookConfigInstaller, "disable-webhook-config-installer", false,
		"disable the installer in the webhook server, so it won't install webhook configuration resources during bootstrapping")

	flag.Parse()

	c, err := conf.ParseFile("config.yml")
	if err != nil {
		glog.Println("Error parsing config.yml:", err)
		os.Exit(1)
	}

	if *webhook {
		startWebhookServer(c, disableWebhookConfigInstaller, *webhookPort)
	}

	if *dashboard {
		startDashboardServer(c, *dashboardPort)
	}

	if !*dashboard && !*webhook {
		glog.Println("Must specify either -webhook, -dashboard, or both")
		os.Exit(1)
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
	glog.Fatal(http.ListenAndServe(":" + portStr, nil))
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

	entryLog.Info("setting up webhook server")
	as, err := webhook.NewServer(FairwindsName, mgr, webhook.ServerOptions{
		Port:                          int32(port),
		CertDir:                       "/tmp/cert",
		DisableWebhookConfigInstaller: &disableWebhookConfigInstaller,
		BootstrapOptions: &webhook.BootstrapOptions{
			ValidatingWebhookConfigName: FairwindsName,
			Secret: &apitypes.NamespacedName{
				Namespace: FairwindsName,
				Name:      FairwindsName,
			},

			Service: &webhook.Service{
				Namespace: FairwindsName,
				Name:      FairwindsName,

				// Selectors should select the pods that runs this webhook server.
				Selectors: map[string]string{
					"app": FairwindsName,
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

	p := validator.NewWebhook("pod", mgr, validator.Validator{Config: c}, &corev1.Pod{})
	d := validator.NewWebhook("deploy", mgr, validator.Validator{Config: c}, &appsv1.Deployment{})
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
