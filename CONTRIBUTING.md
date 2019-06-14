# Contributing

Issues, whether bugs, tasks, or feature requests are essential for keeping Polaris great. We believe it should be as easy as possible to contribute changes that get things working in your environment. There are a few guidelines that we need contributors to follow so that we can keep on top of things.

## Code of Conduct

This project adheres to a [code of conduct](CODE_OF_CONDUCT.md). Please review this document before contributing to this project.

## Sign the CLA
Before you can contribute, you will need to sign the [Contributor License Agreement](https://cla-assistant.io/reactiveops/polaris).

## Project Structure

Polaris is built on top of [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime). It can run in 3 different modes, a dashboard, a webhook, or a reporter that prints or exports validation results. All of these modes make use of the shared `validator` and `config` packages. Adding new validations is possible by only making additions to those packages.

## Getting Started

We label issues with the ["good first issue" tag](https://github.com/reactiveops/polaris/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22) if we believe they'll be a good starting point for new contributors. If you're interested in working on an issue, please start a conversation on that issue, and we can help answer any questions as they come up.

## Setting Up Your Development Environment
### Prerequisites
* A properly configured Golang environment with Go 1.11 or higher
* If you want to see the local changes you make on a Polaris dashboard, you will need access to a Kubernetes cluster defined in `~/.kube/config`

### Installation
* Install the project with `go get github.com/reactiveops/polaris`
* Change into the polaris directory which is installed at `$GOPATH/src/github.com/reactiveops/polaris`
* See the dashboard with `go run main.go --dashboard`, then open http://localhost:8080/
* See the audit data  `go run main.go --audit`. This command shows the audit information on the command line. 

## Running Tests

The following commands are all required to pass as part of Polaris testing:

```
go list ./... | grep -v vendor | xargs golint -set_exit_status
go list ./... | grep -v vendor | xargs go vet
go test ./pkg/... -v -coverprofile cover.out
```

## Creating a New Issue

If you've encountered an issue that is not already reported, please create an issue that contains the following:

- Clear description of the issue
- Steps to reproduce it
- Appropriate labels

## Creating a Pull Request

Each new pull request should:

- Reference any related issues
- Add tests that show the issues have been solved
- Pass existing tests and linting
- Contain a clear indication of if they're ready for review or a work in progress
- Be up to date and/or rebased on the master branch

## Creating a new release

### Minor/patch releases
Minor and patch releases only need to change this repo. The Helm chart and deploy scripts
will automatically pull in the latest changes.

To deploy a minor or patch release, follow steps 2 and 3 from "Major releases" below.

### Major releases
Major releases need to change both this repository and the
[Helm chart repo](https://github.com/reactiveops/charts/).

The steps are:
1. Create a PR in the [charts repo](https://github.com/reactiveops/charts/)
    1. Use a branch named `polaris-latest`
    2. Bump the version number in:
        1. stable/polaris/README.md
        2. stable/polaris/Chart.yaml
        3. stable/polaris/values.yaml
    3. **Don't merge yet!**
2. Create a PR for this repo
    1. Bump the version number in:
        1. main.go
        2. README.md
    2. Update CHANGELOG.md
    3. Merge your PR
3. Tag the latest branch for this repo
    1. Pull the latest for the `master` branch
    2. Run `git tag $VERSION && git push --tags`
    3. Wait for CircleCI to finish the build for the tag, which will:
        1. Create the proper image tag in quay.io
        2. Add an entry to the releases page on GitHub
4. Merge the PR for the charts repo you created in step 1.

