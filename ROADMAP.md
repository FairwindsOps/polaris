# Polaris Roadmap
We plan to continue expanding the list of checks available to Polaris users,
as well as the project's overall functionality.

If you have ideas for a new check, or for new Polaris features,
you can submit a pull request for this file, or open an issue in GitHub.

## Check non-controllers
We would like to implement checks for non-controller types, such as:
* Ingresses
* Services
* RBAC roles/bindings

## Expand list of checks
We'd like to increase our library of checks. Not all checks need to be enabled by default,
so anything that would be useful across different organizations is worth checking in.

## Support more controllers in the validating webhook
Currently the webhook only checks for a fixed set of controllers.

While we can't listen for _all_ possible controller types, we can block Pods that violate policy.
We've chosen not to do this for now, because it could prevent existing controllers from scaling.

However, there may be a way to check for the owner's age, in order to determine if it's a new
controller or a pre-existing controller.
