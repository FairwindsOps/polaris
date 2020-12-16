# Exemptions
Sometimes a workload really does need to do things that Polaris considers insecure. For instance,
many of the `kube-system` workloads need to run as root, or need access to the host network. In these
cases, we can add **exemptions** to allow the workload to pass Polaris checks.

Exemptions can be added two ways: by annotating a controller, or editing the Polaris config.

## Annotations
To exempt a controller from all checks via annotations, use the annotation `polaris.fairwinds.com/exempt=true`, e.g.
```
kubectl annotate deployment my-deployment polaris.fairwinds.com/exempt=true
```

To exempt a controller from a particular check via annotations, use an annotation in the form of `polaris.fairwinds.com/<check>-exempt=true`, e.g.
```
kubectl annotate deployment my-deployment polaris.fairwinds.com/cpuRequestsMissing-exempt=true
```

## Config

To exempt a controller via the config, you have to specify a namespace (optional), a list of controller names, and a list of rules, e.g.
```yaml
exemptions:
  # exemption valid for kube-system namespace
  - namespace: kube-system
    controllerNames:
      - dns-controller
    rules:
      - hostNetworkSet
  # exemption valid in all namespaces
  - controllerNames:
      - dns-controller
    rules:
      - hostNetworkSet
```

