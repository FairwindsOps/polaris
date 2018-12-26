package config

import (
	corev1 "k8s.io/api/core/v1"
)

// Configuration contains all of the config for the validation checks.
type Configuration struct {
	Resources    RequestsAndLimits
	Healthchecks Probes
}

// RequestsAndLimits contains config for resource requests and limits.
type RequestsAndLimits struct {
	Requests resourceList
	Limits   resourceList
}

type resourceList map[corev1.ResourceName]ResourceMinMax

// ResourceMinMax sets a range for a min and max setting for a resource.
type ResourceMinMax struct {
	Min string
	Max string
}

// Probes contains config for the readiness and liveness probes.
type Probes struct {
	Readiness resourceRequire
	Liveness  resourceRequire
}

type resourceRequire map[require]bool

type require string
