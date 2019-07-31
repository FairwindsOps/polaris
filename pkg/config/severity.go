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

package config

// Severity represents the severity of action to take (Ignore, Warning, Error).
type Severity string

const (
	// SeverityIgnore ignores validation failures
	SeverityIgnore Severity = "ignore"

	// SeverityWarning warns on validation failures
	SeverityWarning Severity = "warning"

	// SeverityError errors on validation failures
	SeverityError Severity = "error"
)

// IsActionable returns true if the severity level is warning or error
func (severity *Severity) IsActionable() bool {
	return *severity == SeverityWarning || *severity == SeverityError
}
