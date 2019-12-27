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
	"context"
	"io"
	"log"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var confInvalid = `test`

var confValidYAML = `
checks:
  cpuRequestsMissing: warning
controllersToScan:
  - Deployments
`

var confValidJSON = `
{
  "checks": {
    "cpuRequestsMissing": "warning"
  },
  "controllersToScan": ["Deployments"]
}
`

func TestParseError(t *testing.T) {
	_, err := Parse([]byte(confInvalid))
	expectedErr := "Decoding config failed: error unmarshaling JSON: while decoding JSON: json: cannot unmarshal string into Go value of type config.Configuration"
	assert.EqualError(t, err, expectedErr)
}

func TestParseYaml(t *testing.T) {
	parsedConf, err := Parse([]byte(confValidYAML))
	assert.NoError(t, err, "Expected no error when parsing YAML config")

	testParsedConfig(t, &parsedConf)
}

func TestParseJson(t *testing.T) {
	parsedConf, err := Parse([]byte(confValidJSON))
	assert.NoError(t, err, "Expected no error when parsing JSON config")

	testParsedConfig(t, &parsedConf)
}

func TestConfigFromURL(t *testing.T) {
	var err error
	var parsedConf Configuration
	srv := &http.Server{Addr: ":8081"}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, confValidYAML)
	})

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %s", err)
		}
	}()
	time.Sleep(time.Second)

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
	assert.Error(t, err)
	assert.Regexp(t, regexp.MustCompile("connection refused"), err.Error())
}

func testParsedConfig(t *testing.T, config *Configuration) {
	assert.Equal(t, SeverityWarning, config.Checks["cpuRequestsMissing"])
	assert.Equal(t, Severity(""), config.Checks["cpuLimitsMissing"])
	assert.ElementsMatch(t, []SupportedController{Deployments}, config.ControllersToScan)
}
