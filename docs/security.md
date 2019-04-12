# Security

Fairwinds supports a number of checks related to security.

key | default | description
----|---------|------------
`security.hostIPCSet` | `error` | Fails when `hostIPC` attribute is configured.
`security.hostPIDSet` | `error` | Fails when `hostPID` attribute is configured.
`security.notReadOnlyRootFileSystem` | `warning` | Fails when `securityContext.readOnlyRootFilesystem` is not true.
`security.privilegeEscalationAllowed` | `error` | Fails when `securityContext.allowPrivilegeEscalation` is true.
`security.runAsRootAllowed` | `error` | Fails when `securityContext.runAsNonRoot` is not true.
`security.runAsPrivileged` | `error` | Fails when `securityContext.privileged` is true.

## Security Capabilities

Additional validations are available to ensure pods are running with a limited set of capabilities. Under `security.capabilities`, there are `error` and `warning` sections indicating the severity of failures for the following checks.

key | default | description
----|---------|------------
`security.capabilities.error.ifAnyAdded` | [`SYS_ADMIN`, `NET_ADMIN`, `ALL`] | Fails when any of the listed capabilities have been added.
`security.capabilities.error.ifAnyAddedBeyond` | `nil` | Fails when any capabilities have been added beyond the specified list.
`security.capabilities.error.ifAnyNotDropped` | `nil` | Fails when any of the listed capabilities have not been dropped.
`security.capabilities.warning.ifAnyAdded` | `nil` | Fails when any of the listed capabilities have been added.
`security.capabilities.warning.ifAnyAddedBeyond` | [`CHOWN`, `DAC_OVERRIDE`, `FSETID`, `FOWNER`, `MKNOD`, `NET_RAW`, `SETGID`, `SETUID`, `SETFCAP`, `SETPCAP`, `NET_BIND_SERVICE`, `SYS_CHROOT`, `KILL`,`AUDIT_WRITE`] | Fails when any capabilities have been added beyond the specified list.
`security.capabilities.warning.ifAnyNotDropped` | `nil` | Fails when any of the listed capabilities have not been dropped.

## Background

TODO

## Further Reading
- [Kubernetes Docs: Configure a Security Context for a Pod or Container](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)

- [Kubernetes Docs: Set capabilities for a Container](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/#set-capabilities-for-a-container)

- [Linux Programmer's Manual: Capabilities](http://man7.org/linux/man-pages/man7/capabilities.7.html)
