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

var resourceConfYAML1 = `---
resources:
  cpuRequests:
    error:
      below: 100m
      above: 1
    warning:
      below: 200m
      above: 800m
  memoryRequests:
    error:
      below: 100M
      above: 3G
    warning:
      below: 200M
      above: 2G
  cpuLimits:
    error:
      below: 100m
      above: 2
    warning:
      below: 300m
      above: 1800m
  memoryLimits:
    error:
      below: 200M
      above: 6G
    warning:
      below: 300M
      above: 4G
`

var resourceConfJSON1 = `{
	"resources": {
		"cpuRequests": {
			"error": {
				"below": "100m",
				"above": 1
			},
			"warning": {
				"below": "200m",
				"above": "800m"
			}
		},
		"memoryRequests": {
			"error": {
				"below": "100M",
				"above": "3G"
			},
			"warning": {
				"below": "200M",
				"above": "2G"
			}
		},
		"cpuLimits": {
			"error": {
				"below": "100m",
				"above": 2
			},
			"warning": {
				"below": "300m",
				"above": "1800m"
			}
		},
		"memoryLimits": {
			"error": {
				"below": "200M",
				"above": "6G"
			},
			"warning": {
				"below": "300M",
				"above": "4G"
			}
		}
	}
}`

func TestParseError(t *testing.T) {
	_, err := Parse([]byte(resourceConfInvalid1))
	expectedErr := "Decoding config failed: error unmarshaling JSON: while decoding JSON: json: cannot unmarshal string into Go value of type config.Configuration"
	assert.EqualError(t, err, expectedErr)
}

func TestParseYaml(t *testing.T) {
	parsedConf, err := Parse([]byte(resourceConfYAML1))
	assert.NoError(t, err, "Expected no error when parsing YAML config")

	testParsedConfig(t, &parsedConf)
}

func TestParseJson(t *testing.T) {
	parsedConf, err := Parse([]byte(resourceConfJSON1))
	assert.NoError(t, err, "Expected no error when parsing JSON config")

	testParsedConfig(t, &parsedConf)
}

func testParsedConfig(t *testing.T, config *Configuration) {
	cpuRequests := config.Resources.CPURequests
	assert.Equal(t, int64(100), cpuRequests.Error.Below.ScaledValue(resource.Milli))
	assert.Equal(t, int64(1000), cpuRequests.Error.Above.ScaledValue(resource.Milli))
	assert.Equal(t, int64(200), cpuRequests.Warning.Below.ScaledValue(resource.Milli))
	assert.Equal(t, int64(800), cpuRequests.Warning.Above.ScaledValue(resource.Milli))

	memRequests := config.Resources.MemoryRequests
	assert.Equal(t, int64(100), memRequests.Error.Below.ScaledValue(resource.Mega))
	assert.Equal(t, int64(3000), memRequests.Error.Above.ScaledValue(resource.Mega))
	assert.Equal(t, int64(200), memRequests.Warning.Below.ScaledValue(resource.Mega))
	assert.Equal(t, int64(2000), memRequests.Warning.Above.ScaledValue(resource.Mega))

	cpuLimits := config.Resources.CPULimits
	assert.Equal(t, int64(100), cpuLimits.Error.Below.ScaledValue(resource.Milli))
	assert.Equal(t, int64(2000), cpuLimits.Error.Above.ScaledValue(resource.Milli))
	assert.Equal(t, int64(300), cpuLimits.Warning.Below.ScaledValue(resource.Milli))
	assert.Equal(t, int64(1800), cpuLimits.Warning.Above.ScaledValue(resource.Milli))

	memLimits := config.Resources.MemoryLimits
	assert.Equal(t, int64(200), memLimits.Error.Below.ScaledValue(resource.Mega))
	assert.Equal(t, int64(6000), memLimits.Error.Above.ScaledValue(resource.Mega))
	assert.Equal(t, int64(300), memLimits.Warning.Below.ScaledValue(resource.Mega))
	assert.Equal(t, int64(4000), memLimits.Warning.Above.ScaledValue(resource.Mega))
}
