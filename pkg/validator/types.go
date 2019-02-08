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

// ResultMessage contains a message and a type indicator (success, warning, or failure).
type ResultMessage struct {
	Message string
	Type    string
}

// Score represents a percentage of validations that were successful.
func (rs *ResultSummary) Score() uint {
	return uint(float64(rs.Successes) / float64(rs.Successes+rs.Warnings+rs.Failures) * 100)
}

// WarningWidth is a UI specific helper that helps determine the width of a progress bar.
func (rs *ResultSummary) WarningWidth(fullWidth uint) uint {
	return uint(float64(rs.Successes+rs.Warnings) / float64(rs.Successes+rs.Warnings+rs.Failures) * float64(fullWidth))
}

// SuccessWidth is a UI specific helper that helps determine the width of a progress bar.
func (rs *ResultSummary) SuccessWidth(fullWidth uint) uint {
	return uint(float64(rs.Successes) / float64(rs.Successes+rs.Warnings+rs.Failures) * float64(fullWidth))
}

// HTMLSpecialCharCode is a UI specific helper that provides an HTML char code.
func (rm *ResultMessage) HTMLSpecialCharCode() string {
	switch rm.Type {
	case "success":
		return "9745"
	case "warning":
		return "9888"
	default:
		return "9746"
	}
}
