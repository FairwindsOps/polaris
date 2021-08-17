---
meta:
  - name: description
    content: "Fairwinds Polaris | Contribution Guidelines"
---
# Contributing

Issues, whether bugs, tasks, or feature requests are essential for keeping Polaris great. We believe it should be as easy as possible to contribute changes that get things working in your environment. There are a few guidelines that we need contributors to follow so that we can keep on top of things.

## Code of Conduct

This project adheres to a [code of conduct](code-of-conduct.md). Please review this document before contributing to this project.

## Sign the CLA
Before you can contribute, you will need to sign the [Contributor License Agreement](https://cla-assistant.io/fairwindsops/polaris).

## Project Structure

Polaris is built on top of [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime). It can run in 3 different modes, a dashboard, a webhook, or a reporter that prints or exports validation results. All of these modes make use of the shared `validator` and `config` packages. Adding new validations is possible by only making additions to those packages.

## Getting Started

We label issues with the ["good first issue" tag](https://github.com/FairwindsOps/polaris/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22) if we believe they'll be a good starting point for new contributors. If you're interested in working on an issue, please start a conversation on that issue, and we can help answer any questions as they come up.

## Setting Up Your Development Environment
### Prerequisites
* A properly configured Golang environment with Go 1.11 or higher
* If you want to see the local changes you make on a Polaris dashboard, you will need access to a Kubernetes cluster defined in `~/.kube/config`

### Installation
* Install the project with `go get github.com/fairwindsops/polaris`
* Change into the polaris directory which is installed at `$GOPATH/src/github.com/fairwindsops/polaris`
* See the dashboard with `go run main.go dashboard`, then open http://localhost:8080/
* See the audit data  `go run main.go audit`. This command shows the audit information on the command line. 

## Running Tests

The following commands are all required to pass as part of Polaris testing:

```bash
go list ./... | grep -v vendor | xargs golint -set_exit_status
go list ./... | grep -v vendor | xargs go vet
go test ./pkg/... -v -coverprofile cover.out
```

### Webhook tests
```bash
kind create cluster --wait=90s --image kindest/node:v1.15.11 --name polaris-test
docker build -t quay.io/fairwinds/polaris:debug . # or use your own registry
docker push quay.io/fairwinds/polaris:debug
helm repo add jetstack https://charts.jetstack.io
kubectl create ns cert-manager
helm install cert-manager jetstack/cert-manager --namespace cert-manager --version 0.16.1 --set "installCRDs=true" --wait
POLARIS_IMAGE=quay.io/fairwinds/polaris:debug ./test/webhook_test.sh
```
to avoid the final cleanup for debugging purposes, you can run
```bash
SKIP_FINAL_CLEANUP=true IMAGE_TAG=debug ./test/webhook_test.sh
```
## Creating a New Issue

If you've encountered an issue that is not already reported, please create a [new issue](https://github.com/FairwindsOps/polaris/issues), choose `Bug Report`, `Feature Request` or `Misc.` and follow the instructions in the template. 


## Creating a Pull Request

Each new pull request should:

- Reference any related issues
- Add tests that show the issues have been solved
- Pass existing tests and linting
- Contain a clear indication of if they're ready for review, or a work in progress
- Be up to date and/or rebased on the master branch

## Creating a new release

### Patch releases
Patch releases only need to change this repo. The Helm chart and deploy scripts
will automatically pull in the latest changes.

If the release involves changes to anything in the `deploy/` folder (e.g. new RBAC permissions),
it needs to be a minor or major release in order to prevent breaking the Helm chart.

1. Create a PR for this repo
    1. Bump the version number in:
        1. main.go
        2. README.md
    2. Update CHANGELOG.md
    3. Merge your PR
2. Tag the latest branch for this repo
    1. Pull the latest commit for the `master` branch (which you just merged in your PR)
    2. Run `git tag $VERSION && git push --tags`
    3. Make sure CircleCI runs successfully for the new tag - this will push images to quay.io and create a release in GitHub
        1. If CircleCI fails, check with Codeowners ASAP

### Minor/Major releases
Minor and major releases need to change both this repository and the
[Helm chart repo](https://github.com/FairwindsOps/charts/).

The steps are:
1. Modify the [Helm chart](https://github.com/FairwindsOps/charts/stable/polaris)
    1. Clone the helm charts repo
        1. `git clone https://github.com/FairwindsOps/charts`
        2. `git checkout -b yourname/update-polaris`
    1. Bump the version number in `stable/polaris/Chart.yaml`
    2. Make any necessary changes to the chart to support the new version of Polaris (e.g. new RBAC permissions)
    3. **Don't merge yet!**
2. Create a PR for this repo
    1. Create a new branch named `yourname/update-version`
    2. Bump the version number in:
        1. main.go
        2. README.md
    3. Regenerate the deployment files. Assuming you've cloned the charts repo to `~/git/charts`:
        1. `CHARTS_DIR=~/git/charts ./scripts/generate-deployment-files.sh`
    4. Update CHANGELOG.md
    5. Merge your PR
3. Tag the latest branch for this repo
    1. Pull the latest for the `master` branch
    2. Run `git tag $VERSION && git push --tags`
    3. Make sure CircleCI runs successfully for the new tag - this will push images to quay.io and create a release in GitHub
        1. If CircleCI fails, check with Codeowners ASAP
4. Create and merge a PR for your changes to the Helm chart
