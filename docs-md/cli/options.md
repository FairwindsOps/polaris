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

