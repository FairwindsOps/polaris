package report

import (
	"fmt"
	"strings"

	"github.com/reactiveops/fairwinds/pkg/types"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("Fairwinds report")

// Results contains the validation check results.
type Results struct {
	Pass           bool
	FailMsg        string
	Containers     []types.ContainerResults
	InitContainers []types.ContainerResults
}

// Format structures the validation results to return back to k8s API.
func (r *Results) Format() (bool, string) {
	var sb strings.Builder

	for _, container := range r.Containers {
		if len(container.Failures) == 0 {
			r.Pass = true
		}

		r.Pass = false
		s := fmt.Sprintf("\nContainer: %s\n Failure/s:\n", container.Name)
		sb.WriteString(s)
		for _, failure := range container.Failures {
			sb.WriteString(failure.Reason())
		}
	}

	for _, container := range r.InitContainers {
		if len(container.Failures) == 0 && r.Pass == true {
			return r.Pass, r.FailMsg
		}

		r.Pass = false
		s := fmt.Sprintf("\nInitContainer: %s\n Failure/s:\n", container.Name)
		sb.WriteString(s)
		for _, failure := range container.Failures {
			sb.WriteString(failure.Reason())
		}
	}

	r.FailMsg = sb.String()
	return r.Pass, r.FailMsg
}
