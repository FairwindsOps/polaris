// Copyright 2026 FairwindsOps Inc
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
	"encoding/json"
	"fmt"
	"sort"

	"github.com/fairwindsops/polaris/pkg/config"
)

const (
	sarifVersion        = "2.1.0"
	sarifSchema         = "https://json.schemastore.org/sarif-2.1.0.json"
	sarifToolName       = "Polaris"
	sarifInformationURI = "https://github.com/FairwindsOps/polaris"
)

// sarifReport is the top-level SARIF 2.1.0 document.
// See https://docs.oasis-open.org/sarif/sarif/v2.1.0/sarif-v2.1.0.html
type sarifReport struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []sarifRun `json:"runs"`
}

type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
}

type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

type sarifDriver struct {
	Name           string      `json:"name"`
	InformationURI string      `json:"informationUri"`
	Rules          []sarifRule `json:"rules"`
}

type sarifRule struct {
	ID               string               `json:"id"`
	Name             string               `json:"name,omitempty"`
	ShortDescription sarifMessage         `json:"shortDescription"`
	Properties       *sarifRuleProperties `json:"properties,omitempty"`
}

type sarifRuleProperties struct {
	Category string `json:"category,omitempty"`
}

type sarifResult struct {
	RuleID    string          `json:"ruleId"`
	Level     string          `json:"level"`
	Message   sarifMessage    `json:"message"`
	Locations []sarifLocation `json:"locations,omitempty"`
}

type sarifMessage struct {
	Text string `json:"text"`
}

type sarifLocation struct {
	LogicalLocations []sarifLogicalLocation `json:"logicalLocations,omitempty"`
}

type sarifLogicalLocation struct {
	FullyQualifiedName string `json:"fullyQualifiedName"`
	Kind               string `json:"kind,omitempty"`
}

// sarifLevel maps a Polaris severity to a SARIF result level.
func sarifLevel(severity config.Severity) string {
	switch severity {
	case config.SeverityDanger:
		return "error"
	case config.SeverityWarning:
		return "warning"
	case config.SeverityIgnore:
		return "note"
	default:
		return "none"
	}
}

// resourceLocation builds a SARIF logical-location name for a Kubernetes resource.
func resourceLocation(namespace, kind, name string) string {
	if namespace != "" {
		return fmt.Sprintf("%s/%s/%s", namespace, kind, name)
	}
	return fmt.Sprintf("%s/%s", kind, name)
}

// GetSarifOutput returns the audit results encoded as a SARIF 2.1.0 report.
//
// Following SARIF's finding-oriented semantics, only failed checks are emitted
// as results. Each result carries a logical location identifying the Kubernetes
// resource (and container, when the check applies to one).
func (res AuditData) GetSarifOutput() ([]byte, error) {
	rules := map[string]sarifRule{}
	results := []sarifResult{}

	addMessages := func(messages ResultSet, location string) {
		for _, msg := range messages {
			if msg.Success {
				continue
			}
			if _, ok := rules[msg.ID]; !ok {
				rule := sarifRule{
					ID:               msg.ID,
					Name:             msg.ID,
					ShortDescription: sarifMessage{Text: msg.Message},
				}
				if msg.Category != "" {
					rule.Properties = &sarifRuleProperties{Category: msg.Category}
				}
				rules[msg.ID] = rule
			}
			results = append(results, sarifResult{
				RuleID:  msg.ID,
				Level:   sarifLevel(msg.Severity),
				Message: sarifMessage{Text: msg.Message},
				Locations: []sarifLocation{{
					LogicalLocations: []sarifLogicalLocation{{
						FullyQualifiedName: location,
						Kind:               "resource",
					}},
				}},
			})
		}
	}

	for _, result := range res.Results {
		base := resourceLocation(result.Namespace, result.Kind, result.Name)
		addMessages(result.Results, base)
		if result.PodResult != nil {
			addMessages(result.PodResult.Results, base)
			for _, cr := range result.PodResult.ContainerResults {
				addMessages(cr.Results, base+"/container/"+cr.Name)
			}
		}
	}

	// Sort results for deterministic output, since the audit may visit
	// resources in a non-deterministic (map) order.
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].RuleID != results[j].RuleID {
			return results[i].RuleID < results[j].RuleID
		}
		li := results[i].Locations[0].LogicalLocations[0].FullyQualifiedName
		lj := results[j].Locations[0].LogicalLocations[0].FullyQualifiedName
		if li != lj {
			return li < lj
		}
		return results[i].Message.Text < results[j].Message.Text
	})

	ruleList := make([]sarifRule, 0, len(rules))
	for _, rule := range rules {
		ruleList = append(ruleList, rule)
	}
	sort.Slice(ruleList, func(i, j int) bool { return ruleList[i].ID < ruleList[j].ID })

	report := sarifReport{
		Schema:  sarifSchema,
		Version: sarifVersion,
		Runs: []sarifRun{{
			Tool: sarifTool{Driver: sarifDriver{
				Name:           sarifToolName,
				InformationURI: sarifInformationURI,
				Rules:          ruleList,
			}},
			Results: results,
		}},
	}
	return json.MarshalIndent(report, "", "  ")
}
