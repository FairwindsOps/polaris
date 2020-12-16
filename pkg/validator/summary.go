package validator

import (
	"sort"
	"strings"

	"github.com/fairwindsops/polaris/pkg/config"
)

// CountSummary provides a high level overview of success, warnings, and errors.
type CountSummary struct {
	Successes uint
	Warnings  uint
	Dangers    uint
}

// CountSummaryByCategory is a map from category to CountSummary
type CountSummaryByCategory map[string]CountSummary

// GetScore returns an overall score in [0, 100] for the CountSummary
func (cs CountSummary) GetScore() uint {
	total := (cs.Successes * 2) + cs.Warnings + (cs.Dangers * 2)
	if total == 0 {
		return 0 // Prevent divide by 0.
	}
	return uint((float64(cs.Successes*2) / float64(total)) * 100)
}

// AddSummary adds two CountSummaries together
func (cs *CountSummary) AddSummary(other CountSummary) {
	cs.Successes += other.Successes
	cs.Warnings += other.Warnings
	cs.Dangers += other.Dangers
}

// AddResult adds a single result to the summary
func (cs *CountSummary) AddResult(result ResultMessage) {
	if result.Success == false {
		if result.Severity == config.SeverityWarning {
			cs.Warnings++
		} else {
			cs.Dangers++
		}
	} else {
		cs.Successes++
	}
}

// AddSummary adds two CountSummaryByCategories together
func (csc CountSummaryByCategory) AddSummary(other CountSummaryByCategory) {
	categories := []string{}
	for cat := range csc {
		categories = append(categories, cat)
	}
	for cat := range other {
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
		cs.AddResult(result)
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
		cs.AddResult(result)
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

// GetSummary summarizes a Result
func (c Result) GetSummary() CountSummary {
	summary := c.Results.GetSummary()
	summary.AddSummary(c.PodResult.GetSummary())
	return summary
}

// GetSummaryByCategory summarizes a Result
func (c Result) GetSummaryByCategory() CountSummaryByCategory {
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
func (a AuditData) GetResultsByNamespace() map[string][]*Result {
	allResults := map[string][]*Result{}
	for idx, ctrlResult := range a.Results {
		nsResults, ok := allResults[ctrlResult.Namespace]
		if !ok {
			nsResults = []*Result{}
		}
		nsResults = append(nsResults, &a.Results[idx])
		allResults[ctrlResult.Namespace] = nsResults
	}
	for ns := range allResults {
		sort.SliceStable(allResults[ns], func(i, j int) bool {
			kind := strings.Compare(allResults[ns][i].Kind, allResults[ns][j].Kind)
			if kind != 0 {
				return kind == -1
			}
			name := strings.Compare(allResults[ns][i].Name, allResults[ns][j].Name)
			if name != 0 {
				return name == -1
			}
			return true
		})
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

// GetDangers returns the error messages in a result set
func (rs ResultSet) GetDangers() []ResultMessage {
	errors := []ResultMessage{}
	for _, msg := range rs {
		if msg.Success == false && msg.Severity == config.SeverityDanger {
			errors = append(errors, msg)
		}
	}
	return errors
}

// GetSortedResults returns messages sorted as errors, then warnings, then successes
func (rs ResultSet) GetSortedResults() []ResultMessage {
	messages := []ResultMessage{}
	messages = append(messages, rs.GetDangers()...)
	messages = append(messages, rs.GetWarnings()...)
	messages = append(messages, rs.GetSuccesses()...)
	return messages
}
