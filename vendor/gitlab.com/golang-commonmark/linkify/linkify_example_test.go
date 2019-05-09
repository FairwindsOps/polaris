// Copyright 2015 The Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linkify

import "fmt"

func Example() {
	input := `
	Check out this link to http://google.com
You can also email support@example.com to view more.

Some more links: fsf.org http://www.gnu.org/licenses/gpl-3.0.en.html 127.0.0.1
                 localhost:80	github.com/trending?l=Go	//reddit.com/r/golang
mailto:r@golang.org some.nonexistent.host.name flibustahezeous3.onion
`
	for _, l := range Links(input) {
		fmt.Printf("Scheme: %-8s  URL: %s\n", l.Scheme, input[l.Start:l.End])
	}

	// Output:
	// Scheme: http:     URL: http://google.com
	// Scheme: mailto:   URL: support@example.com
	// Scheme:           URL: fsf.org
	// Scheme: http:     URL: http://www.gnu.org/licenses/gpl-3.0.en.html
	// Scheme:           URL: 127.0.0.1
	// Scheme:           URL: localhost:80
	// Scheme:           URL: github.com/trending?l=Go
	// Scheme: //        URL: //reddit.com/r/golang
	// Scheme: mailto:   URL: mailto:r@golang.org
	// Scheme:           URL: some.nonexistent.host.name
	// Scheme:           URL: flibustahezeous3.onion
}
