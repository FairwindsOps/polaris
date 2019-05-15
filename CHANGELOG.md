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
