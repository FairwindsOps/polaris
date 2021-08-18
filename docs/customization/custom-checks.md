---
meta:
  - name: description
    content: "Fairwinds Polaris | Documentation: Create your own checks, you can use JSON Schema"
---
# Custom Checks

If you'd like to create your own checks, you can use [JSON Schema](https://json-schema.org/).
This is how built-in Polaris checks are defined as well - you can see all the built-in checks
in the [checks folder](https://github.com/FairwindsOps/polaris/tree/master/checks) for examples.

If you write a check that could be useful for others, feel free to open a PR to add it in!

## Basic Example
For example, to disallow images from quay.io:

```yaml
checks:
  imageRegistry: warning

customChecks:
  imageRegistry:
    successMessage: Image comes from allowed registries
    failureMessage: Image should not be from disallowed registry
    category: Security
    target: Container
    schema:
      '$schema': http://json-schema.org/draft-07/schema
      type: object
      properties:
        image:
          type: string
          not:
            pattern: ^quay.io
```

## Available Options
All custom checks should go under the `customChecks` field in your Polaris config, keyed by the
check ID. Note that you'll also have to set its severity in the `checks` section of your Polaris config.

* `successMessage` - the message to show when the check succeeds
* `failureMessage` - the message to show when the check fails
* `category` - one of `Security`, `Efficiency`, or `Reliability`
* `target` - specifies the type of resource to check. This can be:
  * a group and kind, e.g. `apps/Deployment` or `networking.k8s.io/Ingress`
  * `Controller`, to check _any_ resource that contains a pod spec (e.g. Deployments, CronJobs, StatefulSets), as well as naked Pods
  * `Pod`, same as `Controller`, but the schema applies to the Pod spec rather than the top-level controller
  * `Container` same as `Controller`, but the schema applies to all Container specs rather than the top-level controller
* `controllers` - if `target` is `Controller`, `Pod` or `Container`, you can use this to change which types of controllers are checked
* `controllers.include` - _only_ check these controllers
* `controllers.exclude` - check all controllers except these
* `containers` - if `target` is `Container`, you can use this to decide if `initContainers`, `containers`, or both should be checked
* `containers.exclude` - can be set to a list including `initContainer` or `container`
* `schema` - the JSON Schema to check against, as a YAML object
* `schemaString` - this JSON Schema to check against, as a YAML or JSON string. See [Templating](#templating) below
  * Note: only _one_ of `schema` and `schemaString` can be specified.
* `additionalSchemas` - see [Multi-Resource Checks](#multi-resource-checks) below
* `additionalSchemaStrings` - see [Multi-Resource Checks](#multi-resource-checks) below
  * Note: only _one_ of `additionalSchemas` and `additionalSchemaStrings` can be specified.

## Checking CPU and Memory
We extend JSON Schema with `resourceMinimum` and `resourceMaximum` fields to help compare memory and CPU resource
strings like `1000m` and `1G`. Here's an example check that memory and CPU falls within a certain range.
```yaml
customChecks:
  resourceLimits:
    containers:
      exclude:
      - initContainer
    successMessage: Resource limits are within the required range
    failureMessage: Resource limits should be within the required range
    category: Resources
    target: Container
    schema:
      '$schema': http://json-schema.org/draft-07/schema
      type: object
      required:
      - resources
      properties:
        resources:
          type: object
          required:
          - limits
          properties:
            limits:
              type: object
              required:
              - memory
              - cpu
              properties:
                memory:
                  type: string
                  resourceMinimum: 100M
                  resourceMaximum: 6G
                cpu:
                  type: string
                  resourceMinimum: 100m
                  resourceMaximum: "2"
```

## Resource Presence
You can test for the presence of a resource in each Namespace. For example, to
ensure an AlertmanagerConfig is in every Namespace:
```yaml
successMessage: Namespace has monitoring configuration
failureMessage: Namespace should have monitoring configuration
category: Security
target: Namespace
schema: {}
additionalSchemas:
  monitoring.coreos.com/AlertmanagerConfig: {}
```

## Templating
You can also utilize go templating in your JSON schema in order to match one field against another.
E.g. here is the built-in check to ensure that the `name` annotation matches the object's name:
```yaml
successMessage: Label app.kubernetes.io/name matches metadata.name
failureMessage: Label app.kubernetes.io/name must match metadata.name
target: Controller
schema:
  '$schema': http://json-schema.org/draft-07/schema
  type: object
  properties:
    metadata:
      type: object
      required: ["labels"]
      properties:
        labels:
          type: object
          required: ["app.kubernetes.io/name"]
          properties:
            app.kubernetes.io/name:
              const: "{{ .metadata.name }}"
```

You can also use the full [Go template syntax](https://golang.org/pkg/text/template/), though
you may need to specify your schema as a string in order to use concepts like `range`. E.g.
this check ensures that at least one of the object's labels is present in `matchLabels`:
```yaml
schemaString: |
  type: object
  properties:
    spec:
      type: object
      required: ["selector"]
      properties:
        selector:
          type: object
          required: ["matchLabels"]
          properties:
            matchLabels:
              type: object
              anyOf:
              {{ range $key, $value := .metadata.labels }}
              - properties:
                  "{{ $key }}":
                    type: string
                    const: {{ $value }}
                required: ["{{ $key }}"]
              {{ end }}
```

## Multi-Resource Checks
You can write checks that span multiple resources. This is helpful for ensuring e.g.
that every Deployment has a PDB or an HPA associated with it.

Here's the check to ensure that every Deployment has a PDB:
```yaml
successMessage: A PodDisruptionBudget is attached
failureMessage: Should have a PodDisruptionBudget
category: Reliability
target: Controller
controllers:
  include:
  - Deployment
schema:
  '$schema': http://json-schema.org/draft-07/schema
  type: object
  properties:
    metadata:
      type: object
      properties:
        labels:
          type: object
          minProperties: 1
additionalSchemaStrings:
  policy/PodDisruptionBudget: |
    type: object
    properties:
      spec:
        type: object
        required: ["selector"]
        properties:
          selector:
            type: object
            required: ["matchLabels"]
            properties:
              matchLabels:
                type: object
                anyOf:
                {{ range $key, $value := .metadata.labels }}
                - properties:
                    "{{ $key }}":
                      type: string
                      const: {{ $value }}
                  required: ["{{ $key }}"]
                {{ end }}
```

## JSON vs YAML
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

