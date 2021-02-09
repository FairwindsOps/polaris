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
	"time"

	"github.com/fairwindsops/polaris/pkg/config"
)

const (
	// PolarisOutputVersion is the version of the current output structure
	PolarisOutputVersion = "1.0"
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
