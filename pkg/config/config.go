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
	Resources    RequestsAndLimits `json:"resources"`
	HealthChecks Probes            `json:"healthChecks"`
	Images       Images            `json:"images"`
	HostNetwork  HostNetwork       `json:"hostNetwork"`
}

// RequestsAndLimits contains config for resource requests and limits.
type RequestsAndLimits struct {
	Requests ResourceList `json:"requests"`
	Limits   ResourceList `json:"limits"`
}

// ResourceList maps the resource name to a range on min and max values.
type ResourceList map[corev1.ResourceName]ResourceMinMax

// ResourceMinMax sets a range for a min and max setting for a resource.
type ResourceMinMax struct {
	Min *resource.Quantity `json:"min"`
	Max *resource.Quantity `json:"max"`
}

// Probes contains config for the readiness and liveness probes.
type Probes struct {
	Readiness ResourceRequire `json:"readiness"`
	Liveness  ResourceRequire `json:"liveness"`
}

// ResourceRequire indicates if this resource should be validated.
type ResourceRequire struct {
	Require bool `json:"require"`
}

// Images contains the config for images.
type Images struct {
	TagRequired    bool     `json:"tagRequired"`
	WhitelistRepos []string `json:"whitelistRepos"`
}

// HostNetwork contains the config for host networking validations.
type HostNetwork struct {
	HostPort bool `json:"hostPort"`
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
