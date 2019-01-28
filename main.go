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

	conf "github.com/reactiveops/fairwinds/pkg/config"
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

	var disableWebhookConfigInstaller bool
	flag.BoolVar(&disableWebhookConfigInstaller, "disable-webhook-config-installer", false,
		"disable the installer in the webhook server, so it won't install webhook configuration resources during bootstrapping")
	flag.Parse()

	c, err := conf.ParseFile("config.yml")
	if err != nil {
		log.Error(err, "error parsing config file")
		os.Exit(1)
	}

	if *webhook {
		err = startWebhookServer(c, disableWebhookConfigInstaller)
	}
	if err != nil {
		log.Error(err, "error starting webhook server")
		os.Exit(1)
	}

	if *dashboard {
		startDashboardServer(c)
	}

}

func startDashboardServer(c conf.Configuration) {
	http.HandleFunc("/validate", func(w http.ResponseWriter, r *http.Request) { validator.DeployHandler(w, r, c) })
	http.HandleFunc("/ping", validator.PingHandler)
	glog.Println("Starting Fairwinds dashboard webserver on port 8080.")
	glog.Fatal(http.ListenAndServe(":8080", nil))
}

func startWebhookServer(c conf.Configuration, disableWebhookConfigInstaller bool) error {
	logf.SetLogger(logf.ZapLogger(false))
	entryLog := log.WithName("entrypoint")

	// Setup a Manager
	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{})
	if err != nil {
		return err
	}

	as, err := webhook.NewServer(FairwindsName, mgr, webhook.ServerOptions{
		Port:                          9876,
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
		return err
	}

	p, err := validator.NewWebhook("pod", mgr, validator.Validator{Config: c}, &corev1.Pod{})
	if err != nil {
		return err
	}
	d, err := validator.NewWebhook("deploy", mgr, validator.Validator{Config: c}, &appsv1.Deployment{})
	if err != nil {
		return err
	}
	entryLog.Info("Registering webhooks to the webhook server...")
	if err = as.Register(p, d); err != nil {
		return err
	}

	entryLog.Info("Starting Fairwinds webhook server on port 9876.")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		return err
	}
	return nil
}
