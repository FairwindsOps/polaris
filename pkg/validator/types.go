package validator

// NamespacedResult groups resource results by namespace.
type NamespacedResult struct {
	Summary *ResultSummary
	Results []ResourceResult
}

// NamespacedResults is a mapping of namespace name to the validation results.
type NamespacedResults map[string]*NamespacedResult

// ResourceResult groups container results by parent resource.
type ResourceResult struct {
	Name             string
	Type             string
	Summary          *ResultSummary
	ContainerResults []ContainerResult
	PodResults       []PodResult
}

// ResultSummary provides a high level overview of success, warnings, and failures.
type ResultSummary struct {
	Successes uint
	Warnings  uint
	Failures  uint
}

// ContainerResult provides a list of validation messages for each container.
type ContainerResult struct {
	Name     string
	Messages []ResultMessage
}

// PodResult provides a list of validation messages for each pod.
type PodResult struct {
	Name             string
	Messages         []ResultMessage
	ContainerResults []ContainerResult
}

// ResultMessage contains a message and a type indicator (success, warning, or failure).
type ResultMessage struct {
	Message string
	Type    string
}

// Score represents a percentage of validations that were successful.
func (rs *ResultSummary) Score() uint {
	return uint(float64(rs.Successes) / float64(rs.Successes+rs.Warnings+rs.Failures) * 100)
}

