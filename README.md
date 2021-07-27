<div align="center" class="no-border">
  <img src="https://polaris.docs.fairwinds.com/img/polaris-logo.png" alt="Polaris Logo">
  <br>
  <h3>Best Practices for Kubernetes Workload Configuration</h3>
  <a href="https://github.com/FairwindsOps/polaris">
    <img src="https://img.shields.io/static/v1.svg?label=Version&message=4.0.8&color=239922">
  </a>
  <a href="https://goreportcard.com/report/github.com/FairwindsOps/polaris">
    <img src="https://goreportcard.com/badge/github.com/FairwindsOps/polaris">
  </a>
  <a href="https://circleci.com/gh/FairwindsOps/polaris">
    <img src="https://circleci.com/gh/FairwindsOps/polaris.svg?style=svg">
  </a>
  <a href="https://insights.fairwinds.com/gh/FairwindsOps/polaris">
    <img src="https://insights.fairwinds.com/v0/gh/FairwindsOps/polaris/badge.svg">
  </a>
</div>

Fairwinds' Polaris keeps your clusters sailing smoothly. It runs a variety of checks to ensure that
Kubernetes pods and controllers are configured using best practices, helping you avoid
problems in the future.

Polaris can be run in three different modes:
* As a [dashboard](https://polaris.docs.fairwinds.com/dashboard), so you can audit what's running inside your cluster.
* As an [admission controller](https://polaris.docs.fairwinds.com/admission-controller), so you can automatically reject workloads that don't adhere to your organization's policies.
* As a [command-line tool](https://polaris.docs.fairwinds.com/infrastructure-as-code), so you can test local YAML files, e.g. as part of a CI/CD process.

<p align="center">
  <img src="https://polaris.docs.fairwinds.com/img/architecture.svg" alt="Polaris Architecture" width="550"/>
</p>

## Documentation
Check out the [documentation at docs.fairwinds.com](https://polaris.docs.fairwinds.com)

## Join the Fairwinds Open Source Community

The goal of the Fairwinds Community is to exchange ideas, influence the open source roadmap, and network with fellow Kubernetes users. [Chat with us on Slack](https://join.slack.com/t/fairwindscommunity/shared_invite/zt-e3c6vj4l-3lIH6dvKqzWII5fSSFDi1g) or [join the user group](https://www.fairwinds.com/open-source-software-user-group) to get involved!


## Other Projects from Fairwinds

Enjoying Polaris? Check out some of our other projects:
* [Goldilocks](https://github.com/FairwindsOps/Goldilocks) - Right-size your Kubernetes Deployments by compare your memory and CPU settings against actual usage
* [Pluto](https://github.com/FairwindsOps/Pluto) - Detect Kubernetes resources that have been deprecated or removed in future versions
* [Nova](https://github.com/FairwindsOps/Nova) - Check to see if any of your Helm charts have updates available
* [rbac-manager](https://github.com/FairwindsOps/rbac-manager) - Simplify the management of RBAC in your Kubernetes clusters

## Fairwinds Insights
<p align="center">
  <a href="https://www.fairwinds.com/polaris-user-insights-demo?utm_source=polaris&utm_medium=ad&utm_campaign=polarisad">
    <img src="https://polaris.docs.fairwinds.com/img/insights-banner.png" alt="Fairwinds Insights" width="550"/>
  </a>
</p>

If you're interested in running Polaris in multiple clusters,
tracking the results over time, integrating with Slack, Datadog, and Jira,
or unlocking other functionality, check out
[Fairwinds Insights](https://www.fairwinds.com/polaris-user-insights-demo?utm_source=polaris&utm_medium=polaris&utm_campaign=polaris), a platform for auditing and enforcing policy in Kubernetes clusters.
