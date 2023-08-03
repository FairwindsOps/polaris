// Copyright 2019 FairwindsOps Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package validator

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/thoas/go-funk"

	"github.com/fairwindsops/polaris/pkg/config"
)

const (
	// PolarisOutputVersion is the version of the current output structure
	PolarisOutputVersion = "1.0"
)

var (
	successMessage = "🎉 Success"
	dangerMessage  = "❌ Danger"
	warningMessage = "😬 Warning"
)

var (
	titleColor = color.New(color.FgBlue).Add(color.Bold)
	checkColor = color.New(color.FgCyan)
)

// AuditData contains all the data from a full Polaris audit
type AuditData struct {
	PolarisOutputVersion string
	AuditTime            string
	SourceType           string
	SourceName           string
	DisplayName          string
	ClusterInfo          ClusterInfo
	Results              []Result
	Score                uint
}

// FilterResultsBySeverityLevel includes results according to the provided severity level:
// 'danger' is the least verbose, 'warning' is medium verbosity, default behavior will
// include all results, which currently also includes 'ignore'
func (res AuditData) FilterResultsBySeverityLevel(severityLevel config.Severity) AuditData {
	resCopy := res
	resCopy.Results = []Result{}

	filteredResults := funk.Map(res.Results, func(auditDataResult Result) Result {
		return auditDataResult.filterResultsBySeverityLevel(severityLevel)
	}).([]Result)

	for _, result := range filteredResults {
		if result.isNotEmpty() {
			resCopy.Results = append(resCopy.Results, result)
		}
	}

	return resCopy
}

// RemoveSuccessfulResults removes all tests that have passed
func (res AuditData) RemoveSuccessfulResults() AuditData {
	resCopy := res
	resCopy.Results = []Result{}

	filteredResults := funk.Map(res.Results, func(auditDataResult Result) Result {
		return auditDataResult.removeSuccessfulResults()
	}).([]Result)

	for _, result := range filteredResults {
		if result.isNotEmpty() {
			resCopy.Results = append(resCopy.Results, result)
		}
	}

	return resCopy
}

// ClusterInfo contains Polaris results as well as some high-level stats
type ClusterInfo struct {
	Version     string
	Nodes       int
	Pods        int
	Namespaces  int
	Controllers int
}

// ResultMessage is the result of a given check
type ResultMessage struct {
	ID        string
	Message   string
	Details   []string
	Success   bool
	Severity  config.Severity
	Category  string
	Mutations []config.Mutation
}

// ResultSet contiains the results for a set of checks
type ResultSet map[string]ResultMessage

func (res ResultSet) isNotEmpty() bool {
	return len(res) > 0
}

func (res ResultSet) removeSuccessfulResults() ResultSet {
	newResults := ResultSet{}
	for k, resultMessage := range res {
		if !resultMessage.Success {
			newResults[k] = resultMessage
		}
	}
	return newResults
}

func (res ResultSet) filterResultsBySeverityLevel(severityLevel config.Severity) ResultSet {
	newResults := ResultSet{}
	for k, resultMessage := range res {
		switch severityLevel {
		case config.SeverityDanger:
			if resultMessage.Severity == config.SeverityDanger {
				newResults[k] = resultMessage
			}
		case config.SeverityWarning:
			if resultMessage.Severity == config.SeverityDanger || resultMessage.Severity == config.SeverityWarning {
				newResults[k] = resultMessage
			}
		default:
			return res
		}
	}
	return newResults
}

// Result provides results for a Kubernetes object
type Result struct {
	Name        string
	Namespace   string
	Kind        string
	Results     ResultSet
	PodResult   *PodResult
	CreatedTime time.Time
}

func (res Result) removeSuccessfulResults() Result {
	resCopy := res
	resCopy.Results = res.Results.removeSuccessfulResults()
	if res.PodResult != nil {
		podCopy := res.PodResult.removeSuccessfulResults()
		resCopy.PodResult = &podCopy
	}
	return resCopy
}

func (res Result) filterResultsBySeverityLevel(severityLevel config.Severity) Result {
	resCopy := res
	resCopy.Results = res.Results.filterResultsBySeverityLevel(severityLevel)
	if res.PodResult != nil {
		podCopy := res.PodResult.filterResultsBySeverityLevel(severityLevel)
		resCopy.PodResult = &podCopy
	}
	return resCopy
}

func (res Result) isNotEmpty() bool {
	if res.PodResult != nil {
		return res.PodResult.isNotEmpty()
	}
	return res.Results.isNotEmpty()
}

// PodResult provides a list of validation messages for each pod.
type PodResult struct {
	Name             string
	Results          ResultSet
	ContainerResults []ContainerResult
}

func (res PodResult) removeSuccessfulResults() PodResult {
	resCopy := PodResult{}
	resCopy.Results = res.Results.removeSuccessfulResults()
	resCopy.ContainerResults = funk.Map(res.ContainerResults, func(containerResult ContainerResult) ContainerResult {
		return containerResult.removeSuccessfulResults()
	}).([]ContainerResult)
	return resCopy
}

func (res PodResult) filterResultsBySeverityLevel(severityLevel config.Severity) PodResult {
	resCopy := PodResult{}
	resCopy.Results = res.Results.filterResultsBySeverityLevel(severityLevel)
	resCopy.ContainerResults = funk.Map(res.ContainerResults, func(containerResult ContainerResult) ContainerResult {
		return containerResult.filterResultsBySeverityLevel(severityLevel)
	}).([]ContainerResult)
	return resCopy
}

func (res PodResult) isNotEmpty() bool {
	for _, cr := range res.ContainerResults {
		if cr.isNotEmpty() {
			return true
		}
	}
	return res.Results.isNotEmpty()
}

// ContainerResult provides a list of validation messages for each container.
type ContainerResult struct {
	Name    string
	Results ResultSet
}

func (res ContainerResult) removeSuccessfulResults() ContainerResult {
	resCopy := res
	resCopy.Results = res.Results.removeSuccessfulResults()
	return resCopy
}

func (res ContainerResult) filterResultsBySeverityLevel(severityLevel config.Severity) ContainerResult {
	resCopy := res
	resCopy.Results = res.Results.filterResultsBySeverityLevel(severityLevel)
	return resCopy
}

func (res ContainerResult) isNotEmpty() bool {
	return res.Results.isNotEmpty()
}

func fillString(id string, l int) string {
	for len(id) < l {
		id += " "
	}
	return id
}

// GetPrettyOutput returns a human-readable string
func (res AuditData) GetPrettyOutput(useColor bool) string {
	color.NoColor = !useColor
	str := titleColor.Sprint(fmt.Sprintf("Polaris audited %s %s at %s\n", res.SourceType, res.SourceName, res.AuditTime))
	str += color.CyanString(fmt.Sprintf("    Nodes: %d | Namespaces: %d | Controllers: %d\n", res.ClusterInfo.Nodes, res.ClusterInfo.Namespaces, res.ClusterInfo.Controllers))
	str += color.GreenString(fmt.Sprintf("    Final score: %d\n", res.Score))
	str += "\n"
	for _, result := range res.Results {
		str += result.GetPrettyOutput() + "\n"
	}
	color.NoColor = false
	return str
}

// GetPrettyOutput returns a human-readable string
func (res Result) GetPrettyOutput() string {
	str := titleColor.Sprint(fmt.Sprintf("%s %s", res.Kind, res.Name))
	if res.Namespace != "" {
		str += titleColor.Sprint(fmt.Sprintf(" in namespace %s", res.Namespace))
	}
	str += "\n"
	str += res.Results.GetPrettyOutput()
	if res.PodResult != nil {
		str += res.PodResult.GetPrettyOutput()
	}
	return str
}

// GetPrettyOutput returns a human-readable string
func (res PodResult) GetPrettyOutput() string {
	str := res.Results.GetPrettyOutput()
	for _, cont := range res.ContainerResults {
		str += cont.GetPrettyOutput()
	}
	return str
}

// GetPrettyOutput returns a human-readable string
func (res ContainerResult) GetPrettyOutput() string {
	str := titleColor.Sprint(fmt.Sprintf("  Container %s\n", res.Name))
	str += res.Results.GetPrettyOutput()
	return str
}

const minIDLength = 40

// GetPrettyOutput returns a human-readable string
func (res ResultSet) GetPrettyOutput() string {
	indent := "    "
	str := ""
	for _, msg := range res {
		status := color.GreenString(successMessage)
		if !msg.Success {
			if msg.Severity == config.SeverityWarning {
				status = color.YellowString(warningMessage)
			} else {
				status = color.RedString(dangerMessage)
			}
		}
		if color.NoColor {
			status = strings.Fields(status)[1] // remove emoji
		}
		str += fmt.Sprintf("%s%s %s\n", indent, checkColor.Sprint(fillString(msg.ID, minIDLength-len(indent))), status)
		str += fmt.Sprintf("%s    %s - %s\n", indent, msg.Category, msg.Message)
	}
	return str
}
