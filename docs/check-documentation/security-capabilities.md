# Security Capabilities

Polaris supports a number of checks to ensure pods are running with a limited set of capabilities. Under `security.capabilities`, there are `danger` and `warning` sections indicating the severity of failures for the following checks.

key | default | description
----|---------|------------
`security.capabilities.danger.ifAnyAdded` | [`SYS_ADMIN`, `NET_ADMIN`, `ALL`] | Fails when any of the listed capabilities have been added.
`security.capabilities.danger.ifAnyAddedBeyond` | `nil` | Fails when any capabilities have been added beyond the specified list.
`security.capabilities.danger.ifAnyNotDropped` | `nil` | Fails when any of the listed capabilities have not been dropped.
`security.capabilities.warning.ifAnyAdded` | `nil` | Fails when any of the listed capabilities have been added.
`security.capabilities.warning.ifAnyAddedBeyond` | [`CHOWN`, `DAC_OVERRIDE`, `FSETID`, `FOWNER`, `MKNOD`, `NET_RAW`, `SETGID`, `SETUID`, `SETFCAP`, `SETPCAP`, `NET_BIND_SERVICE`, `SYS_CHROOT`, `KILL`,`AUDIT_WRITE`] | Fails when any capabilities have been added beyond the specified list.
`security.capabilities.warning.ifAnyNotDropped` | `nil` | Fails when any of the listed capabilities have not been dropped.

## Background

Linux Capabilities allow you to specify privileges for a process at a granular level. The [default list of capabilities](https://github.com/moby/moby/blob/master/oci/defaults.go#L15) included with a container are already fairly minimal, but often can be further restricted.

With Kubernetes configuration, these capabilities can be added or removed by adjusting `securityContext.capabilities`.

## Further Reading

- [Kubernetes Docs: Set capabilities for a Container](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/#set-capabilities-for-a-container)
- [Linux Programmer's Manual: Capabilities](http://man7.org/linux/man-pages/man7/capabilities.7.html)
