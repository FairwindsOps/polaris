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
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type checkMarshal struct {
	Controllers []SupportedController `json:"controllers"`
}

func TestUnmarshalSupportedControllers(t *testing.T) {
	for idx, controllerString := range ControllerStrings {
		// Check taking all strings and convert them into enums
		object := checkMarshal{}
		jsonBytes := []byte(fmt.Sprintf(`{"controllers":["%v"]}`, controllerString))
		err := json.Unmarshal(jsonBytes, &object)
		if idx == 0 {
			if err == nil {
				// Assure the first element always should NOT unmarshal
				t.Errorf("Expected the first element (%s) to fail json unmarshal. First element in this array should always be 'Unsupported'", controllerString)
			}
		} else if err != nil {
			t.Errorf("Could not unmarshal json (%s) to a Supported controller; Received (%v)", jsonBytes, err)
		}
	}

	badJSON := []byte(`{"controllers":[{"not":"valid_structure"}]}`)
	err := json.Unmarshal(badJSON, &checkMarshal{})
	if err == nil {
		t.Error("expected invalid schema json to fail unmarshal")
	}
}

func TestMarshalSupportedControllers(t *testing.T) {
	for idx, controllerString := range ControllerStrings {
		controllerType := GetSupportedControllerFromString(controllerString)
		if idx == 0 {
			assert.Equal(t, SupportedController(0), controllerType)
		} else {
			assert.NotEqual(t, SupportedController(0), controllerType)
		}

		object := checkMarshal{
			Controllers: []SupportedController{controllerType},
		}
		_, err := json.Marshal(object)
		if idx == 0 {
			if err == nil {
				t.Errorf("Expected (%s) to throw an error. Reserving the first element in the enum to be an invalid config", controllerString)
			}
		} else if err != nil {
			t.Errorf("Could not write json output for element (%s); Received Error: (%s)", controllerString, err)
		}
	}
}

func TestCheckIfControllerKindIsConfiguredForValidation(t *testing.T) {
	config := Configuration{}
	for _, controllerString := range ControllerStrings[1:] {
		controllerEnum := GetSupportedControllerFromString(controllerString)
		assert.NotEqual(t, SupportedController(0), controllerEnum)
		config.ControllersToScan = append(config.ControllersToScan, controllerEnum)
	}

	validControllerKinds := []string{
		"deployment",
		"statefulset",
	}

	invalidControllerKinds := []string{
		"nonExistent",
	}

	for _, kind := range validControllerKinds {
		if ok := config.CheckIfKindIsConfiguredForValidation(kind); !ok {
			t.Errorf("Kind (%s) expected to be valid for configuration.", kind)
		}
	}

	for _, kind := range invalidControllerKinds {
		if ok := config.CheckIfKindIsConfiguredForValidation(kind); ok {
			t.Errorf("Kind (%s) should not be a valid controller to check", kind)
		}
	}
}

func TestGetSupportedControllerFromString(t *testing.T) {
	fixture := map[string]SupportedController{
		"":            Unsupported,
		"asdfasdf":    Unsupported,
		"\000":        Unsupported,
		"deployMENTS": Deployments,
		"JOB":         Jobs,
	}

	for inputString, expectedType := range fixture {
		resolvedType := GetSupportedControllerFromString(inputString)
		assert.Equal(t, expectedType, resolvedType, fmt.Sprintf("Expected (%s) to return (%s) controller type.", inputString, expectedType))
	}
}
