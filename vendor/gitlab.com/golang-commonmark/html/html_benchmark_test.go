// Copyright 2015 The Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package html

import (
	"html"
	"testing"
)

var escapeBenchmarkStrings = []string{
	"",
	"a b c d e f g h i j k l m n o p q r s t u v w x y z ",
	`<a href="http://google.com?q=&">google.com</a>`,
}

func BenchmarkEscapeString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, s := range escapeBenchmarkStrings {
			EscapeString(s)
		}
	}
}

func BenchmarkEscapeStringStdlib(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, s := range escapeBenchmarkStrings {
			html.EscapeString(s)
		}
	}
}
