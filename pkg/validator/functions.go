package validator

import (
	"encoding/json"
	"fmt"

	"github.com/qri-io/jsonschema"
)

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

	if hpa.Spec.MaxReplicas > *hpa.Spec.MinReplicas {
		return true, []jsonschema.ValError{}, nil
	}

	return false, []jsonschema.ValError{{PropertyPath: "spec.maxReplicas", Message: fmt.Sprintf("maxReplicas (%d) must be greater than minReplicas (%d)", hpa.Spec.MaxReplicas, *hpa.Spec.MinReplicas)}}, nil
}
