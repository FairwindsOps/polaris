package validator

type NamespacedResult struct {
	Summary *ResultSummary
	Results []ResourceResult
}

type NamespacedResults map[string]*NamespacedResult

type ResourceResult struct {
	Name             string
	Type             string
	Summary          *ResultSummary
	ContainerResults []ContainerResult
}

type ResultSummary struct {
	Successes uint
	Warnings  uint
	Failures  uint
}

type ContainerResult struct {
	Name     string
	Messages []ResultMessage
}

type ResultMessage struct {
	Message string
	Type    string
}

func (rs *ResultSummary) Score() uint {
	return uint(float64(rs.Successes) / float64(rs.Successes+rs.Warnings+rs.Failures) * 100)
}

func (rs *ResultSummary) WarningWidth(fullWidth uint) uint {
	return uint(float64(rs.Successes+rs.Warnings) / float64(rs.Successes+rs.Warnings+rs.Failures) * float64(fullWidth))
}

func (rs *ResultSummary) SuccessWidth(fullWidth uint) uint {
	return uint(float64(rs.Successes) / float64(rs.Successes+rs.Warnings+rs.Failures) * float64(fullWidth))
}

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
