# Using the Admission Controller
Polaris can be run as an admission controller that acts as a validating webhook.
This accepts the same configuration as the dashboard, and can run the same validations.

The webhook will reject any workloads that trigger a danger-level check.
This is indicative of the greater goal of Polaris, not just to encourage better
configuration through dashboard visibility, but to actually enforce it with this webhook.

Note that Polaris will not alter your workloads, only block workloads that don't conform to the configured policies.


