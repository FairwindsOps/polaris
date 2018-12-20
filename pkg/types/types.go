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

// NewFailure is a factory function for a Failure.
func NewFailure(name, expected, actual string) *Failure {
	return &Failure{
		Name:     name,
		Expected: expected,
		Actual:   actual,
	}
}

// Reason returns a string that describes the reason for a Failure.
func (f *Failure) Reason() string {
	return fmt.Sprintf("- %s: Expected: %s, Actual: %s.\n",
		f.Name,
		f.Expected,
		f.Actual,
	)
}

// ContainerResults has the results of the validation checks for containers.
type ContainerResults struct {
	Name     string
	Failures []Failure
}

// AddFailure creates a new Failure and adds it to ContainerResults.
func (c *ContainerResults) AddFailure(name, expected, actual string) {
	f := NewFailure(name, expected, actual)
	c.Failures = append(c.Failures, *f)
}
