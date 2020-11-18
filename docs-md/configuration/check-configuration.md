# Checks
Each check can be assigned a `severity`. Only checks with a severity of `danger` or `warning` will be validated. The results of these validations are visible on the dashboard. In the case of the validating webhook, only failures with a severity of `danger` will result in a change being rejected.

Polaris validation checks fall into several different categories:

- [Security](check-documentation/security.md)
- [Reliability](check-documentation/reliability.md)
- [Efficiency](check-documentation/efficiency.md)


