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

type ResourceMinMax struct {
	Min *resource.Quantity
	Max *resource.Quantity
}

type ResourceList map[corev1.ResourceName]ResourceMinMax

type RequestsAndLimits struct {
	Requests ResourceList
	Limits   ResourceList
}

type Configuration struct {
	Resources RequestsAndLimits
}

// ParseFile parses config from a file
func ParseFile(path string) (Configuration, error) {
	rawBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return Configuration{}, err
	}
	return Parse(rawBytes)
}

// Parse parses config from a byte array
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
