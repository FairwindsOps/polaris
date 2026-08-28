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
`automountServiceAccountToken` | `warning` | Fails when `automountServiceAccountToken` is automounted.
`hostIPCSet` | `danger` | Fails when `hostIPC` attribute is configured.
`hostPIDSet` | `danger` | Fails when `hostPID` attribute is configured.
`linuxHardening` | `danger` | Fails when neither `AppArmor`, `Seccomp`, `SELinux`, or dropping Linux Capabilities is in use.
`notReadOnlyRootFilesystem` | `warning` | Fails when `securityContext.readOnlyRootFilesystem` is not true.
`privilegeEscalationAllowed` | `danger` | Fails when `securityContext.allowPrivilegeEscalation` is true.
`runAsRootAllowed` | `warning` | Fails when `securityContext.runAsNonRoot` is not true.
`runAsPrivileged` | `danger` | Fails when `securityContext.privileged` is true.
`insecureCapabilities` | `warning` | Fails when `securityContext.capabilities` includes one of the capabilities [listed here](https://github.com/FairwindsOps/polaris/tree/master/pkg/config/checks/insecureCapabilities.yaml)
`dangerousCapabilities` | `danger` | Fails when `securityContext.capabilities` includes one of the capabilities [listed here](https://github.com/FairwindsOps/polaris/tree/master/pkg/config/checks/dangerousCapabilities.yaml)
`hostNetworkSet` | `warning` | Fails when `hostNetwork` attribute is configured.
`hostPortSet` | `warning` | Fails when `hostPort` attribute is configured.
`tlsSettingsMissing` | `warning` | Fails when an Ingress lacks TLS settings.
`gatewayTLSMissing` | `warning` | Fails when an HTTPS, GRPC, or terminating TLS Gateway listener lacks certificate references.
`gatewayAllowedRoutesAll` | `warning` | Fails when a Gateway listener allows Routes from every namespace.
`gatewayInsecureFrontendValidation` | `warning` | Fails when Gateway frontend client certificate validation allows insecure fallback.
`gatewayCrossNamespaceCertificateRef` | `warning` | Fails when a Gateway references a certificate in another namespace without a matching ReferenceGrant. Cluster audits only.
`httpRouteWildcardOrEmptyHost` | `warning` | Fails when an HTTPRoute omits hostnames or uses a wildcard hostname.
`httpRouteInsecureListener` | `warning` | Fails when an HTTPRoute serves application traffic over HTTP without a full HTTPS redirect. Cluster audits only.
`httpRouteCrossNamespaceBackendRef` | `warning` | Fails when an HTTPRoute references a backend in another namespace without a matching ReferenceGrant. Cluster audits only.
`httpRouteBackendTLSMissing` | `warning` | Fails when an HTTPRoute TLS backend lacks a BackendTLSPolicy or kgateway BackendConfigPolicy. Cluster audits only.
`kgatewayBackendTLSVerificationDisabled` | `warning` | Fails when a kgateway BackendConfigPolicy disables TLS certificate verification.
`sensitiveContainerEnvVar` | `danger` | Fails when the container sets potentially sensitive environment variables.
`sensitiveConfigmapContent` | `danger` | Fails when potentially sensitive content is detected in the ConfigMap keys or values.
`missingNetworkPolicy` | `warning`
`clusterrolePodExecAttach` | `danger` | Fails when the ClusterRole allows Pods/exec or pods/attach.
`rolePodExecAttach` | `danger` | Fails when the Role allows Pods/exec or pods/attach.
`clusterrolebindingPodExecAttach` | `danger` | Fails when the ClusterRoleBinding references a ClusterRole that allows Pods/exec, allows pods/attach, or that does not exist.
`rolebindingRolePodExecAttach` | `danger` | Fails when the RoleBinding references a Role that allows Pods/exec, allows pods/attach, or that does not exist.
`rolebindingClusterRolePodExecAttach` | `danger` | Fails when the RoleBinding references a ClusterRole that allows Pods/exec, allows pods/attach, or that does not exist.
`clusterrolebindingClusterAdmin` | `danger` | Fails when the ClusterRoleBinding references the default cluster-admin ClusterRole or one with wildcard permissions.
`rolebindingClusterAdminClusterRole` | `danger` | Fails when the RoleBinding references the default cluster-admin ClusterRole or one with wildcard permissions.
`rolebindingClusterAdminRole` | `danger` | Fails when the RoleBinding references a Role with wildcard permissions.

## Background

Securing workloads in Kubernetes is an important part of overall cluster security. The overall goal should be to ensure that containers are running with as minimal privileges as possible. This includes avoiding privilege escalation, not running containers with a root user, not giving excessive access to the host network, and using read only file systems wherever possible.

### Gateway API

Gateway API separates listeners, routes, and backend TLS policy across different resources. Polaris checks standard `Gateway` and `HTTPRoute` resources for listener TLS, namespace isolation, host specificity, HTTPS redirects, cross-namespace authorization, and backend TLS. These checks work with conformant implementations such as kgateway.

`httpRouteBackendTLSMissing` also recognizes kgateway's `Backend` and `BackendConfigPolicy` resources. It identifies TLS backends from ports 443 and 8443, Service port names and `appProtocol`, and kgateway static Backend ports. `kgatewayBackendTLSVerificationDisabled` checks the kgateway-specific `insecureSkipVerify` setting. Authentication, authorization, and rate-limiting requirements are organization-specific and should be implemented as custom checks.

Checks marked "Cluster audits only" need related resources that are not available when Polaris evaluates a single admission request. They pass without a resource provider rather than rejecting an object without enough context.

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
