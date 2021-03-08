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
	"time"

	"github.com/fairwindsops/polaris/pkg/config"
)

const (
	// PolarisOutputVersion is the version of the current output structure
	PolarisOutputVersion = "1.0"
)

const (
	successMessage = "✅ Success"
	dangerMessage  = "❌ Danger"
	warningMessage = "⚠️ Warning"
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

// RemoveSuccessfulResults remove all test that have passed.
func (res *AuditData) RemoveSuccessfulResults() {
	for _, auditDataResult := range res.Results {
		auditDataResult.removeSuccessfulResults()
	}
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
	ID       string
	Message  string
	Success  bool
	Severity config.Severity
	Category string
}

// ResultSet contiains the results for a set of checks
type ResultSet map[string]ResultMessage

func (res ResultSet) removeSuccessfulResults() {
	for k, resultMessage := range res {
		if resultMessage.Success {
			delete(res, k)
		}
	}
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

func (res *Result) removeSuccessfulResults() {
	res.Results.removeSuccessfulResults()
	res.PodResult.removeSuccessfulResults()
}

// PodResult provides a list of validation messages for each pod.
type PodResult struct {
	Name             string
	Results          ResultSet
	ContainerResults []ContainerResult
}

func (res *PodResult) removeSuccessfulResults() {
	if res == nil {
		return
	}
	res.Results.removeSuccessfulResults()
	for _, containerResult := range res.ContainerResults {
		containerResult.removeSuccessfulResults()
	}
}

// ContainerResult provides a list of validation messages for each container.
type ContainerResult struct {
	Name    string
	Results ResultSet
}

func (res *ContainerResult) removeSuccessfulResults() {
	res.Results.removeSuccessfulResults()
}

func (res AuditData) GetPrettyOutput() string {
	str := fmt.Sprintf("Polaris audited %s %s at %s\n", res.SourceType, res.SourceName, res.AuditTime)
	str += fmt.Sprintf("Nodes: %d Namespaces: %d | Controllers: %d\n", res.ClusterInfo.Nodes, res.ClusterInfo.Namespaces, res.ClusterInfo.Controllers)
	str += fmt.Sprintf("Final score: %d\n", res.Score)
	str += "\n"
	for _, result := range res.Results {
		str += result.GetPrettyOutput() + "\n"
	}
	return str
}

func (res Result) GetPrettyOutput() string {
	str := fmt.Sprintf("%s %s in namespace %s\n", res.Kind, res.Name, res.Namespace)
	str += res.Results.GetPrettyOutput()
	if res.PodResult != nil {
		str += res.PodResult.GetPrettyOutput()
	}
	return str
}

func (res PodResult) GetPrettyOutput() string {
	str := res.Results.GetPrettyOutput()
	for _, cont := range res.ContainerResults {
		str += cont.GetPrettyOutput() + "\n"
	}
	return str
}

func (res ContainerResult) GetPrettyOutput() string {
	str := fmt.Sprintf("  container %s\n", res.Name)
	str += res.Results.GetPrettyOutput()
	return str
}

func (res ResultSet) GetPrettyOutput() string {
	str := ""
	for _, msg := range res {
		status := successMessage
		if !msg.Success {
			if msg.Severity == config.SeverityWarning {
				status = warningMessage
			} else {
				status = dangerMessage
			}
		}
		str += fmt.Sprintf("  %s: %s", msg.ID, status)
		str += fmt.Sprintf("      %s - %s", msg.Category, msg.Message)
		str += "\n"
	}
	return str
}
