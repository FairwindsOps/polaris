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

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"
)

var resourceConfInvalid1 = `test`

var resourceConfYaml1 = `---
resources:
  requests:
    cpu:
      min: 100m
      max: 1
    memory:
      min: 100M
      max: 3G
  limits:
    cpu:
      min: 150m
      max: 2
    memory:
      min: 150M
      max: 4G
`

var resourceConfJson1 = `{
	"resources": {
		"requests": {
			"cpu": {
				"min": "100m",
				"max": "1"
			},
			"memory": {
				"min": "100M",
				"max": "3G"
			}
		},
		"limits": {
			"cpu": {
				"min": "150m",
				"max": "2"
			},
			"memory": {
				"min": "150M",
				"max": "4G"
			}
		}
	}
}`

func TestParseError(t *testing.T) {
	_, err := Parse([]byte(resourceConfInvalid1))
	assert.EqualError(t, err, "Decoding config failed: error unmarshaling JSON: json: cannot unmarshal string into Go value of type config.Configuration")
}

func TestParseYaml(t *testing.T) {
	parsedConf, err := Parse([]byte(resourceConfYaml1))
	assert.NoError(t, err, "Expected no error when parsing config")

	requests := parsedConf.Resources.Requests
	assert.Equal(t, int64(100), requests["cpu"].Min.ScaledValue(resource.Milli))
	assert.Equal(t, int64(1000), requests["cpu"].Max.ScaledValue(resource.Milli))
	assert.Equal(t, int64(100), requests["memory"].Min.ScaledValue(resource.Mega))
	assert.Equal(t, int64(3000), requests["memory"].Max.ScaledValue(resource.Mega))

	limits := parsedConf.Resources.Limits
	assert.Equal(t, int64(150), limits["cpu"].Min.ScaledValue(resource.Milli))
	assert.Equal(t, int64(2000), limits["cpu"].Max.ScaledValue(resource.Milli))
	assert.Equal(t, int64(150), limits["memory"].Min.ScaledValue(resource.Mega))
	assert.Equal(t, int64(4000), limits["memory"].Max.ScaledValue(resource.Mega))
}

func TestParseJson(t *testing.T) {
	parsedConf, err := Parse([]byte(resourceConfJson1))
	assert.NoError(t, err, "Expected no error when parsing config")

	requests := parsedConf.Resources.Requests
	assert.Equal(t, int64(100), requests["cpu"].Min.ScaledValue(resource.Milli))
	assert.Equal(t, int64(1000), requests["cpu"].Max.ScaledValue(resource.Milli))
	assert.Equal(t, int64(100), requests["memory"].Min.ScaledValue(resource.Mega))
	assert.Equal(t, int64(3000), requests["memory"].Max.ScaledValue(resource.Mega))

	limits := parsedConf.Resources.Limits
	assert.Equal(t, int64(150), limits["cpu"].Min.ScaledValue(resource.Milli))
	assert.Equal(t, int64(2000), limits["cpu"].Max.ScaledValue(resource.Milli))
	assert.Equal(t, int64(150), limits["memory"].Min.ScaledValue(resource.Mega))
	assert.Equal(t, int64(4000), limits["memory"].Max.ScaledValue(resource.Mega))
}
