---
sidebarDepth: 0
meta:
  - name: description
    content: "Fairwinds Polaris | Changelog"

---
## 4.1.0
* Handle case-insentitivity in capabilities checks
* Change test for PDB disruptions to better handle IaC

## 4.0.9
* Update Alpine base image

## 4.0.8
* Fix support for namespace checks

## 4.0.7
* Fix goreleaser format

## 4.0.6
* Change goreleaser format
* Fix `--helm-values` flag

## 4.0.5
* Bugfix for repeated objects on the dashboard

## 4.0.4
* Bugfix for validating webhook and non-pod checks

## 4.0.3
* Fixed bad interaction between `--set-exit-score-below` and `--only-show-failed-tests`
* Dependency updates
* Support for Helm chart scanning

## 4.0.2
* Goreleaser fix

## 4.0.1
* Goreleaser fix

## 4.0.0
* Add support for arbitrary resources, like Ingress or PodDisruptionBudget
* Add support check templating (see docs)
* Add support for multi-resource checks (see docs)

### Breaking Changes
* In custom checks, `jsonSchema` is now `schemaString`
* Check `pdbDisruptionsAllowedGreaterThanZero` is now called `pdbDisruptionsIsZero`

## 3.2.0
* Add `--format=pretty` option for CLI output

## 3.1.6
* Fix nil pointer issue with --only-output-failed-tests

## 3.1.5
* Fix UI display of Ingress checks

## 3.1.4
* Fixes for exemption annotations for the admission controller

## 3.1.3
* Fixes for `privilegeEscalationAllowed` and `insecureCapabilities` checks to take Kubernetes defaults into account

## 3.1.2
* Start checking deployment configuration using Fairwinds Insights

## 3.1.1
* Updated to alpine:3.13

## 3.1.0
* Added support for Ingress objects
* Fixes for exemptions, including support for exempting entire namespaces

## 3.0.0
* **Breaking** - fixed inconsistency in how controller-level checks are handled
Custom checks with `target: Controller` should remove `Object` from the top-level of the
JSON schema (see changes to `./checks/multipleReplicasForDeployment.yaml`)

## 2.0.1
* Fixed Polaris deployment process

## 2.0.0
* Standardize categories of checks into Security, Reliability, and Efficiency
* Changes to the dashboard UI
* Update controller-runtime

## 1.2.1
* Update date on dashboard footer

## 1.2.0
* Add ability to audit a single workload
* Enable `pullPolicyAlways` by default
* Fix for finding parent resources

## 1.1.1
* Show controller checks on dashboard
* Fix for orphaned pods w/ controller checks

## 1.1.0
* Add namespace filter in UI
* Add priorityClass check
* Support reading from STDIN
* Ensure severity is set for all custom checks
* Support audit files which use \r or \r\n as newline character
* Add option to exempt an entire controller from checks via config file
* Fixed case where parent resources trigger error
* Fixed UI zero-state

## 1.0.3
* Fixed case where parent resources trigger error
* Fixed dashboard link when `--base-path` is set

## 1.0.2
* Fixed case where custom CRDs are not covered by RBAC

## 1.0.1
* Added ARM binaries to releases

## 1.0.0
### New Features
* Added support for custom checks using JSON Schema
* Added support for arbitrary controllers, rather than a pre-configured set
    * removed support for `controllers_to_scan` in config
* Added the ability to exempt a particular controller from a particular check.
* Docker image now includes the default config

### Breaking Changes
* Breaking changes in both input and output formats. See [Examples](https://github.com/FairwindsOps/polaris/tree/master/examples) for examples of the new formats.
    * removed config-level configuration for checks like max/min memory settings
    * changed severity `error` to `danger`
* Breaking changes to the CLI
    * CLI flag `--set-exit-code-on-error` is now `--set-exit-code-on-danger`
    * Flags `--version`, `--dashboard`, `--webhook`, and `--audit` are now arguments
    * Port flags are now just `--port`

## 0.6.0
* Fixed webhook support in Kubernetes 1.16
  * this also removes support for 1.8
* Added support for exemptions via controller annotations

## 0.5.2
* Fixed missing success messages for resource requests/limits

## 0.5.1
* Added a few more exemptions
* Started checking exemptions based on controller name prefix
* `runAsUser != 0` now passes the `runAsNonRoot` check

## 0.5.0
* Added `--load-audit-file` flag to run the dashboard from an existing audit
* Added an `ID` field to each check in the output
* Skip health checks for jobs, cronjobs, initcontainers
* Added support for exemptions
* Fixed dashboard base path option

## 0.4.0
* Added additional Pod Controllers to scan PodSpec (`jobs`, `cronjobs`, `daemonsets`, `replicationcontrollers`)

## 0.3.1
* Changed dashboard branding to refer to new org name Fairwinds

## 0.3.0
* Added `--set-exit-code-on-error` and `--set-exit-code-below-score` flags to better support CI/CD

## 0.2.1
* [Fix](https://github.com/FairwindsOps/polaris/issues/146): Fixed logic on RunAsNonRoot check to incorporate settings in podSpec

## 0.2.0
* Added `--output-format` flag for better CI/CD support
* Added `--display-name` flag
* Added support for StatefulSets
* Show error message if no kubeconfig is set

## 0.1.5
* [Fix](https://github.com/FairwindsOps/polaris/issues/125): ignore limits/requests for initContainers
* [Fix](https://github.com/FairwindsOps/polaris/issues/132): support custom base path

## 0.1.4
* [Fix](https://github.com/FairwindsOps/polaris/issues/116): details pages getting template errors
* [Fix](https://github.com/FairwindsOps/polaris/issues/114): support all auth providers
* [Fix](https://github.com/FairwindsOps/polaris/issues/112): Ignore readiness probe for initContainers

## 0.1.3
* [Fix](https://github.com/FairwindsOps/polaris/issues/109): dashboard not updating when running persistently

## 0.1.2
* Stored all third-party assets (e.g. Charts.js) to local files to support offline dashboard viewing
* Fix: custom configs in `ConfigMap` not respected

## 0.1.1
* [Fix](https://github.com/FairwindsOps/polaris/issues/93): missing `config.yaml` and dashboard assets in binary releases
* Added some tests and better error handling

## 0.1.0
* Dashboard fully functional
* Validating webhook functional, but still considered beta
* Checks:
  * Health
    * readiness probe missing
    * liveness probe missing
  * Images
    * tag not specified
    * pull policy not always
  * Networking
    * host network set
    * host port set
  * Resources
    * cpu/memory requests missing
    * cpu/memory limits missing
    * cpu/memory ranges exceeded
  * Security
    * security capabilities
    * host IPC set
    * host PID set
    * not read-only fs
    * privilege escalation allowed
    * run as root allowed
    * run as privileged
