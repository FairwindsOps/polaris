package config

import (
	corev1 "k8s.io/api/core/v1"
)

type ResourceMinMax struct {
	Min string
	Max string
}

type ResourceList map[corev1.ResourceName]ResourceMinMax

type RequestsAndLimits struct {
	Requests ResourceList
	Limits   ResourceList
}

type Configuration struct {
	Resources RequestsAndLimits
}
