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
	"testing"

	"github.com/stretchr/testify/assert"
)

var confExemptRuleTest = `
checks:
  ANY: warning
  OTHER: warning
exemptions:
  - controllerNames:
    - test
    rules:
    - ANY
`

var confExemptTest = `
checks:
  ANY: warning
exemptions:
  - controllerNames:
      - test
`

func TestInclusiveExemption(t *testing.T) {
	parsedConf, _ := Parse([]byte(confExemptTest))
	applicable := parsedConf.IsActionable("ANY", "test")
	applicableOtherController := parsedConf.IsActionable("ANY", "other")

	assert.False(t, applicable, "Expected all checks to be exempted when their controller is specified.")
	assert.True(t, applicableOtherController, "Expected checks to only be exempted when their controller is specified.")
}

func TestIndividualRuleException(t *testing.T) {
	parsedConf, _ := Parse([]byte(confExemptRuleTest))
	applicable := parsedConf.IsActionable("ANY", "test")
	applicableOtherRule := parsedConf.IsActionable("OTHER", "test")
	applicableOtherRuleOtherController := parsedConf.IsActionable("OTHER", "other")
	applicableRuleOtherController := parsedConf.IsActionable("ANY", "other")

	assert.False(t, applicable, "Expected all checks to be exempted when their controller and rule are specified.")
	assert.True(t, applicableOtherRule, "Expected checks to only be exempted when their controller and rule are specified.")
	assert.True(t, applicableOtherRuleOtherController, "Expected checks to only be exempted when their controller and rule are specified.")
	assert.True(t, applicableRuleOtherController, "Expected checks to only be exempted when their controller and rule are specified.")
}
