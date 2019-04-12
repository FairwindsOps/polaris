# Networking

Fairwinds supports a number of checks related to pod networking.

key | default | description
----|---------|------------
`networking.hostNetworkSet` | `warning` | Fails when `hostNetwork` attribute is configured.
`networking.hostPortSet` | `warning` | Fails when `hostPort` attribute is configured.


## Background

Although Kubernetes allows you to deploy a pod with access to the host network namespace, it's rarely a good idea. A pod running with the `hostNetwork` attribute enabled will have access to the loopback device, services listening on localhost, and could be used to snoop on network activity of other pods on the same node. There are certain examples where setting `hostNetwork` to true is required, such as deploying a networking plugin like Flannel.

Setting the `hostPort` attribute on a container will ensure that it is accessible on that specific port on each node it is deployed to. Unfortunately when this is specified, it limits where a pod can actually be scheduled in a cluster.


## Further Reading

- [Kubernetes Docs: Configuration Best Practices](https://kubernetes.io/docs/concepts/configuration/overview/#services)

- [Accessing Kubernetes Pods from Outside of the Cluster](http://alesnosek.com/blog/2017/02/14/accessing-kubernetes-pods-from-outside-of-the-cluster/)