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
	"testing"

	cfg "github.com/fairwindsops/polaris/pkg/config"
	"github.com/stretchr/testify/require"
)

func Test_validateChecks(t *testing.T) {
	configured := map[string]cfg.Severity{
		"hostIPCSet":      cfg.SeverityWarning,
		"hostPIDSet":      cfg.SeverityDanger,
		"runAsPrivileged": cfg.SeverityDanger,
		"tagNotSpecified": cfg.SeverityDanger,
	}

	require.NoError(t, validateChecks([]string{"hostIPCSet"}, configured))
	require.NoError(t, validateChecks([]string{"hostIPCSet", "runAsPrivileged"}, configured))

	err := validateChecks([]string{"hostIPCSet", "priviledgedContainer"}, configured)
	require.ErrorContains(t, err, `unknown check "priviledgedContainer"`)
	require.ErrorContains(t, err, "runAsPrivileged")

	require.ErrorContains(t, validateChecks([]string{"hostIPCSSet"}, configured), `unknown check "hostIPCSSet"`)
}
