# Images

Polaris supports a number of checks related to the image specified by pods.

key | default | description
----|---------|------------
`images.tagNotSpecified` | `error` | Fails when an image tag is either not specified or `latest`.
`images.pullPolicyNotAlways` | `ignore` | Fails when an image pull policy is not `always`.

## Background

Docker's `latest` tag is applied by default to images where a tag hasn't been specified. Not specifying a specific version of an image can lead to a wide variety of problems. The underlying image could include unexpected breaking changes that break your application whenever the latest image is pulled. Reusing the same tag for multiple versions of an image can lead to different nodes in the same cluster having different versions of an image, even if the tag is identical.

Related to that, relying on cached versions of a Docker image can become a security vulnerability. By default, an image will be pulled if it isn't already cached on the node attempting to run it. This can result in variations in images that are running per node, or potentially provide a way to gain access to an image without having direct access to the ImagePullSecret. With that in mind, it's often better to ensure the a pod has `pullPolicy: Always` specified, so images are always pulled directly from their source. This is not a check enabled by default with Polaris as organizations may not wish to add the overhead involved with pulling images for each pod.

## Further Reading

- [What's Wrong With The Docker :latest Tag?](https://vsupalov.com/docker-latest-tag/)

- [Kubernetes’ AlwaysPullImages Admission Control — the Importance, Implementation, and Security Vulnerability in its Absence](https://medium.com/@trstringer/kubernetes-alwayspullimages-admission-control-the-importance-implementation-and-security-d83ff3815840)
