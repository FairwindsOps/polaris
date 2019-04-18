## Build locally
This package is best built using [packr](https://github.com/gobuffalo/packr), which provides
a thin wrapper around the go compiler in order to include static HTML/CSS/JS assets.
```bash
git clone https://github.com/reactiveops/fairwinds $GOPATH/src/github.com/reactiveops/fairwinds
go get -u github.com/gobuffalo/packr/v2/packr2
packr2 build -a -o fairwinds *.go
./fairwinds -h
```

## Run on-cluster

On GKE, you'll need to run the following command first:
```
kubectl create clusterrolebinding cluster-admin-binding \
  --clusterrole cluster-admin \
  --user $(gcloud config get-value account)
```

Then apply the config:
```
kubectl apply -f deploy/all.yaml
```


## Options

* `config`: Specify a location for the Fairwinds config
* `dashboard`: Runs the webserver for Fairwinds dashboard.
* `dashboard-port`: Port for the dashboard webserver (default 8080)
* `webhook`: Runs the webhook webserver.
* `webhook-port`: Port for the webhook webserver (default 9876)
* `disable-webhook-config-installer`: disable the installer in the webhook server, so it won't install webhook configuration resources during bootstrapping
* `kubeconfig`: Paths to a kubeconfig. Only required if out-of-cluster.
* `master`: The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.

## Helm Deploy Option

* Create release with Helm:
```
helm upgrade --install fairwinds deploy/helm/fairwinds/ --namespace fairwinds --recreate-pods
kubectl port-forward --namespace fairwinds svc/fairwinds-fairwinds-dashboard 8080:80 &
open http://localhost:8080
```

## Run tests
```
go list ./... | grep -v vendor | xargs golint -set_exit_status
go list ./... | grep -v vendor | xargs go vet
go test ./pkg/... -v -coverprofile cover.out
```
