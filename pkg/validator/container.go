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

	conf "github.com/reactiveops/fairwinds/pkg/config"
	"github.com/reactiveops/fairwinds/pkg/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func validateContainer(conf conf.Configuration, container corev1.Container) types.ContainerResults {
	results := types.ContainerResults{
		Name: container.Name,
	}
	results = resources(conf.Resources, container, results)
	results = probes(conf.Resources, container, results)
	results = tag(conf.Resources, container, results)

	return results
}

func resources(conf conf.ResourceRequestsAndLimits, c corev1.Container, results types.ContainerResults) types.ContainerResults {
	confCPUmin, err := resource.ParseQuantity(conf.Requests["cpu"].Min)
	if err != nil {
		log.Error(err, "cpu min parse quan")
	}
	// CPUmax, err := resource.ParseQuantity(conf["cpu"].Max)
	// if err != nil {
	// 	log.Error(err, "cpu max parse quan")
	// }

	containerRequests := c.Resources.Requests.Cpu()
	if containerRequests.MilliValue() < confCPUmin.MilliValue() {
		results.AddFailure("CPU requests", confCPUmin.String(), containerRequests.String())
	}

	if c.Resources.Requests.Memory().IsZero() {
		results.AddFailure("Memory requests", "placeholder", "placeholder")
	}
	if c.Resources.Limits.Cpu().IsZero() {
		results.AddFailure("CPU limits", "placeholder", "placeholder")
	}
	if c.Resources.Limits.Memory().IsZero() {
		results.AddFailure("Memory limits", "placeholder", "placeholder")
	}
	return results
}

func probes(conf conf.ResourceRequestsAndLimits, c corev1.Container, results types.ContainerResults) types.ContainerResults {
	if c.ReadinessProbe == nil {
		results.AddFailure("Readiness Probe", "placeholder", "placeholder")
	}

	if c.LivenessProbe == nil {
		results.AddFailure("Liveness Probe", "placeholder", "placeholder")
	}
	return results
}

func tag(conf conf.ResourceRequestsAndLimits, c corev1.Container, results types.ContainerResults) types.ContainerResults {
	img := strings.Split(c.Image, ":")
	if len(img) == 1 || img[1] == "latest" {
		results.AddFailure("Image Tag", "not latest", "latest")
	}
	return results
}

func hostPort(conf conf.ResourceRequestsAndLimits, c corev1.Container, results types.ContainerResults) types.ContainerResults {
	for _, port := range c.Ports {
		if port.HostPort != 0 {
			results.AddFailure("Host port", "placeholder", "placeholder")
		}
	}
	return results
}
