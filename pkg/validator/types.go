package validator

// namespaceResults: [{
//   name: "kube-system",
//   resourceResults: [{
//     name: "example-deployment",
//     type: "DaemonSet",
//     summary: {
//       successes: 28,
//       warnings: 12,
//       failures: 18,
//     },
//     messages: [{
//       message: "Resource requests are not set",
//       type: "failure",
//     }]
//   }]
// }]

type NamespacedResult struct {
	Namespace string
	Summary   *ResultSummary
	Results   []ResourceResult
}

type ResourceResult struct {
	Name     string
	Type     string
	Summary  *ResultSummary
	Messages []ResultMessage
}

type ResultSummary struct {
	Successes uint
	Warnings  uint
	Failures  uint
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
