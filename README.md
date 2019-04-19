<p align="center">
  <img src="/public/images/logo.png" alt="Fairwinds Logo" />
</p>

Fairwinds aims to keep your cluster sailing smoothly. It runs a variety of checks to ensure that Kubernetes deployments are configured using best practices that will avoid potential problems in the future. The project includes two primary parts:

- A dashboard to display the results of these validations on your existing deployments
- A webhook that can prevent poorly configured deployments from reaching your cluster

## Dashboard

The Fairwinds Dashboard provides an overview of your current deployments in a cluster along with their validation scores. An overall score is provided for a cluster on a 0 - 100 scale. Results for each validation are grouped by namespace and deployment.

<p align="center">
  <img src="/dashboard-screenshot.png" alt="Fairwinds Dashboard" />
</p>

### Deploying

To deploy Fairwinds with kubectl:

```
kubectl apply -f deploy/all.yaml
```

Fairwinds can also be deployed with Helm:

```
helm upgrade --install fairwinds deploy/helm/fairwinds/ --namespace fairwinds
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

Fairwinds includes experimental support for an optional validating webhook. This accepts the same configuration as the dashboard, and can run the same validations. This webhook will reject any deployments that trigger a validation error. This is indicative of the greater goal of Fairwinds, not just to encourage better configuration through dashboard visibility, but to actually enforce it with this webhook. *Although we are working towards greater stability and better test coverage, we do not currently consider this webhook component production ready.*

Unfortunately we have not found a way to disply warnings as part of `kubectl` output unless we are rejecting a deployment altogether. That means that any checks with a severity of `warning` will still pass webhook validation, and the only evidence of that warning will either be in the Fairwinds dashboard or the Fairwinds webhook logs.

## CLI Options

* `config`: Specify a location for the Fairwinds config
* `dashboard`: Runs the webserver for Fairwinds dashboard.
* `dashboard-port`: Port for the dashboard webserver (default 8080)
* `webhook`: Runs the webhook webserver.
* `webhook-port`: Port for the webhook webserver (default 9876)
* `disable-webhook-config-installer`: disable the installer in the webhook server, so it won't install webhook configuration resources during bootstrapping
* `kubeconfig`: Paths to a kubeconfig. Only required if out-of-cluster.

## Configuration

Fairwinds supports a wide range of validations covering a number of Kubernetes best practices. Here's a sample configuration file that includes all currently supported checks. The [default configuration](https://github.com/reactiveops/fairwinds/blob/master/config.yaml) contains a number of those checks. This repository also includes a sample [full configuration file](https://github.com/reactiveops/fairwinds/blob/master/config-full.yaml) that enables all available checks.

Each check can be assigned a `severity`. Only checks with a severity of `error` or `warning` will be validated. The results of these validations are visible on the dashboard. In the case of the validating webhook, only failures with a severity of `error` will result in a change being rejected.

Fairwinds validation checks fall into several different categories:

- [Health Checks](docs/health-checks.md)
- [Images](docs/images.md)
- [Networking](docs/networking.md)
- [Resources](docs/resources.md)
- [Security](docs/security.md)

## License
Apache License 2.0
