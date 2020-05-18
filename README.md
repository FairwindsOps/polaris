<div align="center">
  <img src="/img/polaris-logo.png" alt="Polaris Logo" />
  <br>

  [![Version][version-image]][version-link] [![CircleCI][circleci-image]][circleci-link] [![Go Report Card][goreport-image]][goreport-link]
</div>

[version-image]: https://img.shields.io/static/v1.svg?label=Version&message=1.0.0&color=239922
[version-link]: https://github.com/FairwindsOps/polaris

[goreport-image]: https://goreportcard.com/badge/github.com/FairwindsOps/polaris
[goreport-link]: https://goreportcard.com/report/github.com/FairwindsOps/polaris

[circleci-image]: https://circleci.com/gh/FairwindsOps/polaris.svg?style=svg
[circleci-link]: https://circleci.com/gh/FairwindsOps/polaris.svg

Fairwinds' Polaris keeps your clusters sailing smoothly. It runs a variety of checks to ensure that
Kubernetes pods and controllers are configured using best practices, helping you avoid
problems in the future. Polaris can be run in a few different modes:

Polaris can be run in three different modes:
* As a [dashboard](#dashboard), so you can audit what's running inside your cluster.
* As a [validating webhook](#webhook), so you can automatically reject workloads that don't adhere to your organization's policies.
* As a [command-line tool](#cli), so you can test local YAML files, e.g. as part of a CI/CD process.

**Want to learn more?** Reach out on [the Slack channel](https://fairwindscommunity.slack.com/messages/polaris) ([request invite](https://join.slack.com/t/fairwindscommunity/shared_invite/zt-e3c6vj4l-3lIH6dvKqzWII5fSSFDi1g)), send an email to `opensource@fairwinds.com`, or join us for [office hours on Zoom](https://fairwindscommunity.slack.com/messages/office-hours)

# Dashboard Quickstart

```bash
kubectl apply -f https://github.com/FairwindsOps/polaris/releases/latest/download/dashboard.yaml
kubectl port-forward --namespace polaris svc/polaris-dashboard 8080:80
```
With the port forwarding in place, you can open http://localhost:8080 in your browser to view the dashboard.

* * *

# Components

## Dashboard
> [View installation instructions](docs/usage.md#dashboard)

The Polaris dashboard is a way to get a simple visual overview of the current state of your Kubernetes workloads as well as a roadmap for what can be improved. The dashboard provides a cluster wide overview as well as breaking out results by category, namespace, and workload.

<p align="center">
  <img src="/img/dashboard-screenshot.png" alt="Polaris Dashboard" width="550"/>
</p>

Our default standards in Polaris are rather high, so don’t be surprised if your score is lower than you might expect. A key goal for Polaris was to set a high standard and aim for great configuration by default. If the defaults we’ve included are too strict, it’s easy to adjust the configuration as part of the deployment configuration to better suit your workloads.


## Admission Controller: Validating Webhook
> [View installation instructions](docs/usage.md#webhook)

Polaris can be run as an admission controller that acts as a validating webhook. This accepts the same configuration as the dashboard, and can run the same validations. This webhook will reject any workloads that trigger a danger-level check. This is indicative of the greater goal of Polaris, not just to encourage better configuration through dashboard visibility, but to actually enforce it with this webhook. Polaris will not fix your workloads, only block them. 

Unfortunately we have not found a way to display warnings as part of `kubectl` output unless we are rejecting a workload altogether. That means that any checks with a severity of `warning` will still pass webhook validation, and the only evidence of that warning will either be in the Polaris dashboard or the Polaris webhook logs.

## CLI
> [View installation instructions](docs/usage.md#cli)

Polaris can also be used on the command line, either to audit local files or a running cluster.
This is particularly helpful for running Polaris against your infrastructure-as-code as part of a
CI/CD pipeline. Use the available [command line flags](docs/usage.md#running-with-ci-cd)
to cause CI/CD to fail if your Polaris score drops below a certain threshold, or if any danger-level issues arise.

# Installation and Usage
See the [Usage Guide](/docs/usage.md) for details on different methods for installing and using Polaris.

# Contributing
PRs welcome! Check out the [Contributing Guidelines](CONTRIBUTING.md),
[Code of Conduct](CODE_OF_CONDUCT.md), and [Roadmap](ROADMAP.md) for more information.

# Further Information
A history of changes to this project can be viewed in the [Changelog](CHANGELOG.md)

If you'd like to learn more about Polaris, or if you'd like to speak with
a Kubernetes expert, you can contact `info@fairwinds.com` or [visit our website](https://fairwinds.com)

# License
Apache License 2.0
