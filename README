## Build locally

```bash
git clone https://github.com/reactiveops/fairwinds $GOPATH/src/github.com/reactiveops/fairwinds
go build -a -o fairwinds *.go
./fairwinds -h
```

## Options

* `dashboard` Runs the webserver for Fairwinds dashboard.
* `dashboard-port`: Port for the dashboard webserver (default 8080)
* `webhook`: Runs the webhook webserver.
* `webhook-port`: Port for the webhook webserver (default 9876)
* `disable-webhook-config-installer`: disable the installer in the webhook server, so it won't install webhook configuration resources during bootstrapping
* `kubeconfig`: Paths to a kubeconfig. Only required if out-of-cluster.
* `master`: The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.
