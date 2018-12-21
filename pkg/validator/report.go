package validator

import (
	"fmt"
	"strings"
)

// Results contains the validation check results.
type Results struct {
	Pass                     bool
	FailMsg                  string
	ContainerValidations     []ContainerValidation
	InitContainerValidations []ContainerValidation
}

// Format structures the validation results to return back to k8s API.
func (r *Results) Format() (bool, string) {
	var sb strings.Builder
	r.Pass = true

	for _, cv := range r.ContainerValidations {
		if !containerValidation(sb, cv) {
			r.Pass = false
		}
	}

	for _, cv := range r.InitContainerValidations {
		if !containerValidation(sb, cv) {
			r.Pass = false
		}
	}

	r.FailMsg = sb.String()
	return r.Pass, r.FailMsg
}

func containerValidation(sb strings.Builder, cv ContainerValidation) bool {
	if len(cv.Failures) == 0 {
		return true
	}

	s := fmt.Sprintf("\nContainer: %s\n Failure/s:\n", cv.Container.Name)
	sb.WriteString(s)
	for _, failure := range cv.Failures {
		sb.WriteString(failure.Reason())
	}
	return false
}
