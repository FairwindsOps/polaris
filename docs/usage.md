# Installation and Usage
Polaris can be installed on your cluster using kubectl or Helm. It can also
be run as a local binary, which will use your kubeconfig to connect to the cluster
or run against local YAML files.

## Configuration
Polaris supports a wide range of validations covering a number of Kubernetes best practices.
Here's a [sample configuration file](/examples/config-full.yaml) that includes all currently supported checks.
The [default configuration](/examples/config.yaml) contains a number of those checks.


### Checks
Each check can be assigned a `severity`. Only checks with a severity of `error` or `warning` will be validated. The results of these validations are visible on the dashboard. In the case of the validating webhook, only failures with a severity of `error` will result in a change being rejected.

Polaris validation checks fall into several different categories:

- [Health Checks](check-documentation/health-checks.md)
- [Images](check-documentation/images.md)
- [Networking](check-documentation/networking.md)
- [Resources](check-documentation/resources.md)
- [Security](check-documentation/security.md)

### Exemptions
Exemptions can be added two ways: by annotating a controller, or editing the Polaris config.

To exempt a controller via annotations, use the annotation `polaris.fairwinds.com/exempt=true`, e.g.
```
kubectl annotate deployment my-deployment polaris.fairwinds.com/exempt=true
```

To exempt a controller via the config, you have to specify a list of controller names and a list of rules, e.g.
```yaml
exemptions:
  - controllerNames:
      - dns-controller
    rules:
      - hostNetworkSet
```

# Installing
There are several ways to install and use Polaris. Below outline ways to install using `kubectl`, `helm` and `local binary`.

## Dashboard
The dashboard can be installed on a cluster using kubectl or Helm. It can also be run locally,
connecting to your cluster using the credentials stored in your `KUBECONFIG`.

### kubectl
```bash
kubectl apply -f https://github.com/fairwindsops/polaris/releases/latest/download/dashboard.yaml
kubectl port-forward --namespace polaris svc/polaris-dashboard 8080:80
```
### Helm
```bash
helm repo add reactiveops-stable https://charts.reactiveops.com/stable
helm upgrade --install polaris reactiveops-stable/polaris --namespace polaris
kubectl port-forward --namespace polaris svc/polaris-dashboard 8080:80
```

### Local Binary
You'll need a valid `KUBECONFIG` set up for the dashboard to connect to your cluster.

Binary releases can be dowloaded from the [releases page](https://github.com/fairwindsops/polaris/releases)
or can be installed with [Homebrew](https://brew.sh/):
```bash
brew tap reactiveops/tap
brew install reactiveops/tap/polaris
polaris --dashboard --dashboard-port 8080
```

## Webhook
### kubectl
```bash
kubectl apply -f https://github.com/fairwindsops/polaris/releases/latest/download/webhook.yaml
```

### Helm
```bash
helm repo add reactiveops-stable https://charts.reactiveops.com/stable
helm upgrade --install polaris reactiveops-stable/polaris --namespace polaris \
  --set webhook.enable=true --set dashboard.enable=false
```

## CLI
### Installation
Binary releases can be downloaded from the [releases page](https://github.com/fairwindsops/polaris/releases)
or can be installed with [Homebrew](https://brew.sh/):
```bash
brew tap reactiveops/tap
brew install reactiveops/tap/polaris
polaris --version
```

You can run audits on the command line and see the output as JSON, YAML, or a raw score:
```bash
polaris --audit --output-format yaml > report.yaml
polaris --audit --output-format score
# 92
```

Both the dashboard and audits can run against a local directory or YAML file
rather than a cluster:
```bash
polaris --audit --audit-path ./deploy/
```

#### Running with CI/CD
You can integrate Polaris into CI/CD for repositories containing infrastructure-as-code.
For example, to fail if polaris detects *any* error-level issues, or if the score drops below 90%:
```bash
polaris --audit --audit-path ./deploy/ \
  --set-exit-code-on-error \
  --set-exit-code-below-score 90
```

For more on exit code meanings, see [exit-code docs](exit-codes.md).

#### CLI Options

```
# high-level flags
-version
      Prints the version of Polaris
-config string
      Location of Polaris configuration file
-kubeconfig string
      Path to a kubeconfig. Only required if out-of-cluster.
-log-level string
      Logrus log level (default "info")
-master string
      The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.

# dashboard flags
-dashboard
      Runs the webserver for Polaris dashboard.
-dashboard-base-path string
      Path on which the dashboard is served (default "/")
-dashboard-port int
      Port for the dashboard webserver (default 8080)
-display-name string
      An optional identifier for the audit

# audit flags
-audit
      Runs a one-time audit.
-audit-path string
      If specified, audits one or more YAML files instead of a cluster
-output-file string
      Destination file for audit results
-output-format string
      Output format for results - json, yaml, or score (default "json")
-output-url string
      Destination URL to send audit results
-set-exit-code-below-score int
      When running with --audit, set an exit code of 4 when the score is below this threshold (1-100)
-set-exit-code-on-error
      When running with --audit, set an exit code of 3 when the audit contains error-level issues.

# webhook flags
-webhook
      Runs the webhook webserver.
-webhook-port int
      Port for the webhook webserver (default 9876)
-disable-webhook-config-installer
      disable the installer in the webhook server, so it won't install webhook configuration resources during bootstrapping
```

