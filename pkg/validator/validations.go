package validator

import (
	"encoding/json"
	"fmt"

	"github.com/qri-io/jsonschema"
)

type customValidator func(data interface{}) (bool, []jsonschema.ValError, error)

// customValidators is a map of validation functions that can be used in schema checks
// sometimes we need to validate things that aren't covered by the JSON validation schema
var customValidators = map[string]customValidator{
	"hpaMaxAvailability": validateHPAMaxAvailability,
}

type HorizontalPodAutoscalerView struct {
	Spec struct {
		MinReplicas *int `json:"minReplicas"`
		MaxReplicas int  `json:"maxReplicas"`
	} `json:"spec"`
}

func validateHPAMaxAvailability(data any) (bool, []jsonschema.ValError, error) {
	jsonString, err := json.Marshal(data)
	if err != nil {
		return false, nil, err
	}

	hpa := HorizontalPodAutoscalerView{}
	err = json.Unmarshal(jsonString, &hpa)
	if err != nil {
		return false, nil, err
	}

	if hpa.Spec.MinReplicas == nil {
		return true, []jsonschema.ValError{}, nil
	}

	if hpa.Spec.MaxReplicas != *hpa.Spec.MinReplicas {
		return true, []jsonschema.ValError{}, nil
	}

	return false, []jsonschema.ValError{{PropertyPath: "spec.maxReplicas", Message: fmt.Sprintf("maxReplicas (%d) and minReplicas (%d) should be different", hpa.Spec.MaxReplicas, *hpa.Spec.MinReplicas)}}, nil
}
