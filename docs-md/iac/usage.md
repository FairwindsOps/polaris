# Audit your IaC

Polaris can be used on the command line to audit local files.
This is particularly helpful for running Polaris against your infrastructure-as-code as part of a
CI/CD pipeline. Use the available [command line flags](#running-in-a-ci-pipeline)
to cause CI/CD to fail if your Polaris score drops below a certain threshold, or if any danger-level issues arise.

#### Running in a CI pipeline
You can tell the CLI to set an exit code if it detects certain issues with your
YAML files.
For example, to fail if polaris detects *any* danger-level issues, or if the score drops below 90%:
```bash
polaris audit --audit-path ./deploy/ \
  --set-exit-code-on-danger \
  --set-exit-code-below-score 90
```

