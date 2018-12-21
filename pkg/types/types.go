package types

import (
	"fmt"
)

// Failure contains information about the failing validation.
type Failure struct {
	Name     string
	Expected string
	Actual   string
}

// Reason returns a string that describes the reason for a Failure.
func (f *Failure) Reason() string {
	return fmt.Sprintf("- %s: Expected: %s, Actual: %s.\n",
		f.Name,
		f.Expected,
		f.Actual,
	)
}
