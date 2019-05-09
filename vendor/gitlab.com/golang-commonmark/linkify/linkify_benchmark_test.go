// Copyright 2015 The Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linkify

import "testing"

func BenchmarkLinks(b *testing.B) {
	input := `
	Check out this link to http://google.com
You can also email support@example.com to view more.

Some more links: fsf.org http://www.gnu.org/licenses/gpl-3.0.en.html 127.0.0.1
                 localhost:80	github.com/trending	//reddit.com/r/golang
mailto:r@golang.org some.nonexistent.host.name flibustahezeous3.onion
`

	for i := 0; i < b.N; i++ {
		Links(input)
	}
}
