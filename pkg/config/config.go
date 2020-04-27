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
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	packr "github.com/gobuffalo/packr/v2"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// Configuration contains all of the config for the validation checks.
type Configuration struct {
	DisplayName        string                 `json:"displayName"`
	Checks             map[string]Severity    `json:"checks"`
	CustomChecks       map[string]SchemaCheck `json:"customChecks"`
	Exemptions         []Exemption            `json:"exemptions"`
	DisallowExemptions bool                   `json:"disallowExemptions"`
}

// Exemption represents an exemption to normal rules
type Exemption struct {
	Rules           []string `json:"rules"`
	ControllerNames []string `json:"controllerNames"`
}

// ParseFile parses config from a file.
func ParseFile(path string) (Configuration, error) {
	var rawBytes []byte
	var err error
	if path == "" {
		configBox := packr.New("Config", "../../examples")
		rawBytes, err = configBox.Find("config.yaml")
	} else if strings.HasPrefix(path, "https://") || strings.HasPrefix(path, "http://") {
		//path is a url
		response, err2 := http.Get(path)
		if err2 != nil {
			return Configuration{}, err2
		}
		rawBytes, err = ioutil.ReadAll(response.Body)
	} else {
		//path is local
		rawBytes, err = ioutil.ReadFile(path)
	}
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
	}
	return conf, nil
}
