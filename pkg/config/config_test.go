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
	"fmt"
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
`

var confValidJSON = `
{
  "checks": {
    "cpuRequestsMissing": "warning"
  }
}
`

var confCustomChecks = `
checks:
  foo: warning
customChecks:
  foo:
    successMessage: Security context is set
    failureMessage: Security context should be set
    category: Security
    target: Container
    schema:
      '$schema': http://json-schema.org/draft-07/schema
      type: object
      required:
      - securityContext
`

var confCustomChecksWithJSONSchema = `
checks:
  foo: warning
customChecks:
  foo:
    successMessage: Security context is set
    failureMessage: Security context should be set
    category: Security
    target: Container
    jsonSchema: >
      {
        "$schema": "http://json-schema.org/draft-07/schema",
        "type": "object",
        "required": ["securityContext"]
      }
`

var confCustomChecksMissing = `
customChecks:
  foo:
    successMessage: Security context is set
    failureMessage: Security context should be set
    category: Security
    target: Container
    schema:
      '$schema': http://json-schema.org/draft-07/schema
      type: object
      required:
      - securityContext

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

func TestConfigWithCustomChecks(t *testing.T) {
	valid := map[string]interface{}{
		"securityContext": map[string]interface{}{
			"foo": "bar",
		},
	}
	invalid := map[string]interface{}{
		"notSecurityContext": map[string]interface{}{},
	}

	parsedConf, err := Parse([]byte(confCustomChecks))
	assert.NoError(t, err, "Expected no error when parsing YAML config")
	assert.Equal(t, 1, len(parsedConf.CustomChecks))
	check, err := parsedConf.CustomChecks["foo"].TemplateForResource(map[string]interface{}{})
	isValid, _, err := check.CheckObject(valid)
	assert.NoError(t, err)
	assert.Equal(t, true, isValid)
	isValid, _, err = check.CheckObject(invalid)
	assert.NoError(t, err)
	assert.Equal(t, false, isValid)

	parsedConf, err = Parse([]byte(confCustomChecksWithJSONSchema))
	assert.NoError(t, err, "Expected no error when parsing YAML config")
	assert.Equal(t, 1, len(parsedConf.CustomChecks))
	isValid, problems, err := parsedConf.CustomChecks["foo"].CheckObject(valid)
	assert.NoError(t, err)
	if !assert.Equal(t, true, isValid) {
		fmt.Println(problems[0].PropertyPath, problems[0].InvalidValue, problems[0].Message)
	}
	isValid, _, err = check.CheckObject(invalid)
	assert.NoError(t, err)
	assert.Equal(t, false, isValid)
}

func TestCustomChecksMissingSeverity(t *testing.T) {
	_, err := Parse([]byte(confCustomChecksMissing))
	assert.Error(t, err, "Expected error when check has no severity set")
}

func testParsedConfig(t *testing.T, config *Configuration) {
	assert.Equal(t, SeverityWarning, config.Checks["cpuRequestsMissing"])
	assert.Equal(t, Severity(""), config.Checks["cpuLimitsMissing"])
}
