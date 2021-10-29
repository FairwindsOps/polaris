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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPrettyOutputResultSet(t *testing.T) {
	set := ResultSet{
		"hostPIDSet": {ID: "hostPIDSet", Message: "Host PID is not configured", Success: true, Severity: "danger", Category: "Security"},
		"hostIPCSet": {ID: "hostIPCSet", Message: "Host IPC is not configured", Success: true, Severity: "danger", Category: "Security"},
	}
	results := set.GetPrettyOutput()
	expectedResult := fmt.Sprintf(`    hostIPCSet                           %s Success
        Security - Host IPC is not configured
    hostPIDSet                           %s Success
        Security - Host PID is not configured
`, "\x8e\x89", "\x8e\x89")
	assert.Equal(t, expectedResult, results)
}
