# Contributing

Issues, whether bugs, tasks, or feature requests are essential for keeping Polaris great. We believe it should be as easy as possible to contribute changes that get things working in your environment. There are a few guidelines that we need contributors to follow so that we can have a chance of keeping on top of things.

## Code of Conduct

This project adheres to a [code of conduct](CODE_OF_CONDUCT.md). Please review this document before contributing to this project.

## Project Structure

Polaris is built on top of [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime). It can run in 3 different modes, a dashboard, a webhook, or a reporter that prints or exports validation results. All of these modes make use of the shared `validator` and `config` packages. Adding new validations is possible by only making additions to those packages.

## Getting Started

We label issues with the ["good first issue" tag](https://github.com/reactiveops/polaris/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22) if we believe they'll be a good starting point for new contributors. If you're interested in working on an issue, please start a conversation on that issue, and we can help answer any questions as they come up.

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

