# Polaris Roadmap
We plan to continue expanding the list of checks available to Polaris users,
as well as the project's overall functionality.

If you have ideas for a new check, or for new Polaris features,
you can submit a pull request for this file, or open an issue in GitHub.

## Q3 2019
Below is a list of work we plan to get done this quarter. Some more details can be found under
[Future Work](#future_work), or
[in the milestone](https://github.com/FairwindsOps/polaris/milestone/1)
* Rule exceptions - we'd like to provide a way of overriding Polaris checks on individual controllers, e.g. through annotations
* More controller types - we plan to implement checks for more than just deployments
* Image repo checks - we plan to provide a check that ensures all repos conform to a set of user-defined patterns
* OPA integration (investigate only) - we plan to investigate the potential use of OPA to let users define custom Polaris checks

## Future Work
### Validating Webhook
The validating webhook rejects incoming workloads if they cause `error`-level
issues, as defined by the Polaris configuration.

The webhook server is currently functional, but largely untested. Because
the validating webhook has the potential to cause headaches for users,
we have marked it as `beta` for now.

We will continue to gather feedback here and will move the webhook out of `beta`
when we feel it's ready.

#### TODO
* Pass `warning`-level messages back to `kubectl` if possible (may require changes
to core k8s or `kubectl`)

### Ability to override checks
We plan to add the ability to override Polaris checks for particular namespaces
and workloads. This may be something that is set in the Polaris configuration,
or an annotation that can be added to the resource definition.

This is important, as certain workloads have good reason to skip particular Polaris
checks - for instance, the `dns-controller` deployment in `kube-system` needs to have
the host network configured.

### Images Registries Check
We'd like to allow users to restrict images to a list of allowed/disallowed registries.
We'll likely allow patterns, e.g. `*.dkr.ecr.*.amazonaws.com`

### New controller types
Currently we only look at Deployments. We plan to also validate these types of controllers:
* StatefulSet
* DaemonSet
* Job
* CronJob
* ReplicaSet

### Check non-controllers
We would like to implement checks for non-controller types, such as:
* Ingress
* Service
* Secret

