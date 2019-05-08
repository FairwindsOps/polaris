// Copyright 2015 The Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package markdown

import "testing"

func TestMatchTagName(t *testing.T) {
	type testCase struct {
		in   string
		want string
	}
	testCases := []testCase{
		{"", ""},
		{"a", ""},
		{"/a>", "a"},
		{"a>", "a"},
		{"A>", "a"},
		{"a\n", "a"},
		{"br/>", "br"},
		{"em", ""},
		{"waytoolongtobeatag>", ""},
		{"waytoolongtobeatag", ""},
	}
	for _, tc := range testCases {
		got := matchTagName(tc.in)
		if got != tc.want {
			t.Errorf("matchTagName(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
