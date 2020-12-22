# Exemptions
Sometimes a workload really does need to do things that Polaris considers insecure. For instance,
many of the `kube-system` workloads need to run as root, or need access to the host network. In these
cases, we can add **exemptions** to allow the workload to pass Polaris checks.

Exemptions can be added in a few different ways: 
 - Namespace: By editing the Polaris config.
 - Controller: By annotating a controller, or editing the Polaris config.
 - Container: By editing the Polaris config.

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

You can add exemptions by using a combination of namespace, controller names, and container names via the config. You have to specify a list of rules and at least one of the following: a namespace, a list of controller names, or a list of container names, e.g.
```yaml
exemptions:
  # exemption valid in kube-system namespace and dns-controller controller for all containers
  - namespace: kube-system
    controllerNames:
      - dns-controller
    rules:
      - hostNetworkSet
  # exemption valid in all namespaces and dns-controller controller for all containers
  - controllerNames:
      - dns-controller
    rules:
      - hostNetworkSet
  # exemption valid in kube-system namespace and all controllers for coredns container
  - namespace: kube-system
  - containerNames:
      - coredns
    rules:
      - hostNetworkSet
```

