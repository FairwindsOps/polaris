---
meta:
  - name: description
    content: "Fairwinds Polaris | Documentation: Polaris can be run as an admission controller that acts as a validating webhook."
---
# Admission Controller
> Want to manage the Admission Controller across multiple clusters? Check out
> [Fairwinds Insights](https://www.fairwinds.com/fairwinds-polaris-upgrade)

Polaris can be run as an admission controller that acts as a validating webhook.
This accepts the same configuration as the dashboard, and can run the same validations.

The webhook will reject any workloads that trigger a danger-level check.
This is indicative of the greater goal of Polaris, not just to encourage better
configuration through dashboard visibility, but to actually enforce it with this webhook.

Note that Polaris will not alter your workloads, only block workloads that don't conform to the configured policies.

## Installation
A valid TLS certificate is required for the Polaris Validating Webhook. If you have cert-manager installed in your cluster then the install methods below will work.

If you don't use cert-manager, you'll need to:

* Supply a CA Bundle with the `webhook.caBundle`
* Create a TLS secret in your cluster with a valid certificate that uses that CA
* Pass the name of that secret with the webhook.secretName parameter.

### kubectl
```bash
kubectl apply -f https://github.com/fairwindsops/polaris/releases/latest/download/webhook.yaml
```

### Helm
```bash
helm repo add fairwinds-stable https://charts.fairwinds.com/stable
helm upgrade --install polaris fairwinds-stable/polaris --namespace polaris --create-namespace \
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
