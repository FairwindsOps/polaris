# Security

Polaris supports a number of checks related to security.

key | default | description
----|---------|------------
`security.hostIPCSet` | `error` | Fails when `hostIPC` attribute is configured.
`security.hostPIDSet` | `error` | Fails when `hostPID` attribute is configured.
`security.notReadOnlyRootFileSystem` | `warning` | Fails when `securityContext.readOnlyRootFilesystem` is not true.
`security.privilegeEscalationAllowed` | `error` | Fails when `securityContext.allowPrivilegeEscalation` is true.
`security.runAsRootAllowed` | `error` | Fails when `securityContext.runAsNonRoot` is not true.
`security.runAsPrivileged` | `error` | Fails when `securityContext.privileged` is true.

## Security Capabilities

Additional validations are available to ensure pods are running with a limited set of capabilities. More information is available in our [Security Capabilities documentation](security-capabilities.md).

## Background

Securing workloads in Kubernetes is an important part of overall cluster security. The overall goal should be to ensure that containers are running with as minimal privileges as possible. This includes avoiding privilege escalation, not running containers with a root user, and using read only file systems wherever possible.

Much of this configuration can be found in the `securityContext` attribute for both Kubernetes pods and containers. Where configuration is available at both a pod and container level, Polaris validates both.

## Further Reading
- [Kubernetes Docs: Configure a Security Context for a Pod or Container](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)
- [KubeCon 2018 Keynote: Running with Scissors](https://www.youtube.com/watch?v=ltrV-Qmh3oY)
- [Kubernetes Security Book](https://kubernetes-security.info/)
