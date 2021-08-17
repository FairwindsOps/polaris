---
meta:
  - name: description
    content: "Fairwinds Polaris | Documentation on setting checks by severity "
---
# Check Settings
Each check can be assigned a `severity`. Only checks with a severity of `danger` or `warning` will be validated. The results of these validations are visible on the dashboard. In the case of the validating webhook, only failures with a severity of `danger` will result in a change being rejected.

Polaris validation checks fall into several different categories:

- [Security](/checks/security.md)
- [Reliability](/checks/reliability.md)
- [Efficiency](/checks/efficiency.md)

To change the default severity levels, or to turn checks on or off, you can create your own `config.yaml`:
```yaml
checks:
  tagNotSpecified: ignore
  runAsRootAllowed: danger
  pullPolicyNotAlways: warning
```

