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

// ResourceMinMax sets a range for a min and max setting for a resource.
type ResourceMinMax struct {
	Min *resource.Quantity
	Max *resource.Quantity
}

// ResourceList maps the resource name to a range on min and max values.
type ResourceList map[corev1.ResourceName]ResourceMinMax

// RequestsAndLimits contains config for resource requests and limits.
type RequestsAndLimits struct {
	Requests ResourceList
	Limits   ResourceList
}

// Configuration contains all of the config for the validation checks.
type Configuration struct {
	Resources    RequestsAndLimits
	Healthchecks Probes
	Images       Images
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
			return Configuration{}, fmt.Errorf("Decoding config failed: %v", err)
		}
	}
}

// Probes contains config for the readiness and liveness probes.
type Probes struct {
	Readiness resourceRequire
	Liveness  resourceRequire
}

type resourceRequire map[require]bool

type require string

// Images contains the config for images.
type Images struct {
	TagRequired    bool
	WhitelistRepos []string
}
