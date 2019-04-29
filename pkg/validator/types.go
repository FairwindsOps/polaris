// Copyright 2019 ReactiveOps
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

// NamespacedResults is a mapping of namespace name to the validation results.
type NamespacedResults map[string]*NamespaceResult

// NamespaceResult groups container results by parent resource.
type NamespaceResult struct {
	Name       string
	Summary    *ResultSummary
	PodResults []PodResult
}

// CountSummary provides a high level overview of success, warnings, and errors.
type CountSummary struct {
	Successes uint
	Warnings  uint
	Errors    uint
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
}

// ResultMessage contains a message and a type indicator (success, warning, or error).
type ResultMessage struct {
	Message  string
	Type     MessageType
	Category string
}
