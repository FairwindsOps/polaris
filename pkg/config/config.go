package config

import (
	corev1 "k8s.io/api/core/v1"
)

// Configuration contains the config for the validation checks.
type Configuration struct {
	Resources    ResourceRequestsAndLimits
	Healthchecks ResourceHealthChecks
}

type resourceMinMax struct {
	Min string
	Max string
}

type resourceList map[corev1.ResourceName]resourceMinMax

// ResourceRequestsAndLimits contains config for resource requests and limits.
type ResourceRequestsAndLimits struct {
	Requests resourceList
	Limits   resourceList
}

type require string

type resourceRequire map[require]bool

// ResourceHealthChecks contains config for the health check settings for probes.
type ResourceHealthChecks struct {
	Readiness resourceRequire
	Liveness  resourceRequire
}
