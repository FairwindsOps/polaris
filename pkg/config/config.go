// Copyright 2019 ReactiveOps
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

package config

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// Configuration contains all of the config for the validation checks.
type Configuration struct {
	Resources    Resources    `json:"resources"`
	HealthChecks HealthChecks `json:"healthChecks"`
	Images       Images       `json:"images"`
	Networking   Networking   `json:"networking"`
	Security     Security     `json:"security"`
}

// Resources contains config for resource requests and limits.
type Resources struct {
	CPURequestsMissing    Severity       `json:"cpuRequestsMissing"`
	CPURequestRanges      ResourceRanges `json:"cpuRequestRanges"`
	CPULimitsMissing      Severity       `json:"cpuLimitsMissing"`
	CPULimitRanges        ResourceRanges `json:"cpuLimitRanges"`
	MemoryRequestsMissing Severity       `json:"memoryRequestsMissing"`
	MemoryRequestRanges   ResourceRanges `json:"memoryRequestRanges"`
	MemoryLimitsMissing   Severity       `json:"memoryLimitsMissing"`
	MemoryLimitRanges     ResourceRanges `json:"memoryLimitRanges"`
}

// ResourceRanges contains config for requests or limits for a specific resource.
type ResourceRanges struct {
	Warning ResourceRange `json:"warning"`
	Error   ResourceRange `json:"error"`
}

// ResourceRange can contain below and above conditions for validation.
type ResourceRange struct {
	Below *resource.Quantity `json:"below"`
	Above *resource.Quantity `json:"above"`
}

// HealthChecks contains config for readiness and liveness probes.
type HealthChecks struct {
	ReadinessProbeMissing Severity `json:"readinessProbeMissing"`
	LivenessProbeMissing  Severity `json:"livenessProbeMissing"`
}

// Images contains the config for images.
type Images struct {
	TagNotSpecified     Severity          `json:"tagNotSpecified"`
	PullPolicyNotAlways Severity          `json:"pullPolicyNotAlways"`
	Whitelist           ErrorWarningLists `json:"whitelist"`
	Blacklist           ErrorWarningLists `json:"blacklist"`
}

// ErrorWarningLists provides lists of patterns to match or avoid in image tags.
type ErrorWarningLists struct {
	Error   []string `json:"error"`
	Warning []string `json:"warning"`
}

// Networking contains the config for networking validations.
type Networking struct {
	HostAliasSet   Severity `json:"hostAliasSet"`
	HostIPCSet     Severity `json:"hostIPCSet"`
	HostNetworkSet Severity `json:"hostNetworkSet"`
	HostPIDSet     Severity `json:"hostPIDSet"`
	HostPortSet    Severity `json:"hostPortSet"`
}

// Security contains the config for security validations.
type Security struct {
	RunAsRootAllowed           Severity             `json:"runAsRootAllowed"`
	RunAsPrivileged            Severity             `json:"RunAsPrivileged"`
	NotReadOnlyRootFileSystem  Severity             `json:"notReadOnlyRootFileSystem"`
	PrivilegeEscalationAllowed Severity             `json:"privilegeEscalationAllowed"`
	Capabilities               SecurityCapabilities `json:"capabilities"`
}

// SecurityCapabilities contains the config for security capabilities validations.
type SecurityCapabilities struct {
	Error   SecurityCapabilityLists `json:"error"`
	Warning SecurityCapabilityLists `json:"warning"`
}

// SecurityCapabilityLists contains the config for security capabilitie list validations.
type SecurityCapabilityLists struct {
	IfAnyAdded       []corev1.Capability `json:"ifAnyAdded"`
	IfAnyAddedBeyond []corev1.Capability `json:"ifAnyAddedBeyond"`
	IfAnyNotDropped  []corev1.Capability `json:"ifAnyNotDropped"`
}

// ParseFile parses config from a file.
func ParseFile(path string) (Configuration, error) {
	rawBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return Configuration{}, err
	}
	return Parse(rawBytes)
}

// Parse parses config from a byte array.
func Parse(rawBytes []byte) (Configuration, error) {
	reader := bytes.NewReader(rawBytes)
	conf := Configuration{}
	d := yaml.NewYAMLOrJSONDecoder(reader, 4096)
	for {
		if err := d.Decode(&conf); err != nil {
			if err == io.EOF {
				return conf, nil
			}
			return conf, fmt.Errorf("Decoding config failed: %v", err)
		}
	}
}
