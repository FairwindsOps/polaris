---
meta:
  - name: description
    content: "Fairwinds Polaris | Ensure that CPU and memory settings are configured, so that Kubernetes can schedule your workload effectively"
---
# Efficiency

These checks ensure that CPU and memory settings are configured, so that
Kubernetes can schedule your workload effectively.

## Presence Checks

To simplify ensure that these values have been set, the following attributes are available:

key | default | description
----|---------|------------
`cpuRequestsMissing` | `warning` | Fails when `resources.requests.cpu` attribute is not configured.
`memoryRequestsMissing` | `warning` | Fails when `resources.requests.memory` attribute is not configured.
`cpuLimitsMissing` | `warning` | Fails when `resources.limits.cpu` attribute is not configured.
`memoryLimitsMissing` | `warning` | Fails when `resources.limits.memory` attribute is not configured.

## Background

Configuring resource requests and limits for containers running in Kubernetes is an important best practice to follow. Setting appropriate resource requests will ensure that all your applications have sufficient compute resources. Setting appropriate resource limits will ensure that your applications do not consume too many resources.

Having these values appropriately configured ensures that:

* Cluster autoscaling can function as intended. New nodes are scheduled once pods are unable to be scheduled on an existing node due to insufficient resources. This will not happen if resource requests are not configured.

* Each container has sufficient access to compute resources. Without resource requests, a pod may be scheduled on a node that is already overutilized. Without resource limits, a single poorly behaving pod could utilize the majority of resources on a node, significantly impacting the performance of other pods on the same node.

## Further Reading

- [Kubernetes Docs: Managing Compute Resources for Containers](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/)
- [Kubernetes best practices: Resource requests and limits](https://cloud.google.com/blog/products/gcp/kubernetes-best-practices-resource-requests-and-limits)
- [Vertical Pod Autoscaler (can automatically set resource requests and limits)](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler)
