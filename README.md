<<<<<<< HEAD
## Build locally
This package is best built using [packr](https://github.com/gobuffalo/packr), which provides
a thin wrapper around the go compiler in order to include static HTML/CSS/JS assets.
```bash
git clone https://github.com/reactiveops/fairwinds $GOPATH/src/github.com/reactiveops/fairwinds
go get -u github.com/gobuffalo/packr/v2/packr2
packr2 build -a -o fairwinds *.go
./fairwinds -h
=======
<p align="center">
  <img src="/public/images/logo.png" alt="Fairwinds Logo" />
</p>

Fairwinds aims to keep your cluster sailing smoothly. It runs a variety of checks to ensure that Kubernetes deployments are configured using best practices that will avoid potential problems in the future. The project includes two primary parts:

- A dashboard to display the results of these validations on your existing deployments
- A webhook that can prevent poorly configured deployments from reaching your cluster

## Dashboard

The Fairwinds Dashboard provides an overview of your current deployments in a cluster along with their validation scores. An overall score is provided for a cluster on a 0 - 100 scale. Results for each validation are grouped by namespace and deployment.

### Deploying

To deploy Fairwinds with kubectl:

```
kubectl apply -f deploy/all.yaml
>>>>>>> first run at readme updates
```

Fairwinds can also be deployed with Helm:

```
helm upgrade --install fairwinds deploy/helm/fairwinds/ --namespace fairwinds --recreate-pods
```

### Viewing the Dashboard

Once the dashboard is deployed, it can be viewed by using kubectl port-forward:
```
kubectl port-forward --namespace fairwinds svc/fairwinds-fairwinds-dashboard 8080:80 &
open http://localhost:8080
```

### Using a Binary Release

If you'd prefer to run Fairwinds locally, binary releases are available on the [releases page](https://github.com/reactiveops/fairwinds/releases). With a Fairwinds binary, the

## Webhook

The Fairwinds Webhook can run the same checks as the dashboard, but can be deployed and configured separately. When running, the webhook will validate any new or updated deployments in the cluster, and reject any that fail a check with an `error` severity.

Unfortunately we have not found a way to disply warnings as part of `kubectl` output unless we are rejecting a deployment altogether. That means that any checks with a severity of `warning` will still pass webhook validation, and the only evidence of that warning will either be in the Fairwinds dashboard or the Fairwinds webhook logs.


## CLI Options

* `config`: Specify a location for the Fairwinds config
* `dashboard`: Runs the webserver for Fairwinds dashboard.
* `dashboard-port`: Port for the dashboard webserver (default 8080)
* `webhook`: Runs the webhook webserver.
* `webhook-port`: Port for the webhook webserver (default 9876)
* `disable-webhook-config-installer`: disable the installer in the webhook server, so it won't install webhook configuration resources during bootstrapping
* `kubeconfig`: Paths to a kubeconfig. Only required if out-of-cluster.