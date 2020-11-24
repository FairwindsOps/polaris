<div align="center">
  <img src="/docs-md/.vuepress/public/img/polaris-logo.png" alt="Polaris Logo" />
  <br>

  [![Version][version-image]][version-link] [![CircleCI][circleci-image]][circleci-link] [![Go Report Card][goreport-image]][goreport-link]
</div>

[version-image]: https://img.shields.io/static/v1.svg?label=Version&message=1.2.0&color=239922
[version-link]: https://github.com/FairwindsOps/polaris

[goreport-image]: https://goreportcard.com/badge/github.com/FairwindsOps/polaris
[goreport-link]: https://goreportcard.com/report/github.com/FairwindsOps/polaris

[circleci-image]: https://circleci.com/gh/FairwindsOps/polaris.svg?style=svg
[circleci-link]: https://circleci.com/gh/FairwindsOps/polaris.svg

Fairwinds' Polaris keeps your clusters sailing smoothly. It runs a variety of checks to ensure that
Kubernetes pods and controllers are configured using best practices, helping you avoid
problems in the future. Polaris can be run in a few different modes:

Polaris can be run in three different modes:
* As a [dashboard](https://polaris.docs.fairwinds.com/dashboard), so you can audit what's running inside your cluster.
* As an [admission controller](https://polaris.docs.fairwinds.com/admission-controller), so you can automatically reject workloads that don't adhere to your organization's policies.
* As a [command-line tool](https://polaris.docs.fairwinds.com/infrastructure-as-code), so you can test local YAML files, e.g. as part of a CI/CD process.

**Want to learn more?** Reach out on [the Slack channel](https://fairwindscommunity.slack.com/messages/polaris) ([request invite](https://join.slack.com/t/fairwindscommunity/shared_invite/zt-e3c6vj4l-3lIH6dvKqzWII5fSSFDi1g)), send an email to `opensource@fairwinds.com`, or join us for [office hours on Zoom](https://fairwindscommunity.slack.com/messages/office-hours)


## Documentation
Check out the [documentation at docs.fairwinds.com](https://polaris.docs.fairwinds.com), or view the [markdown](./docs-md)

## Integration with Fairwinds Insights
[Fairwinds Insights](https://www.fairwinds.com/insights?utm_campaign=Hosted%20Polaris%20&utm_source=polaris&utm_term=polaris&utm_content=polaris)
is a platform for auditing Kubernetes clusters and enforcing policy. If you'd like to:
* manage Polaris across a fleet of clusters
* track findings over time
* send results to services like Slack and Datadog
* add additional checks from tools like
[Trivy](https://github.com/aquasecurity/trivy),
[Goldilocks](https://github.com/FairwindsOps/goldilocks/), and
[OPA](https://www.openpolicyagent.org)

you can sign up for a [free account here](https://insights.fairwinds.com?source=polaris).

## Contributing
PRs welcome! Check out the [Contributing Guidelines](docs-md/contributing.md) and [Code of Conduct](docs-md/code-of-conduct.md) for more information.

## Further Information
A history of changes to this project can be viewed in the [Changelog](docs-md/changelog.md)

If you'd like to learn more about Polaris, or if you'd like to speak with
a Kubernetes expert, you can contact `info@fairwinds.com` or [visit our website](https://fairwinds.com)

---

<p align="center">
  <img src="/docs-md/.vuepress/public/img/dashboard-screenshot.png" alt="Polaris Dashboard" width="550"/>
</p>

