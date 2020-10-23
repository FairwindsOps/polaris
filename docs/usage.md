# Installation and Usage
Polaris can be installed on your cluster using kubectl or Helm. It can also
be run as a local binary, which will use your kubeconfig to connect to the cluster
or run against local YAML files.

## Configuration
Polaris supports a wide range of validations covering a number of Kubernetes best practices.
Here's a [sample configuration file](/examples/config-full.yaml) that includes all currently supported checks.
The [default configuration](/examples/config.yaml) contains a number of those checks.


### Checks
Each check can be assigned a `severity`. Only checks with a severity of `danger` or `warning` will be validated. The results of these validations are visible on the dashboard. In the case of the validating webhook, only failures with a severity of `danger` will result in a change being rejected.

Polaris validation checks fall into several different categories:

- [Security](check-documentation/security.md)
- [Reliability](check-documentation/reliability.md)
- [Efficiency](check-documentation/efficiency.md)

#### Custom Checks
If you'd like to create your own checks, you can use [JSON Schema](https://json-schema.org/). For example,
to disallow images from quay.io:

```yaml
checks:
  imageRegistry: warning
customChecks:
  imageRegistry:
    successMessage: Image comes from allowed registries
    failureMessage: Image should not be from disallowed registry
    category: Images
    target: Container # target can be "Container" or "Pod"
    schema:
      '$schema': http://json-schema.org/draft-07/schema
      type: object
      properties:
        image:
          type: string
          not:
            pattern: ^quay.io
```

Schemas can also be specified as JSON strings instead of YAML, for easier copy/pasting:
```yaml
customChecks:
  foo:
    jsonSchema: |
      {
        "$schema": "http://json-schema.org/draft-07/schema",
        "type": "object"
      }
```

We extend JSON Schema with `resourceMinimum` and `resourceMaximum` fields to help compare memory and CPU resource
strings like `1000m` and `1G`. You can see an example in [the extended config](/examples/config-full.yaml)

There are additional examples in the [checks folder](/checks).

### Exemptions
Sometimes a workload really does need to do things that Polaris considers insecure. For instance,
many of the `kube-system` workloads need to run as root, or need access to the host network. In these
cases, we can add **exemptions** to allow the workload to pass Polaris checks.

Exemptions can be added two ways: by annotating a controller, or editing the Polaris config.

To exempt a controller from all checks via annotations, use the annotation `polaris.fairwinds.com/exempt=true`, e.g.
```
kubectl annotate deployment my-deployment polaris.fairwinds.com/exempt=true
```

To exempt a controller via the config, you have to specify a namespace (optional), a list of controller names and a list of rules, e.g.
```yaml
exemptions:
  # exception valid for kube-system namespace
  - namespace: kube-system
    controllerNames:
      - dns-controller
    rules:
      - hostNetworkSet
  # exception valid in all namespaces
  - controllerNames:
      - dns-controller
    rules:
      - hostNetworkSet
```

To exempt a controller from a particular check via annotations, use an annotation in the form of `polaris.fairwinds.com/<check>-exempt=true`, e.g.
```
kubectl annotate deployment my-deployment polaris.fairwinds.com/cpuRequestsMissing-exempt=true
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
helm repo add fairwinds-stable https://charts.fairwinds.com/stable
helm upgrade --install polaris fairwinds-stable/polaris --namespace polaris
kubectl port-forward --namespace polaris svc/polaris-dashboard 8080:80
```

### Local Binary
You'll need a valid `KUBECONFIG` set up for the dashboard to connect to your cluster.

Binary releases can be dowloaded from the [releases page](https://github.com/fairwindsops/polaris/releases)
or can be installed with [Homebrew](https://brew.sh/):
```bash
brew tap reactiveops/tap
brew install reactiveops/tap/polaris
polaris dashboard --port 8080
```

You can also point the dashboard to the local filesystem, instead of a live cluster:
```bash
polaris dashboard --port 8080 --audit-path=./deploy/
```

### Local Docker container

```
docker run -d -p8080:8080 -v ~/.kube/config:/opt/app/config:ro  quay.io/fairwinds/polaris:1.2 polaris dashboard --kubeconfig /opt/app/config
```

## Webhook
### kubectl
```bash
kubectl apply -f https://github.com/fairwindsops/polaris/releases/latest/download/webhook.yaml
```

### Helm
```bash
helm repo add fairwindsops-stable https://charts.fairwindsops.com/stable
helm upgrade --install polaris fairwindsops-stable/polaris --namespace polaris \
  --set webhook.enable=true --set dashboard.enable=false
```

## CLI
### Installation
Binary releases can be downloaded from the [releases page](https://github.com/fairwindsops/polaris/releases)
or can be installed with [Homebrew](https://brew.sh/):
```bash
brew tap FairwindsOps/tap
brew install FairwindsOps/tap/polaris
polaris version
```

You can run audits on the command line and see the output as JSON, YAML, or a raw score:
```bash
polaris audit --format yaml > report.yaml
polaris audit --format score
# 92
```

Audits can run against a local directory or YAML file rather than a cluster:
```bash
polaris audit --audit-path ./deploy/

# or to use STDIN
cat pod.yaml | polaris audit --audit-path -
```

You can also run the audit on a single resource instead of the entire cluster:
```bash
polaris audit --resource "nginx-ingress/Deployment.apps/v1/default-backend"
```

#### Running with CI/CD
You can integrate Polaris into CI/CD for repositories containing infrastructure-as-code.
For example, to fail if polaris detects *any* danger-level issues, or if the score drops below 90%:
```bash
polaris audit --audit-path ./deploy/ \
  --set-exit-code-on-danger \
  --set-exit-code-below-score 90
```

For more on exit code meanings, see [exit-code docs](exit-codes.md).

#### CLI Options

```
# top-level commands
audit
      Runs a one-time audit.
dashboard
      Runs the webserver for Polaris dashboard.
help
      Prints help, if you give it a command then it will print help for that command. Same as -h
version
      Prints the version of Polaris
webhook
      Runs the webhook webserver

# high-level flags
-c, --config string
      Location of Polaris configuration file
--disallow-exemptions
      Disallow any exemptions from configuration file.
-h, --help
      Help for Polaris (same as help command)
--kubeconfig string
      Path to a kubeconfig. Only required if out-of-cluster.
--log-level string
      Logrus log level (default "info")
--master string
      The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.

# dashboard flags
--audit-path string
      If specified, audits one or more YAML files instead of a cluster
--base-path string
      Path on which the dashboard is served (default "/")
--display-name string
      An optional identifier for the audit
--load-audit-file string
      Runs the dashboard with data saved from a past audit.
-p, --port int
      Port for the dashboard webserver (default 8080)

# audit flags
--audit-path string
      If specified, audits one or more YAML files instead of a cluster
--resource string
      If specified, audit a specific resource, in the format namespace/kind/version/name, e.g. nginx-ingress/Deployment.apps/v1/default-backend
--display-name string
      An optional identifier for the audit
--format string
      Output format for results - json, yaml, or score (default "json")
--output-file string
      Destination file for audit results
--output-url string
      Destination URL to send audit results
--set-exit-code-below-score int
      Set an exit code of 4 when the score is below this threshold (1-100)
--set-exit-code-on-danger
      Set an exit code of 3 when the audit contains danger-level issues.

# webhook flags
--disable-webhook-config-installer
      disable the installer in the webhook server, so it won't install webhook configuration resources during bootstrapping
-p, --port int
      Port for the webhook webserver (default 9876)
```
