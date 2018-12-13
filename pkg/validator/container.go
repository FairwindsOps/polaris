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
	c := containerResults{
		Name: container.Name,
	}

	log.Info("validateing Container:", "container resources", container.Resources)
	if container.Resources.Requests.Cpu().IsZero() {
		sb.WriteString("- CPU requests are not set.\n")
	}
	if container.Resources.Requests.Memory().IsZero() {
		sb.WriteString("- Memory requests are not set.\n")
	}
	if container.Resources.Limits.Cpu().IsZero() {
		sb.WriteString("- CPU limits are not set.\n")
	}
	if container.Resources.Limits.Memory().IsZero() {
		sb.WriteString("- Memory limits are not set.\n")
	}
	c.Reason = sb.String()

	return c
}
