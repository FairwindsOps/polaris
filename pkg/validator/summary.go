package validator

import (
	"github.com/fairwindsops/polaris/pkg/config"
)

// CountSummary provides a high level overview of success, warnings, and errors.
type CountSummary struct {
	Successes uint
	Warnings  uint
	Errors    uint
}

// CountSummaryByCategory is a map from category to CountSummary
type CountSummaryByCategory map[string]CountSummary

// GetScore returns an overall score in [0, 100] for the CountSummary
func (cs CountSummary) GetScore() uint {
	total := (cs.Successes * 2) + cs.Warnings + (cs.Errors * 2)
	return uint((float64(cs.Successes*2) / float64(total)) * 100)
}

// AddSummary adds two CountSummaries together
func (cs *CountSummary) AddSummary(other CountSummary) {
	cs.Successes += other.Successes
	cs.Warnings += other.Warnings
	cs.Errors += other.Errors
}

// AddSummary adds two CountSummaryByCategories together
func (csc CountSummaryByCategory) AddSummary(other CountSummaryByCategory) {
	categories := []string{}
	for cat, _ := range csc {
		categories = append(categories, cat)
	}
	for cat, _ := range other {
		categories = append(categories, cat)
	}
	for _, cat := range categories {
		cur := csc[cat]
		cur.AddSummary(other[cat])
		csc[cat] = cur
	}
}

// GetSummary summarizes a ResultSet
func (rs ResultSet) GetSummary() CountSummary {
	cs := CountSummary{}
	for _, result := range rs {
		if result.Success == false {
			if result.Severity == config.SeverityWarning {
				cs.Warnings += 1
			} else {
				cs.Errors += 1
			}
		} else {
			cs.Successes += 1
		}
	}
	return cs
}

// GetSummaryByCategory summarizes a ResultSet
func (rs ResultSet) GetSummaryByCategory() CountSummaryByCategory {
	summaries := CountSummaryByCategory{}
	for _, result := range rs {
		cs, ok := summaries[result.Category]
		if !ok {
			cs = CountSummary{}
		}
		if result.Success == false {
			if result.Severity == config.SeverityWarning {
				cs.Warnings += 1
			} else {
				cs.Errors += 1
			}
		} else {
			cs.Successes += 1
		}
		summaries[result.Category] = cs
	}
	return summaries
}

// GetSummary summarizes a PodResult
func (p PodResult) GetSummary() CountSummary {
	summary := p.Results.GetSummary()
	for _, containerResult := range p.ContainerResults {
		summary.AddSummary(containerResult.Results.GetSummary())
	}
	return summary
}

// GetSummaryByCategory summarizes a PodResult
func (p PodResult) GetSummaryByCategory() CountSummaryByCategory {
	summaries := p.Results.GetSummaryByCategory()
	for _, containerResult := range p.ContainerResults {
		summaries.AddSummary(containerResult.Results.GetSummaryByCategory())
	}
	return summaries
}

// GetSummary summarizes a ControllerResult
func (c ControllerResult) GetSummary() CountSummary {
	summary := c.Results.GetSummary()
	summary.AddSummary(c.PodResult.GetSummary())
	return summary
}

// GetSummaryByCategory summarizes a ControllerResult
func (c ControllerResult) GetSummaryByCategory() CountSummaryByCategory {
	summary := c.Results.GetSummaryByCategory()
	summary.AddSummary(c.PodResult.GetSummaryByCategory())
	return summary
}

// GetSummary summarizes AuditData
func (a AuditData) GetSummary() CountSummary {
	summary := CountSummary{}
	for _, ctrlResult := range a.Results {
		summary.AddSummary(ctrlResult.GetSummary())
	}
	return summary
}

// GetSummaryByCategory summarizes AuditData
func (a AuditData) GetSummaryByCategory() CountSummaryByCategory {
	summaries := CountSummaryByCategory{}
	for _, ctrlResult := range a.Results {
		summaries.AddSummary(ctrlResult.GetSummaryByCategory())
	}
	return summaries
}

// GetResultsByNamespace organizes results by namespace
func (a AuditData) GetResultsByNamespace() map[string][]*ControllerResult {
	allResults := map[string][]*ControllerResult{}
	for idx, ctrlResult := range a.Results {
		nsResults, ok := allResults[ctrlResult.Namespace]
		if !ok {
			nsResults = []*ControllerResult{}
		}
		nsResults = append(nsResults, &a.Results[idx])
		allResults[ctrlResult.Namespace] = nsResults
	}
	return allResults
}

// GetSuccesses returns the success messages in a result set
func (rs ResultSet) GetSuccesses() []ResultMessage {
	successes := []ResultMessage{}
	for _, msg := range rs {
		if msg.Success {
			successes = append(successes, msg)
		}
	}
	return successes
}

// GetWarnings returns the warning messages in a result set
func (rs ResultSet) GetWarnings() []ResultMessage {
	warnings := []ResultMessage{}
	for _, msg := range rs {
		if msg.Success == false && msg.Severity == config.SeverityWarning {
			warnings = append(warnings, msg)
		}
	}
	return warnings
}

// GetErrors returns the error messages in a result set
func (rs ResultSet) GetErrors() []ResultMessage {
	errors := []ResultMessage{}
	for _, msg := range rs {
		if msg.Success == false && msg.Severity == config.SeverityError {
			errors = append(errors, msg)
		}
	}
	return errors
}

// GetSortedMessages returns messages sorted as errors, then warnings, then successes
func (rs ResultSet) GetSortedMessages() []ResultMessage {
	messages := []ResultMessage{}
	messages = append(messages, rs.GetErrors()...)
	messages = append(messages, rs.GetWarnings()...)
	messages = append(messages, rs.GetSuccesses()...)
	return messages
}
