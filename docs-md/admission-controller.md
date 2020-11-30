# Admission Controller
Polaris can be run as an admission controller that acts as a validating webhook.
This accepts the same configuration as the dashboard, and can run the same validations.

The webhook will reject any workloads that trigger a danger-level check.
This is indicative of the greater goal of Polaris, not just to encourage better
configuration through dashboard visibility, but to actually enforce it with this webhook.

Note that Polaris will not alter your workloads, only block workloads that don't conform to the configured policies.

## Installation
### kubectl
```bash
kubectl apply -f https://github.com/fairwindsops/polaris/releases/latest/download/webhook.yaml
```

### Helm
```bash
helm repo add fairwindsops-stable https://charts.fairwindsops.com/stable
helm upgrade --install polaris fairwindsops-stable/polaris --namespace polaris \
  --set webhook.enable=true --set dashboard.enable=false
```

## Workload Types
The webhook comes with built-in support for a handful of known controller types,
such as Deployments, Jobs, and DaemonSets. To add new controller types,
you can set `webhook.rules` in the
[Helm chart](https://github.com/FairwindsOps/charts/tree/master/stable/polaris)

## Warnings
Unfortunately we have not found a way to display warnings as part of `kubectl`
output unless we are rejecting a workload altogether.

This means that any checks with a severity of `warning` will still pass webhook validation,
and the only evidence of that warning will either be in the Polaris dashboard or the
Polaris webhook logs. This will change in a future version of Kubernetes.

