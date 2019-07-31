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

	"github.com/fairwindsops/polaris/pkg/config"
	conf "github.com/fairwindsops/polaris/pkg/config"
	corev1 "k8s.io/api/core/v1"
)

// MessageType represents the type of Message
type MessageType string

const (
	// MessageTypeSuccess indicates a validation success
	MessageTypeSuccess MessageType = "success"

	// MessageTypeWarning indicates a validation warning
	MessageTypeWarning MessageType = "warning"

	// MessageTypeError indicates a validation error
	MessageTypeError MessageType = "error"
)

// NamespaceResult groups container results by parent resource.
type NamespaceResult struct {
	Name    string
	Summary *ResultSummary

	// TODO: This struct could use some love to reorganize it as just having "results"
	//       and then having methods to return filtered results by type
	//       (deploy, daemonset, etc)
	//       The way this is structured right now makes it difficult to add
	//       additional result types and potentially miss things in the metrics
	//       summary.
	DeploymentResults            []ControllerResult
	StatefulSetResults           []ControllerResult
	DaemonSetResults             []ControllerResult
	JobResults                   []ControllerResult
	CronJobResults               []ControllerResult
	ReplicationControllerResults []ControllerResult
}

// AddResult adds a result to the result sets by leveraging the types supported by NamespaceResult
func (n *NamespaceResult) AddResult(resourceType config.SupportedController, result ControllerResult) error {
	// Iterate all the resource types supported in this struct
	var results *[]ControllerResult
	switch resourceType {
	case conf.Deployments:
		results = &n.DeploymentResults
	case conf.StatefulSets:
		results = &n.StatefulSetResults
	case conf.DaemonSets:
		results = &n.DaemonSetResults
	case conf.Jobs:
		results = &n.JobResults
	case conf.CronJobs:
		results = &n.CronJobResults
	case conf.ReplicationControllers:
		results = &n.ReplicationControllerResults
	default:
		return fmt.Errorf("Unknown Resource Type: (%s) Missing Implementation in NamespacedResult", resourceType)
	}

	// Append the new result to the results pointer loaded from the supported values
	*results = append(*results, result)

	return nil
}

// GetAllControllerResults grabs all the different types of controller results from the namespaced result as a single list for easier iteration
func (n NamespaceResult) GetAllControllerResults() []ControllerResult {
	all := []ControllerResult{}
	all = append(all, n.DeploymentResults...)
	all = append(all, n.StatefulSetResults...)
	all = append(all, n.DaemonSetResults...)
	all = append(all, n.JobResults...)
	all = append(all, n.CronJobResults...)
	all = append(all, n.ReplicationControllerResults...)

	return all
}

// NamespacedResults is a mapping of namespace name to the validation results.
type NamespacedResults map[string]*NamespaceResult

// GetAllControllerResults aggregates all the namespaced results in the set together
func (nsResults NamespacedResults) GetAllControllerResults() []ControllerResult {
	all := []ControllerResult{}
	for _, nsResult := range nsResults {
		all = append(all, nsResult.GetAllControllerResults()...)
	}
	return all
}

func (nsResults NamespacedResults) getNamespaceResult(nsName string) *NamespaceResult {
	nsResult := &NamespaceResult{}
	switch nsResults[nsName] {
	case nil:
		nsResult = &NamespaceResult{
			Summary:                      &ResultSummary{},
			DeploymentResults:            []ControllerResult{},
			StatefulSetResults:           []ControllerResult{},
			DaemonSetResults:             []ControllerResult{},
			JobResults:                   []ControllerResult{},
			CronJobResults:               []ControllerResult{},
			ReplicationControllerResults: []ControllerResult{},
		}
		nsResults[nsName] = nsResult
	default:
		nsResult = nsResults[nsName]
	}
	return nsResult
}

// CountSummary provides a high level overview of success, warnings, and errors.
type CountSummary struct {
	Successes uint
	Warnings  uint
	Errors    uint
}

// GetScore returns an overall score in [0, 100] for the CountSummary
func (cs *CountSummary) GetScore() uint {
	total := (cs.Successes * 2) + cs.Warnings + (cs.Errors * 2)
	return uint((float64(cs.Successes*2) / float64(total)) * 100)
}

func (cs *CountSummary) appendCounts(toAppend CountSummary) {
	cs.Errors += toAppend.Errors
	cs.Warnings += toAppend.Warnings
	cs.Successes += toAppend.Successes
}

// CategorySummary provides a map from category name to a CountSummary
type CategorySummary map[string]*CountSummary

// ResultSummary provides a high level overview of success, warnings, and errors.
type ResultSummary struct {
	Totals     CountSummary
	ByCategory CategorySummary
}

func (rs *ResultSummary) appendResults(toAppend ResultSummary) {
	rs.Totals.appendCounts(toAppend.Totals)
	for category, summary := range toAppend.ByCategory {
		if rs.ByCategory == nil {
			rs.ByCategory = CategorySummary{}
		}
		if _, exists := rs.ByCategory[category]; !exists {
			rs.ByCategory[category] = &CountSummary{}
		}
		rs.ByCategory[category].appendCounts(*summary)
	}
}

// ControllerResult provides a wrapper around a PodResult
type ControllerResult struct {
	Name      string
	Type      string
	PodResult PodResult
}

// ContainerResult provides a list of validation messages for each container.
type ContainerResult struct {
	Name     string
	Messages []*ResultMessage
	Summary  *ResultSummary
}

// PodResult provides a list of validation messages for each pod.
type PodResult struct {
	Name             string
	Summary          *ResultSummary
	Messages         []*ResultMessage
	ContainerResults []ContainerResult
	podSpec          corev1.PodSpec
}

// ResultMessage contains a message and a type indicator (success, warning, or error).
type ResultMessage struct {
	Message  string
	Type     MessageType
	Category string
}
