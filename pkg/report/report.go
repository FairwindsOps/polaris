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

func reason(f types.Failure) string {
	return fmt.Sprintf("- %s: Expected: %s, Actual: %s.\n",
		f.Name,
		f.Expected,
		f.Actual,
	)
}

// Format structures the validation results to return back to k8s API.
func (r *Results) Format() (bool, string) {
	var sb strings.Builder

	for _, ctr := range r.Containers {
		if len(ctr.Failures) == 0 {
			r.Pass = true
		}

		r.Pass = false
		s := fmt.Sprintf("\nContainer: %s\n Failure/s:\n", ctr.Name)
		sb.WriteString(s)
		for _, failure := range ctr.Failures {
			sb.WriteString(reason(failure))
		}
	}

	for _, ctr := range r.InitContainers {
		if len(ctr.Failures) == 0 && r.Pass == true {
			return r.Pass, r.FailMsg
		}

		r.Pass = false
		s := fmt.Sprintf("\nInitContainer: %s\n Failure/s:\n", ctr.Name)
		sb.WriteString(s)
		for _, failure := range ctr.Failures {
			sb.WriteString(reason(failure))
		}
	}

	r.FailMsg = sb.String()
	return r.Pass, r.FailMsg
}
