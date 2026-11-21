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

package cmd

import (
	"context"
	"testing"

	cfg "github.com/fairwindsops/polaris/pkg/config"
	"github.com/fairwindsops/polaris/pkg/kube"
	"github.com/fairwindsops/polaris/pkg/validator"
	"github.com/fairwindsops/polaris/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateExitCodeFlags(t *testing.T) {
	tests := []struct {
		name          string
		onDangerSet   bool
		onSeveritySet bool
		severity      string
		wantErr       string
	}{
		{name: "no flags"},
		{name: "danger flag only", onDangerSet: true},
		{name: "severity warning", onSeveritySet: true, severity: "warning"},
		{name: "severity danger", onSeveritySet: true, severity: "danger"},
		{
			name:          "severity empty",
			onSeveritySet: true,
			severity:      "",
			wantErr:       `invalid value "" for --set-exit-code-on-severity, must be 'warning' or 'danger'`,
		},
		{
			name:          "severity unknown",
			onSeveritySet: true,
			severity:      "error",
			wantErr:       `invalid value "error" for --set-exit-code-on-severity, must be 'warning' or 'danger'`,
		},
		{
			name:          "severity is case sensitive",
			onSeveritySet: true,
			severity:      "Warning",
			wantErr:       `invalid value "Warning" for --set-exit-code-on-severity, must be 'warning' or 'danger'`,
		},
		{
			name:          "both flags",
			onDangerSet:   true,
			onSeveritySet: true,
			severity:      "danger",
			wantErr:       "--set-exit-code-on-danger and --set-exit-code-on-severity cannot be used together",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateExitCodeFlags(tc.onDangerSet, tc.onSeveritySet, tc.severity)
			if tc.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.EqualError(t, err, tc.wantErr)
		})
	}
}

func TestAuditExitCode(t *testing.T) {
	clean := validator.CountSummary{Successes: 10}
	warnings := validator.CountSummary{Successes: 10, Warnings: 2}
	dangers := validator.CountSummary{Successes: 10, Dangers: 1}
	mixed := validator.CountSummary{Successes: 10, Warnings: 2, Dangers: 1}
	// Scores: warnings 90, dangers 90, mixed 83, failingScore 0.
	failingScore := validator.CountSummary{Warnings: 5, Dangers: 5}

	tests := []struct {
		name     string
		summary  validator.CountSummary
		onDanger bool
		severity string
		minScore int
		want     int
	}{
		{name: "no options, issues found", summary: mixed, want: 0},
		{name: "clean audit with warning threshold", summary: clean, severity: "warning", want: 0},
		{name: "clean audit with danger threshold", summary: clean, severity: "danger", want: 0},
		{name: "warnings with warning threshold", summary: warnings, severity: "warning", want: 3},
		{name: "warnings with danger threshold", summary: warnings, severity: "danger", want: 0},
		{name: "dangers with warning threshold", summary: dangers, severity: "warning", want: 3},
		{name: "dangers with danger threshold", summary: dangers, severity: "danger", want: 3},
		{name: "mixed with warning threshold", summary: mixed, severity: "warning", want: 3},
		{name: "mixed with danger threshold", summary: mixed, severity: "danger", want: 3},
		{name: "danger flag with dangers", summary: dangers, onDanger: true, want: 3},
		{name: "danger flag with warnings only", summary: warnings, onDanger: true, want: 0},
		{name: "score above minimum", summary: warnings, minScore: 90, want: 0},
		{name: "score below minimum", summary: mixed, minScore: 90, want: 4},
		{name: "danger flag keeps its precedence over the score", summary: failingScore, onDanger: true, minScore: 90, want: 3},
		{name: "severity beats a failing score", summary: failingScore, severity: "warning", minScore: 90, want: 3},
		{name: "score fails when severity threshold passes", summary: warnings, severity: "danger", minScore: 95, want: 4},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			code, message := auditExitCode(tc.summary, tc.onDanger, tc.severity, tc.minScore)
			assert.Equal(t, tc.want, code)
			if tc.want == 0 {
				assert.Empty(t, message)
			} else {
				assert.NotEmpty(t, message)
			}
		})
	}
}

func TestAuditExitCodeMessages(t *testing.T) {
	_, message := auditExitCode(validator.CountSummary{Warnings: 2, Dangers: 1}, false, "warning", 0)
	assert.Equal(t, "1 danger and 2 warning items found in audit", message)

	_, message = auditExitCode(validator.CountSummary{Dangers: 3}, true, "", 0)
	assert.Equal(t, "3 danger items found in audit", message)

	_, message = auditExitCode(validator.CountSummary{Successes: 1, Warnings: 1}, false, "", 80)
	assert.Equal(t, "Audit score of 66 is less than the provided minimum of 80", message)
}

// The exit code is computed from the audit before output filtering, so hiding
// warnings from the report with --severity does not hide them from the threshold.
func TestAuditExitCodeIgnoresOutputFiltering(t *testing.T) {
	config := cfg.Configuration{
		Checks: map[string]cfg.Severity{
			"readinessProbeMissing": cfg.SeverityDanger,
			"livenessProbeMissing":  cfg.SeverityWarning,
		},
	}
	k8s, dynamicClient := test.SetupTestAPI(test.GetMockControllers("test")...)
	resources, err := kube.CreateResourceProviderFromAPI(context.Background(), k8s, "test", dynamicClient, config)
	require.NoError(t, err)
	auditData, err := validator.RunAudit(context.Background(), config, resources)
	require.NoError(t, err)

	summary := auditData.GetSummary()
	require.NotZero(t, summary.Warnings)
	require.NotZero(t, summary.Dangers)

	filtered := auditData.FilterResultsBySeverityLevel(cfg.SeverityDanger)
	assert.Zero(t, filtered.GetSummary().Warnings, "the report should hide warnings")
	assert.Equal(t, summary, auditData.GetSummary(), "filtering must not change the original audit")

	code, _ := auditExitCode(auditData.GetSummary(), false, "warning", 0)
	assert.Equal(t, exitCodeSeverityFailed, code)
}
