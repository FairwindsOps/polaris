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
To create a new release, simply tag this repo with the new version.

For major and minor releases, don't forget to update the Helm chart at
https://github.com/FairwindsOps/charts
