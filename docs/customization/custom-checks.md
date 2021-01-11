# Custom Checks
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
strings like `1000m` and `1G`. You can see an example in [the extended config](https://github.com/FairwindsOps/polaris/tree/master/examples/config-full.yaml)

There are additional examples in the [checks folder](https://github.com/FairwindsOps/polaris/tree/master/checks).

