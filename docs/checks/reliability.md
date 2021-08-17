---
meta:
  - name: description
    content: "Fairwinds Polaris | Make sure your Kubernetes workloads are always available, and are running the correct image."
---
# Reliability

These checks help to make sure your workloads are always available,
and are running the correct image.

key | default | description
----|---------|------------
`readinessProbeMissing` | `warning` | Fails when a readiness probe is not configured for a pod.
`livenessProbeMissing` | `warning` | Fails when a liveness probe is not configured for a pod.
`tagNotSpecified` | `danger` | Fails when an image tag is either not specified or `latest`.
`pullPolicyNotAlways` | `warning` | Fails when an image pull policy is not `always`.
`priorityClassNotSet` | `ignore` | Fails when a priorityClassName is not set for a pod.
`multipleReplicasForDeployment` | `ignore` | Fails when there is only one replica for a deployment.
`missingPodDisruptionBudget` | `ignore` 

## Background

Readiness and liveness probes can help maintain the health of applications running inside Kubernetes. By default, Kubernetes only knows whether or not a process is running, not if it's healthy. Properly configured readiness and liveness probes will also be able to ensure the health of an application.

Readiness probes are designed to ensure that an application has reached a "ready" state. In many cases there is a period of time between when a webserver process starts and when it is ready to receive traffic. A readiness probe can ensure the traffic is not sent to a pod until it is actually ready to receive traffic.

Liveness probes are designed to ensure that an application stays in a healthy state. When a liveness probe fails, the pod will be restarted.

Docker's `latest` tag is applied by default to images where a tag hasn't been specified. Not specifying a specific version of an image can lead to a wide variety of problems. The underlying image could include unexpected breaking changes that break your application whenever the latest image is pulled. Reusing the same tag for multiple versions of an image can lead to different nodes in the same cluster having different versions of an image, even if the tag is identical.

Related to that, relying on cached versions of a Docker image can become a security vulnerability. By default, an image will be pulled if it isn't already cached on the node attempting to run it. This can result in variations in images that are running per node, or potentially provide a way to gain access to an image without having direct access to the ImagePullSecret. With that in mind, it's often better to ensure the a pod has `pullPolicy: Always` specified, so images are always pulled directly from their source.

## Further Reading

- [What's Wrong With The Docker :latest Tag?](https://vsupalov.com/docker-latest-tag/)
- [Kubernetes’ AlwaysPullImages Admission Control — the Importance, Implementation, and Security Vulnerability in its Absence](https://medium.com/@trstringer/kubernetes-alwayspullimages-admission-control-the-importance-implementation-and-security-d83ff3815840)
- [Kubernetes Docs: Configure Liveness and Readiness Probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-probes/)
- [Utilizing Kubernetes Liveness and Readiness Probes to Automatically Recover From Failure](https://medium.com/spire-labs/utilizing-kubernetes-liveness-and-readiness-probes-to-automatically-recover-from-failure-2fe0314f2b2e)
- [Kubernetes Liveness and Readiness Probes: How to Avoid Shooting Yourself in the Foot](https://blog.colinbreck.com/kubernetes-liveness-and-readiness-probes-how-to-avoid-shooting-yourself-in-the-foot/)
