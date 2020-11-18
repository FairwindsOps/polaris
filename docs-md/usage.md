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



For more on exit code meanings, see [exit-code docs](exit-codes.md).

