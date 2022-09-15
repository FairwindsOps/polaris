// Copyright 2022 FairwindsOps, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fairwindsops/polaris/pkg/config"
	"github.com/fairwindsops/polaris/pkg/kube"
	"github.com/fairwindsops/polaris/pkg/validator"
)

var testCases = []testCase{}

type testCase struct {
	check     string
	filename  string
	resources *kube.ResourceProvider
	failure   bool
	config    config.Configuration
	manifest  string
}

var mutatedYamlContentMap = map[string]string{}
var mutationTestCasesMap = map[string][]testCase{}

func init() {
	checkToTest := os.Getenv("POLARIS_CHECK_TEST")
	_, baseDir, _, _ := runtime.Caller(0)
	baseDir = filepath.Dir(baseDir) + "/checks"
	dirs, err := ioutil.ReadDir(baseDir)
	if err != nil {
		panic(err)
	}
	for _, dir := range dirs {
		check := dir.Name()
		if checkToTest != "" && checkToTest != check {
			continue
		}
		checkDir := baseDir + "/" + check
		cases, err := ioutil.ReadDir(checkDir)
		if err != nil {
			panic(err)
		}
		configString := "checks:\n  " + check + ": danger"
		checkPath := checkDir + "/check.yaml"
		customCheckContent, err := ioutil.ReadFile(checkPath)
		if err == nil {
			lines := strings.Split(string(customCheckContent), "\n")
			for idx := range lines {
				lines[idx] = "    " + lines[idx]
			}
			configString += "\ncustomChecks:\n  " + check + ":\n"
			configString += strings.Join(lines, "\n")
		}
		c, err := config.Parse([]byte(configString))
		if err != nil {
			panic(err)
		}
		for _, tc := range cases {
			if tc.Name() == "check.yaml" {
				continue
			}
			yamlContent, err := os.ReadFile(checkDir + "/" + tc.Name())
			if err != nil {
				panic(err)
			}
			resourceFilename := strings.Replace(tc.Name(), "mutated", "failure", -1)

			resources, err := kube.CreateResourceProviderFromPath(checkDir + "/" + resourceFilename)
			if err != nil {
				panic(err)
			}
			testcase := testCase{
				filename:  tc.Name(),
				check:     check,
				resources: resources,
				failure:   strings.Contains(resourceFilename, "failure"),
				config:    c,
				manifest:  string(yamlContent),
			}

			if strings.Contains(tc.Name(), "mutated") {
				key := fmt.Sprintf("%s/%s", check, tc.Name())
				mutatedYamlContentMap[key] = string(yamlContent)
				testCases, ok := mutationTestCasesMap[check]
				if !ok {
					testCases = []testCase{}
				}
				testCases = append(testCases, testcase)
				mutationTestCasesMap[check] = testCases
			} else {
				testCases = append(testCases, testcase)
			}
		}
	}
}

func TestChecks(t *testing.T) {
	for _, tc := range testCases {
		results, err := validator.ApplyAllSchemaChecksToResourceProvider(&tc.config, tc.resources)
		if err != nil {
			panic(err)
		}
		auditData := validator.AuditData{Results: results}
		summary := auditData.GetSummary()
		total := summary.Successes + summary.Dangers
		msg := fmt.Sprintf("Check %s ran %d times instead of 1", tc.check, total)
		if assert.LessOrEqual(t, uint(1), total, msg) {
			if tc.failure {
				message := "Check " + tc.check + " passed unexpectedly for " + tc.filename
				assert.LessOrEqual(t, uint(1), summary.Dangers, message)
			} else {
				message := "Check " + tc.check + " failed unexpectedly for " + tc.filename
				assert.LessOrEqual(t, uint(1), summary.Successes, message)
				assert.Equal(t, uint(0), summary.Dangers, message)
			}
		}
	}
}
