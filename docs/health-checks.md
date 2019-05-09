# Health Checks

Polaris supports validating the presence of readiness and liveness probes in pods.

key | default | description
----|---------|------------
`healthChecks.readinessProbeMissing` | `warning` | Fails when a readiness probe is not configured for a pod.
`healthChecks.livenessProbeMissing` | `warning` | Fails when a liveness probe is not configured for a pod.

## Background

Readiness and liveness probes can help maintain the health of applications running inside Kubernetes. By default, Kubernetes only knows whether or not a process is running, not if it's healthy. Properly configured readiness and liveness probes will also be able to ensure the health of an application.

Readiness probes are designed to ensure that an application has reached a "ready" state. In many cases there is a period of time between when a webserver process starts and when it is ready to receive traffic. A readiness probe can ensure the traffic is not sent to a pod until it is actually ready to receive traffic.

Liveness probes are designed to ensure that an application stays in a healthy state. When a liveness probe fails, the pod will be restarted.

## Further Reading

- [Kubernetes Docs: Configure Livenss and Readiness Probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-probes/)
- [Utilizing Kubernetes Liveness and Readiness Probes to Automatically Recover From Failure](https://medium.com/spire-labs/utilizing-kubernetes-liveness-and-readiness-probes-to-automatically-recover-from-failure-2fe0314f2b2e)
- [Kubernetes Liveness and Readiness Probes: How to Avoid Shooting Yourself in the Foot](https://blog.colinbreck.com/kubernetes-liveness-and-readiness-probes-how-to-avoid-shooting-yourself-in-the-foot/)
