<div align="center">
  <img src="/img/polaris-logo.png" alt="Polaris Logo" class="no-border" />
</div>

[![Version][version-image]][version-link] [![CircleCI][circleci-image]][circleci-link] [![Go Report Card][goreport-image]][goreport-link]

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
* As a [dashboard](/dashboard), so you can audit what's running inside your cluster.
* As an [admission controller](/admission-controller), so you can automatically reject workloads that don't adhere to your organization's policies.
* As a [command-line utility](/infrastructure-as-code), so you can test local YAML files, e.g. as part of a CI/CD process.

**Want to learn more?** Reach out on [the Slack channel](https://fairwindscommunity.slack.com/messages/polaris) ([request invite](https://join.slack.com/t/fairwindscommunity/shared_invite/zt-e3c6vj4l-3lIH6dvKqzWII5fSSFDi1g)), send an email to `opensource@fairwinds.com`, or join us for [office hours on Zoom](https://fairwindscommunity.slack.com/messages/office-hours)

---

**Get more from Polaris** with [Fairwinds Insights](https://www.fairwinds.com/insights?utm_campaign=Hosted%20Polaris%20&utm_source=polaris&utm_term=polaris&utm_content=polaris) -
Insights can help you track Polaris findings over time, send new findings to Slack and Datadog, and integrate other
Kubernetes auditing tools such as
[Trivy](https://github.com/aquasecurity/trivy) and [Goldilocks](https://github.com/FairwindsOps/goldilocks/)

---

# Dashboard Quickstart

```bash
kubectl apply -f https://github.com/FairwindsOps/polaris/releases/latest/download/dashboard.yaml
kubectl port-forward --namespace polaris svc/polaris-dashboard 8080:80
```
With the port forwarding in place, you can open http://localhost:8080 in your browser to view the dashboard.

* * *

## Contributing
PRs welcome! Check out the [Contributing Guidelines](CONTRIBUTING.md),
[Code of Conduct](CODE_OF_CONDUCT.md), and [Roadmap](ROADMAP.md) for more information.

## Further Information
A history of changes to this project can be viewed in the [Changelog](CHANGELOG.md)

If you'd like to learn more about Polaris, or if you'd like to speak with
a Kubernetes expert, you can contact `info@fairwinds.com` or [visit our website](https://fairwinds.com)

## License
Apache License 2.0
