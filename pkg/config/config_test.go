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
	"context"
	"io"
	"log"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"
)

var resourceConfInvalid1 = `test`

var resourceConfYAML1 = `---
resources:
  cpuRequestRanges:
    error:
      below: 100m
      above: 1
    warning:
      below: 200m
      above: 800m
  memoryRequestRanges:
    error:
      below: 100M
      above: 3G
    warning:
      below: 200M
      above: 2G
  cpuLimitRanges:
    error:
      below: 100m
      above: 2
    warning:
      below: 300m
      above: 1800m
  memoryLimitRanges:
    error:
      below: 200M
      above: 6G
    warning:
      below: 300M
      above: 4G
`

var resourceConfJSON1 = `{
	"resources": {
		"cpuRequestRanges": {
			"error": {
				"below": "100m",
				"above": 1
			},
			"warning": {
				"below": "200m",
				"above": "800m"
			}
		},
		"memoryRequestRanges": {
			"error": {
				"below": "100M",
				"above": "3G"
			},
			"warning": {
				"below": "200M",
				"above": "2G"
			}
		},
		"cpuLimitRanges": {
			"error": {
				"below": "100m",
				"above": 2
			},
			"warning": {
				"below": "300m",
				"above": "1800m"
			}
		},
		"memoryLimitRanges": {
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

func TestConfigFromURL(t *testing.T) {
	var err error
	var parsedConf Configuration
	srv := &http.Server{Addr: ":8081"}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, resourceConfYAML1)
	})

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %s", err)
		}
	}()

	parsedConf, err = ParseFile("http://localhost:8081/exampleURL")
	assert.NoError(t, err, "Expected no error when parsing YAML from URL")
	if err := srv.Shutdown(context.TODO()); err != nil {
		panic(err)
	}
	testParsedConfig(t, &parsedConf)

}

func TestConfigNoServerError(t *testing.T) {
	var err error
	_, err = ParseFile("http://localhost:8081/exampleURL")
	assert.EqualError(t, err, "Get http://localhost:8081/exampleURL: dial tcp [::1]:8081: connect: connection refused")
}

func testParsedConfig(t *testing.T, config *Configuration) {
	cpuRequests := config.Resources.CPURequestRanges
	assert.Equal(t, int64(100), cpuRequests.Error.Below.ScaledValue(resource.Milli))
	assert.Equal(t, int64(1000), cpuRequests.Error.Above.ScaledValue(resource.Milli))
	assert.Equal(t, int64(200), cpuRequests.Warning.Below.ScaledValue(resource.Milli))
	assert.Equal(t, int64(800), cpuRequests.Warning.Above.ScaledValue(resource.Milli))

	memRequests := config.Resources.MemoryRequestRanges
	assert.Equal(t, int64(100), memRequests.Error.Below.ScaledValue(resource.Mega))
	assert.Equal(t, int64(3000), memRequests.Error.Above.ScaledValue(resource.Mega))
	assert.Equal(t, int64(200), memRequests.Warning.Below.ScaledValue(resource.Mega))
	assert.Equal(t, int64(2000), memRequests.Warning.Above.ScaledValue(resource.Mega))

	cpuLimits := config.Resources.CPULimitRanges
	assert.Equal(t, int64(100), cpuLimits.Error.Below.ScaledValue(resource.Milli))
	assert.Equal(t, int64(2000), cpuLimits.Error.Above.ScaledValue(resource.Milli))
	assert.Equal(t, int64(300), cpuLimits.Warning.Below.ScaledValue(resource.Milli))
	assert.Equal(t, int64(1800), cpuLimits.Warning.Above.ScaledValue(resource.Milli))

	memLimits := config.Resources.MemoryLimitRanges
	assert.Equal(t, int64(200), memLimits.Error.Below.ScaledValue(resource.Mega))
	assert.Equal(t, int64(6000), memLimits.Error.Above.ScaledValue(resource.Mega))
	assert.Equal(t, int64(300), memLimits.Warning.Below.ScaledValue(resource.Mega))
	assert.Equal(t, int64(4000), memLimits.Warning.Above.ScaledValue(resource.Mega))
}
