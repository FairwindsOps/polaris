# 0.2.0
* Added `--output-format` flag for better CI/CD support
* Added `--display-name` flag
* Added support for StatefulSets
* Show error message if no kubeconfig is set

# 0.1.5
* [Fix](https://github.com/reactiveops/polaris/issues/125): ignore limits/requests for initContainers
* [Fix](https://github.com/reactiveops/polaris/issues/132): support custom base path

# 0.1.4
* [Fix](https://github.com/reactiveops/polaris/issues/116): details pages getting template errors
* [Fix](https://github.com/reactiveops/polaris/issues/114): support all auth providers
* [Fix](https://github.com/reactiveops/polaris/issues/112): Ignore readiness probe for initContainers

# 0.1.3
* [Fix](https://github.com/reactiveops/polaris/issues/109): dashboard not updating when running persistently

# 0.1.2
* Stored all third-party assets (e.g. Charts.js) to local files to support offline dashboard viewing
* Fix: custom configs in `ConfigMap` not respected

# 0.1.1
* [Fix](https://github.com/reactiveops/polaris/issues/93): missing `config.yaml` and dashboard assets in binary releases
* Added some tests and better error handling

# 0.1.0
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
