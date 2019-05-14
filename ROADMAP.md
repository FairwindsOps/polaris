# Polaris Roadmap
We plan to continue expanding the list of checks available to Polaris users,
as well as the project's overall functionality.

If you have ideas for a new check, or for new Polaris features,
you can submit a pull request for this file, or open an issue in GitHub.

## Features
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

## Ability to override checks
We plan to add the ability to override Polaris checks for particular namespaces
and workloads. This may be something that is set in the Polaris configuration,
or an annotation that can be added to the resource definition.

This is important, as certain workloads have good reason to skip particular Polaris
checks - for instance, the `dns-controller` deployment in `kube-system` needs to have
the host network configured.

## Checks
These are checks that we plan to implement. If you have requests or ideas,
let us know! You can submit a pull request for this file, or open an issue in GitHub.

### Images
* List of allowed/disallowed registries

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

