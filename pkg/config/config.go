package config

import (
	corev1 "k8s.io/api/core/v1"
)

type resourceQuanityRange struct {
	Min string
	Max string
}

type ResourceListRange map[corev1.ResourceName]resourceQuanityRange

type Configuration struct {
	Resources ResourceListRange
}
