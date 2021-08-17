---
meta:
  - name: description
    content: "Fairwinds Polaris | Documentation | Polaris can be used on the command line to audit local Kubernetes manifests stored in YAML files."
---
# Infrastructure as Code
> Want to see results for all your IaC repos in one place? Check out
> [Fairwinds Insights](https://www.fairwinds.com/fairwinds-polaris-upgrade)

Polaris can be used on the command line to audit local Kubernetes manifests stored in YAML files.
This is particularly helpful for running Polaris against your infrastructure-as-code as part of a
CI/CD pipeline. Use the available [command line flags](#running-in-a-ci-pipeline)
to cause CI/CD to fail if your Polaris score drops below a certain threshold, or if any danger-level issues arise.


## Install the CLI
To run Polaris against your YAML manifests, e.g. as part of a Continuous Integration process,
you'll need to install the CLI.

Binary releases can be downloaded from the [releases page](https://github.com/fairwindsops/polaris/releases)
or can be installed with [Homebrew](https://brew.sh/):
```bash
brew tap FairwindsOps/tap
brew install FairwindsOps/tap/polaris
polaris version
```

## Running in a CI pipeline

### Set minimum score for an exit code
You can tell the CLI to set an exit code if it detects certain issues with your YAML files.
For example, to fail if polaris detects *any* danger-level issues, or if the score drops below 90%:
```bash
polaris audit --audit-path ./deploy/ \
  --set-exit-code-on-danger \
  --set-exit-code-below-score 90
```

### Pretty-print results
By default, results are output as JSON. You can get human-readable output with
the `--format=pretty` flag:

```bash
polaris audit --audit-path ./deploy/ \
  --format=pretty
```

You can also disable colors and emoji:
```bash
polaris audit --audit-path ./deploy/ \
  --format=pretty \
  --color=false
```

### Output only showing failed tests
The CLI to gives you ability to display results containing only failed tests. 
For example:
```bash
polaris audit --audit-path ./deploy/ \
  --only-show-failed-tests true
```

### Audit Helm Charts
You can audit helm charts using the `--helm-chart` and `--helm-values` flags:
```
polaris audit \
  --helm-chart ./deploy/chart \
  --helm-values ./deploy/chart/values.yml
```

### As Github Action
#### Setup polaris action

This action downloads a version of [polaris](https://github.com/FairwindsOps/polaris) and adds it to the path. It makes the [polaris cli](https://polaris.docs.fairwinds.com/infrastructure-as-code) ready to use in following steps of the same job.

##### Inputs

###### `version`

The release version to fetch. This has to be in the form `<tag_name>`.

##### Outputs

###### `version`

The version number of the release tag.

##### Example usage

```yaml
uses: fairwindsops/polaris@master
with:
  version: "3.0.3"
```

Example inside a job:

```yaml
steps:
  - uses: actions/checkout@v2
  - name: Setup polaris
    uses: fairwindsops/polaris@master
    with:
      version: 3.0.3

  - name: Use command
    run: polaris version
```
