<div align="center" class="no-border">
  <img src="https://polaris.docs.fairwinds.com/img/polaris-logo.png" alt="Polaris Logo">
  <br>
  <h3>Polaris is an open source policy engine for Kubernetes</h3>
  <a href="https://github.com/FairwindsOps/polaris/releases">
    <img src="https://img.shields.io/github/v/release/FairwindsOps/polaris">
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

Polaris is an open source policy engine for Kubernetes that validates and remediates resource configuration. It includes 30+ built in configuration policies, as well as the ability to build custom policies with JSON Schema. When run on the command line or as a mutating webhook, Polaris can automatically remediate issues based on policy criteria.

Polaris can be run in three different modes:
* As a [dashboard](https://polaris.docs.fairwinds.com/dashboard) - Validate Kubernetes resources against policy-as-code.
* As an [admission controller](https://polaris.docs.fairwinds.com/admission-controller) - Automatically reject or modify workloads that don't adhere to your organization's policies.
* As a [command-line tool](https://polaris.docs.fairwinds.com/infrastructure-as-code) - Incorporate policy-as-code into the CI/CD process to test local YAML files.
<p align="center">
  <img src="https://polaris.docs.fairwinds.com/img/architecture.svg" alt="Polaris Architecture" width="550"/>
</p>

## Documentation
Check out the [documentation at docs.fairwinds.com](https://polaris.docs.fairwinds.com)

<!-- Begin boilerplate -->
## Join the Fairwinds Open Source Community

The goal of the Fairwinds Community is to exchange ideas, influence the open source roadmap,
and network with fellow Kubernetes users.
[Chat with us on Slack](https://join.slack.com/t/fairwindscommunity/shared_invite/zt-2na8gtwb4-DGQ4qgmQbczQyB2NlFlYQQ)
or
[join the user group](https://www.fairwinds.com/open-source-software-user-group) to get involved!

<a href="https://insights.fairwinds.com/auth/register/">
  <img src="https://www.fairwinds.com/hubfs/Doc_Banners/Fairwinds_OSS_User_Group_740x125_v6.png"
  alt="Love Fairwinds Open Source? Automate Fairwinds Open Source for free with Fairwinds Insights. Click to learn more" />
</a>

## Other Projects from Fairwinds

Enjoying Polaris? Check out some of our other projects:
* [Goldilocks](https://github.com/FairwindsOps/Goldilocks) - Right-size your Kubernetes Deployments by compare your memory and CPU settings against actual usage
* [Pluto](https://github.com/FairwindsOps/Pluto) - Detect Kubernetes resources that have been deprecated or removed in future versions
* [Nova](https://github.com/FairwindsOps/Nova) - Check to see if any of your Helm charts have updates available
* [rbac-manager](https://github.com/FairwindsOps/rbac-manager) - Simplify the management of RBAC in your Kubernetes clusters

Or [check out the full list](https://www.fairwinds.com/open-source-software?utm_source=polaris&utm_medium=polaris&utm_campaign=polaris)
## Fairwinds Insights
If you're interested in running Polaris in multiple clusters,
tracking the results over time, integrating with Slack, Datadog, and Jira,
or unlocking other functionality, check out
[Fairwinds Insights](https://fairwinds.com/pricing),
a platform for auditing and enforcing policy in Kubernetes clusters.

<a href="https://fairwinds.com/pricing">
  <img src="https://www.fairwinds.com/hubfs/Doc_Banners/Fairwinds_Polaris_Ad.png" alt="Fairwinds Insights" />
</a>
