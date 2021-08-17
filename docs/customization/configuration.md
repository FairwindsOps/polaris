---
meta:
  - name: description
    content: "Fairwinds Polaris | Learn about Polaris defaults and how to customize configurations. "
---
# Configuration

The default Polaris configuration can be [seen here](https://github.com/FairwindsOps/polaris/blob/master/examples/config.yaml).

You can customize the configuration to do things like:
* Turn checks [on and off](checks.md)
* Change the [severity level](checks.md) of checks
* Add new [custom checks](custom-checks.md)
* Add [exemptions](exemptions.md) for particular workloads or namespaces

To pass in your custom configuration, follow the instructions for your environment:

* CLI - set the `--config` argument to point to your `config.yaml`
* Helm - set the `config` variable in your values file
* kubectl - create a ConfigMap with your `config.yaml`, mount it as a volume, and use the `--config` argument in your Deployment

