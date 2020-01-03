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
	Results              []ControllerResult
}

// ClusterInfo contains Polaris results as well as some high-level stats
type ClusterInfo struct {
	Version                string
	Nodes                  int
	Pods                   int
	Namespaces             int
	Deployments            int
	StatefulSets           int
	DaemonSets             int
	Jobs                   int
	CronJobs               int
	ReplicationControllers int
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

// ControllerResult provides results for a controller
type ControllerResult struct {
	Name      string
	Kind      string
	Messages  ResultSet
	PodResult PodResult
}

// PodResult provides a list of validation messages for each pod.
type PodResult struct {
	Name             string
	Messages         ResultSet
	ContainerResults []ContainerResult
}

// ContainerResult provides a list of validation messages for each container.
type ContainerResult struct {
	Name     string
	Messages ResultSet
}

// CountSummary provides a high level overview of success, warnings, and errors.
type CountSummary struct {
	Successes uint
	Warnings  uint
	Errors    uint
}

// GetScore returns an overall score in [0, 100] for the CountSummary
func (cs CountSummary) GetScore() uint {
	total := (cs.Successes * 2) + cs.Warnings + (cs.Errors * 2)
	return uint((float64(cs.Successes*2) / float64(total)) * 100)
}

func (cs *CountSummary) AddSummary(other CountSummary) {
	cs.Successes += other.Successes
	cs.Warnings += other.Warnings
	cs.Errors += other.Errors
}

// CategorySummary provides a map from category name to a CountSummary
type CategorySummary map[string]*CountSummary

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

func (p PodResult) GetSummary() CountSummary {
	summary := p.Messages.GetSummary()
	for _, containerResult := range p.ContainerResults {
		summary.AddSummary(containerResult.Messages.GetSummary())
	}
	return summary
}

func (c ControllerResult) GetSummary() CountSummary {
	summary := c.Messages.GetSummary()
	summary.AddSummary(c.PodResult.GetSummary())
	return summary
}

func (a AuditData) GetSummary() CountSummary {
	summary := CountSummary{}
	for _, ctrlResult := range a.Results {
		summary.AddSummary(ctrlResult.GetSummary())
	}
	return summary
}

func (rs ResultSet) GetSuccesses() []ResultMessage {
	successes := []ResultMessage{}
	for _, msg := range rs {
		if msg.Success {
			successes = append(successes, msg)
		}
	}
	return successes
}

func (rs ResultSet) GetWarnings() []ResultMessage {
	warnings := []ResultMessage{}
	for _, msg := range rs {
		if msg.Success == false && msg.Severity == config.SeverityWarning {
			warnings = append(warnings, msg)
		}
	}
	return warnings
}

func (rs ResultSet) GetErrors() []ResultMessage {
	errors := []ResultMessage{}
	for _, msg := range rs {
		if msg.Success == false && msg.Severity == config.SeverityError {
			errors = append(errors, msg)
		}
	}
	return errors
}
