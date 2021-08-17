---
meta:
  - name: description
    content: "Fairwinds Polaris | Exemptions"
---
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

To add exemptions via the config, you have to specify at least one or more of the following: 
- A namespace
- A list of controller names
- A list of container names

You can also specify a list of particular rules. If no rules are specified then every rule is exempted. 

Controller names and container names are matched as a prefix, so an empty string will match every controller or container respectively.

For example:
```yaml
exemptions:
  # exemption valid for all rules on all containers in all controllers in default namespace
  - namespace: default
  # exemption valid for hostNetworkSet rule on all containers in dns-controller controller in kube-system namespace
  - namespace: kube-system
    controllerNames:
      - dns-controller
    rules:
      - hostNetworkSet
  # exemption valid for hostNetworkSet rule on all containers in dns-controller controller in all namespaces
  - controllerNames:
      - dns-controller
    rules:
      - hostNetworkSet
  # exemption valid for hostNetworkSet rule on coredns container in all controllers in kube-system namespace
  - namespace: kube-system
  - containerNames:
      - coredns
    rules:
      - hostNetworkSet
```

