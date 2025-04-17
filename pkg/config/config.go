// Copyright 2019 FairwindsOps Inc
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
	_ "embed"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// Configuration contains all of the config for the validation checks.
type Configuration struct {
	DisplayName                  string                 `json:"displayName"`
	Checks                       map[string]Severity    `json:"checks"`
	CustomChecks                 map[string]SchemaCheck `json:"customChecks"`
	Exemptions                   []Exemption            `json:"exemptions"`
	DisallowExemptions           bool                   `json:"disallowExemptions"`
	DisallowConfigExemptions     bool                   `json:"disallowConfigExemptions"`
	DisallowAnnotationExemptions bool                   `json:"disallowAnnotationExemptions"`
	Mutations                    []string               `json:"mutations"`
	KubeContext                  string                 `json:"kubeContext"`
	Namespace                    string                 `json:"namespace"`
}

// Exemption represents an exemption to normal rules
type Exemption struct {
	Rules           []string `json:"rules"`
	ControllerNames []string `json:"controllerNames"`
	ContainerNames  []string `json:"containerNames"`
	Namespace       string   `json:"namespace"`
}

//go:embed default.yaml
var defaultConfig []byte

// MergeConfigAndParseFile parses config from a file.
func MergeConfigAndParseFile(customConfigPath string, mergeConfig bool) (Configuration, error) {
	rawBytes, err := mergeConfigFile(customConfigPath, mergeConfig)
	if err != nil {
		return Configuration{}, err
	}

	return Parse(rawBytes)
}

func mergeConfigFile(customConfigPath string, mergeConfig bool) ([]byte, error) {
	logrus.Infof("Loading config from %s", customConfigPath)
	if customConfigPath == "" {
		return defaultConfig, nil
	}

	var customConfigContent []byte
	var err error
	if strings.HasPrefix(customConfigPath, "https://") || strings.HasPrefix(customConfigPath, "http://") {
		// path is a url
		response, err := http.Get(customConfigPath)
		if err != nil {
			return nil, err
		}
		customConfigContent, err = io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
	} else {
		// path is local
		customConfigContent, err = os.ReadFile(customConfigPath)
		logrus.Infof("Custom config: %v", string(customConfigContent))
		if err != nil {
			logrus.Infof("Error reading config file %s: %v", customConfigPath, err)
			return nil, err
		}
	}

	if mergeConfig {
		mergedConfig, err := mergeYaml(defaultConfig, customConfigContent)
		if err != nil {
			return nil, err
		}
		return mergedConfig, nil
	}

	return customConfigContent, nil
}

// Parse parses config from a byte array.
func Parse(rawBytes []byte) (Configuration, error) {
	reader := bytes.NewReader(rawBytes)
	conf := Configuration{}
	d := yaml.NewYAMLOrJSONDecoder(reader, 4096)
	for {
		if err := d.Decode(&conf); err != nil {
			if err == io.EOF {
				break
			}
			return conf, fmt.Errorf("Decoding config failed: %v", err)
		}
	}
	for key, check := range conf.CustomChecks {
		err := check.Initialize(key)
		if err != nil {
			return conf, err
		}
		conf.CustomChecks[key] = check
		if _, ok := conf.Checks[key]; !ok {
			return conf, fmt.Errorf("no severity specified for custom check %s. Please add the following to your configuration:\n\nchecks:\n  %s: warning # or danger/ignore\n\nto enable your check", key, key)
		}
	}
	return conf, conf.Validate()
}

// Validate checks if a config is valid
func (conf Configuration) Validate() error {
	if len(conf.Checks) == 0 {
		return errors.New("No checks were enabled")
	}
	return nil
}
