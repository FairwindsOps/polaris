# Resources

Polaris supports a number of checks related to CPU and Memory requests and limits.

## Presence Checks

To simplify ensure that these values have been set, the following attributes are available:

key | default | description
----|---------|------------
`resources.cpuRequestsMissing` | `warning` | Fails when `resources.requests.cpu` attribute is not configured.
`resources.memoryRequestsMissing` | `warning` | Fails when `resources.requests.memory` attribute is not configured.
`resources.cpuLimitsMissing` | `warning` | Fails when `resources.limits.cpu` attribute is not configured.
`resources.memoryLimitsMissing` | `warning` | Fails when `resources.limits.memory` attribute is not configured.

## Range Checks

Polaris can also verify that those values fall within a certain range. These checks are not enabled by default, and as such do not have default values. The `cpuRequestRanges`, `cpuLimitRanges`, `memoryRequestRanges`, and `memoryLimitRanges` all support the following attributes:

key | description
----|------------
`warning.below` | Warn when resource is below this value (or not defined)
`warning.above` | Warn when resource is above this value
`error.below` | Error when resource is below this value (or not defined)
`error.above` | Error when resource is above this value

## Background

Configuring resource requests and limits for containers running in Kubernetes is an important best practice to follow. Setting appropriate resource requests will ensure that all your applications have sufficient compute resources. Setting appropriate resource limits will ensure that your applications do not consume too many resources.

Having these values appropriately configured ensures that:

* Cluster autoscaling can function as intended. New nodes are scheduled once pods are unable to be scheduled on an existing node due to insufficient resources. This will not happen if resource requests are not configured.

* Each container has sufficient access to compute resources. Without resource requests, a pod may be scheduled on a node that is already overutilized. Without resource limits, a single poorly behaving pod could utilize the majority of resources on a node, significantly impacting the performance of other pods on the same node.

## Further Reading

- [Kubernetes Docs: Managing Compute Resources for Containers](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/)

- [Kubernetes best practices: Resource requests and limits](https://cloud.google.com/blog/products/gcp/kubernetes-best-practices-resource-requests-and-limits)

- [Vertical Pod Autoscaler (can automatically set resource requests and limits)](https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler)
