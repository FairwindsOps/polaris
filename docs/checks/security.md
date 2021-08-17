---
meta:
  - name: description
    content: "Fairwinds Polaris | Security Checks Documentation"
---
# Security

These checks are related to security concerns. Workloads that fail these
checks may make your cluster more vulnerable, often by introducing a path
for privilege escalation.

key | default | description
----|---------|------------
`hostIPCSet` | `danger` | Fails when `hostIPC` attribute is configured.
`hostPIDSet` | `danger` | Fails when `hostPID` attribute is configured.
`notReadOnlyRootFilesystem` | `warning` | Fails when `securityContext.readOnlyRootFilesystem` is not true.
`privilegeEscalationAllowed` | `danger` | Fails when `securityContext.allowPrivilegeEscalation` is true.
`runAsRootAllowed` | `warning` | Fails when `securityContext.runAsNonRoot` is not true.
`runAsPrivileged` | `danger` | Fails when `securityContext.privileged` is true.
`insecureCapabilities` | `warning` | Fails when `securityContext.capabilities` includes one of the capabilities [listed here](https://github.com/FairwindsOps/polaris/tree/master/checks/insecureCapabilities.yaml)
`dangerousCapabilities` | `danger` | Fails when `securityContext.capabilities` includes one of the capabilities [listed here](https://github.com/FairwindsOps/polaris/tree/master/checks/dangerousCapabilities.yaml)
`hostNetworkSet` | `warning` | Fails when `hostNetwork` attribute is configured.
`hostPortSet` | `warning` | Fails when `hostPort` attribute is configured.
`tlsSettingsMissing` | `warning` | Fails when an Ingress lacks TLS settings.

## Background

Securing workloads in Kubernetes is an important part of overall cluster security. The overall goal should be to ensure that containers are running with as minimal privileges as possible. This includes avoiding privilege escalation, not running containers with a root user, not giving excessive access to the host network, and using read only file systems wherever possible.

A pod running with the `hostNetwork` attribute enabled will have access to the loopback device, services listening on localhost, and could be used to snoop on network activity of other pods on the same node. There are certain examples where setting `hostNetwork` to true is required, such as deploying a networking plugin like Flannel.

Setting the `hostPort` attribute on a container will ensure that it is accessible on that specific port on each node it is deployed to. Unfortunately when this is specified, it limits where a pod can actually be scheduled in a cluster.

Much of this configuration can be found in the `securityContext` attribute for both Kubernetes pods and containers. Where configuration is available at both a pod and container level, Polaris validates both.

## Further Reading
- [Kubernetes Docs: Configure a Security Context for a Pod or Container](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)
- [KubeCon 2018 Keynote: Running with Scissors](https://www.youtube.com/watch?v=ltrV-Qmh3oY)
- [Kubernetes Security Book](https://kubernetes-security.info/)
- [Kubernetes Docs: Set capabilities for a Container](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/#set-capabilities-for-a-container)
- [Linux Programmer's Manual: Capabilities](http://man7.org/linux/man-pages/man7/capabilities.7.html)
- [Kubernetes Docs: Configuration Best Practices](https://kubernetes.io/docs/concepts/configuration/overview/#services)
- [Accessing Kubernetes Pods from Outside of the Cluster](http://alesnosek.com/blog/2017/02/14/accessing-kubernetes-pods-from-outside-of-the-cluster/)
