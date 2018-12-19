package types

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

// ContainerResults has the results of the validation checks for containers.
type ContainerResults struct {
	Name     string
	Failures []Failure
}
