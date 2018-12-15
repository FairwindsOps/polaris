// Copyright 2018 ReactiveOps
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package validator

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
)

type containerResults struct {
	Name   string
	Reason string
}

func validateContainer(container corev1.Container) containerResults {
	var sb strings.Builder
	results := containerResults{
		Name: container.Name,
	}

	resources(container, sb)
	probes(container, sb)
	tag(container, sb)

	results.Reason = sb.String()
	return results
}

func resources(c corev1.Container, sb strings.Builder) string {
	log.Info("validating Container:", "container resources", c.Resources)
	if c.Resources.Requests.Cpu().IsZero() {
		sb.WriteString("- CPU requests are not set.\n")
	}
	if c.Resources.Requests.Memory().IsZero() {
		sb.WriteString("- Memory requests are not set.\n")
	}
	if c.Resources.Limits.Cpu().IsZero() {
		sb.WriteString("- CPU limits are not set.\n")
	}
	if c.Resources.Limits.Memory().IsZero() {
		sb.WriteString("- Memory limits are not set.\n")
	}

	return sb.String()
}

func probes(c corev1.Container, sb strings.Builder) string {
	if c.ReadinessProbe == nil {
		sb.WriteString("- Readiness Probe is not set.\n")
	}

	if c.LivenessProbe == nil {
		sb.WriteString("- Liveness Probe is not set.\n")
	}
	return sb.String()
}

func tag(c corev1.Container, sb strings.Builder) string {
	img := strings.Split(c.Image, ":")
	if len(img) == 1 || img[1] == "latest" {
		sb.WriteString("- Image tag is latest.\n")
	}

	return sb.String()
}

func hostPort(c corev1.Container, sb strings.Builder) string {
	for _, port := range c.Ports {
		if port.HostPort != 0 {
			sb.WriteString("- Host Port set.\n")
		}
	}
	return sb.String()
}
