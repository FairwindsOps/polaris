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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fairwindsops/polaris/pkg/config"
)

func TestGetSarifOutput(t *testing.T) {
	auditData := AuditData{
		Results: []Result{
			{
				Name:      "test-deployment",
				Namespace: "default",
				Kind:      "Deployment",
				Results: ResultSet{
					"deploymentMissingReplicas": ResultMessage{
						ID:       "deploymentMissingReplicas",
						Message:  "Should have multiple replicas",
						Success:  false,
						Severity: config.SeverityDanger,
						Category: "Reliability",
					},
					"passingCheck": ResultMessage{
						ID:       "passingCheck",
						Message:  "This passed",
						Success:  true,
						Severity: config.SeverityWarning,
						Category: "Reliability",
					},
				},
				PodResult: &PodResult{
					Name: "test-deployment",
					ContainerResults: []ContainerResult{
						{
							Name: "nginx",
							Results: ResultSet{
								"runAsNonRoot": ResultMessage{
									ID:       "runAsNonRoot",
									Message:  "Should not be allowed to run as root",
									Success:  false,
									Severity: config.SeverityWarning,
									Category: "Security",
								},
							},
						},
					},
				},
			},
		},
	}

	out, err := auditData.GetSarifOutput()
	assert.NoError(t, err)

	var report sarifReport
	assert.NoError(t, json.Unmarshal(out, &report))

	assert.Equal(t, "2.1.0", report.Version)
	assert.Equal(t, sarifSchema, report.Schema)
	assert.Len(t, report.Runs, 1)

	run := report.Runs[0]
	assert.Equal(t, "Polaris", run.Tool.Driver.Name)

	// Only the two failed checks are emitted; the passing check is skipped.
	assert.Len(t, run.Results, 2)
	assert.Len(t, run.Tool.Driver.Rules, 2)

	// Results are emitted in a deterministic order (ruleId, then location).
	assert.Equal(t, []string{"deploymentMissingReplicas", "runAsNonRoot"},
		[]string{run.Results[0].RuleID, run.Results[1].RuleID})

	byRule := map[string]sarifResult{}
	for _, r := range run.Results {
		byRule[r.RuleID] = r
	}

	danger, ok := byRule["deploymentMissingReplicas"]
	assert.True(t, ok)
	assert.Equal(t, "error", danger.Level)
	assert.Equal(t, "Should have multiple replicas", danger.Message.Text)
	assert.Equal(t, "default/Deployment/test-deployment", danger.Locations[0].LogicalLocations[0].FullyQualifiedName)

	warning, ok := byRule["runAsNonRoot"]
	assert.True(t, ok)
	assert.Equal(t, "warning", warning.Level)
	assert.Equal(t, "default/Deployment/test-deployment/container/nginx", warning.Locations[0].LogicalLocations[0].FullyQualifiedName)

	_, passing := byRule["passingCheck"]
	assert.False(t, passing)
}

func TestGetSarifOutputEmpty(t *testing.T) {
	out, err := AuditData{}.GetSarifOutput()
	assert.NoError(t, err)

	var report sarifReport
	assert.NoError(t, json.Unmarshal(out, &report))
	assert.Equal(t, "2.1.0", report.Version)
	assert.Len(t, report.Runs, 1)
	assert.Empty(t, report.Runs[0].Results)
	assert.Empty(t, report.Runs[0].Tool.Driver.Rules)
	// results must serialize as an empty array, not null.
	assert.Contains(t, string(out), `"results": []`)
	assert.NotContains(t, string(out), `"results": null`)
}

func TestSarifLevel(t *testing.T) {
	assert.Equal(t, "error", sarifLevel(config.SeverityDanger))
	assert.Equal(t, "warning", sarifLevel(config.SeverityWarning))
	assert.Equal(t, "note", sarifLevel(config.SeverityIgnore))
	assert.Equal(t, "none", sarifLevel(config.Severity("bogus")))
}
