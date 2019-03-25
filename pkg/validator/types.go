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

