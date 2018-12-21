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
	"fmt"
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

func ensureWithinRange(r types.ContainerResults, resName string, expectedRange conf.ResourceMinMax, actual *resource.Quantity) {
	expectedMin, err := resource.ParseQuantity(expectedRange.Min)
	if err != nil {
		log.Error(err, fmt.Sprintf("Error parsing min quantity for %s", resName))
	} else if expectedMin.MilliValue() > actual.MilliValue() {
		r.AddFailure(resName, expectedMin.String(), actual.String())
	}

	expectedMax, err := resource.ParseQuantity(expectedRange.Max)
	if err != nil {
		log.Error(err, "Error parsing max quantity")
	} else if expectedMax.MilliValue() < actual.MilliValue() {
		r.AddFailure(resName, expectedMax.String(), actual.String())
	}
}

func resources(conf conf.ResourceRequestsAndLimits, c corev1.Container, results types.ContainerResults) types.ContainerResults {
	ensureWithinRange(results, "requests.cpu", conf.Requests["cpu"], c.Resources.Requests.Cpu())
	ensureWithinRange(results, "requests.memory", conf.Requests["memory"], c.Resources.Requests.Memory())
	ensureWithinRange(results, "limits.cpu", conf.Limits["cpu"], c.Resources.Limits.Cpu())
	ensureWithinRange(results, "limits.memory", conf.Limits["memory"], c.Resources.Limits.Memory())

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
