# Setup polaris action

This action downloads a version of [polaris](https://github.com/FairwindsOps/polaris) and adds it to the path. It makes the [polaris cli](https://polaris.docs.fairwinds.com/infrastructure-as-code) ready to use in following steps of the same job.

## Inputs

### `version`

The release version to fetch. This has to be in the form `<tag_name>`.

## Outputs

### `version`

The version number of the release tag.

## Example usage

```yaml
uses: FairwindsOps/polaris
with:
  version: "3.0.3"
```

And inside the job:

```yaml
steps:
  - uses: actions/checkout@v2
  - name: Setup polaris
    uses: FairwindsOps/polaris
    with:
      version: 3.0.3

  - name: Use command
    run: polaris version
```