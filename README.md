<div align="center">
  <img src="/pkg/dashboard/assets/images/polaris-logo.png" alt="Polaris Logo" />
  <br>

  [![Version][version-image]][version-link] [![CircleCI][circleci-image]][circleci-link] [![Go Report Card][goreport-image]][goreport-link]
</div>

[version-image]: https://img.shields.io/static/v1.svg?label=Version&message=0.4.0&color=239922
[version-link]: https://github.com/FairwindsOps/polaris

[goreport-image]: https://goreportcard.com/badge/github.com/FairwindsOps/polaris
[goreport-link]: https://goreportcard.com/report/github.com/FairwindsOps/polaris

[circleci-image]: https://circleci.com/gh/FairwindsOps/polaris.svg?style=svg
[circleci-link]: https://circleci.com/gh/FairwindsOps/polaris.svg

Polaris helps keep your cluster healthy. It runs a variety of checks to ensure that
Kubernetes deployments are configured using best practices, helping you avoid
problems in the future. Polaris can be run in a few different modes:

- A dashboard that provides an overview of how well current deployments are configured within a cluster.
- An experimental validating webhook that can prevent any future deployments that do not live up to a configured standard.
- A command-line audit that can be incorporated into your CI/CD pipeline

**Want to learn more?** Fairwinds holds [office hours on Zoom](https://zoom.us/j/242508205) the first Friday of every month, at 12pm Eastern. You can also reach out via email at `opensource@fairwinds.com`

# Dashboard Quickstart

```
kubectl apply -f https://github.com/FairwindsOps/polaris/releases/latest/download/dashboard.yaml
kubectl port-forward --namespace polaris svc/polaris-dashboard 8080:80
```
With the port forwarding in place, you can open http://localhost:8080 in your browser to view the dashboard.

* * *

# Components
## Dashboard

The Polaris dashboard is a way to get a simple visual overview of the current state of your Kubernetes deployments as well as a roadmap for what can be improved. The dashboard provides a cluster wide overview as well as breaking out results by category, namespace, and deployment.

<p align="center">
  <img src="/dashboard-screenshot.png" alt="Polaris Dashboard" />
</p>

Our default standards in Polaris are rather high, so don’t be surprised if your score is lower than you might expect. A key goal for Polaris was to set a high standard and aim for great configuration by default. If the defaults we’ve included are too strict, it’s easy to adjust the configuration as part of the deployment configuration to better suit your workloads.


## Webhook

Polaris includes experimental support for an optional validating webhook. This accepts the same configuration as the dashboard, and can run the same validations. This webhook will reject any deployments that trigger a validation error. This is indicative of the greater goal of Polaris, not just to encourage better configuration through dashboard visibility, but to actually enforce it with this webhook. *Although we are working towards greater stability and better test coverage, we do not currently consider this webhook component production ready.*

Unfortunately we have not found a way to display warnings as part of `kubectl` output unless we are rejecting a deployment altogether. That means that any checks with a severity of `warning` will still pass webhook validation, and the only evidence of that warning will either be in the Polaris dashboard or the Polaris webhook logs.

# Usage Documentation
See the [Usage Guide](/docs/usage.md) in the docs folder.

# Contributing
PRs welcome! Check out the [Contributing Guidelines](CONTRIBUTING.md),
[Code of Conduct](CODE_OF_CONDUCT.md), and [Roadmap](ROADMAP.md) for more information.

# Further Information
A history of changes to this project can be viewed in the [Changelog](CHANGELOG.md)

If you'd like to learn more about Polaris, or if you'd like to speak with
a Kubernetes expert, you can contact `info@fairwinds.com` or [visit our website](https://fairwinds.com)

# License
Apache License 2.0
