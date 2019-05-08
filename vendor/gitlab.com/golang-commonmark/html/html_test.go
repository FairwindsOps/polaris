// Copyright 2015 The Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package html

import (
	"testing"
)

var parseEntityTests = []struct {
	in     string
	length int
	str    string
}{
	{in: "&;"},
	{in: "&"},
	{in: "&@;"},
	{in: "&@"},
	{in: "&#;"},
	{in: "&#"},
	{in: "&##;"},
	{in: "&##"},
	{in: "&#0"},
	{in: "&#09"},
	{in: "&#0a;"},
	{in: "&#0a"},
	{in: "&#34;", length: 5, str: `"`},
	{in: "&#98765432;", length: 11, str: BadEntity},
	{in: "&#9;", length: 4, str: "\t"},
	{in: "&ffffffffffffffffuuuuuuuuuuuuuuuu;"},
	{in: "&ffffffffffffffffuuuuuuuuuuuuuuuu"},
	{in: "&q;"},
	{in: "&q"},
	{in: "&q#;"},
	{in: "&q#"},
	{in: "&Q"},
	{in: "&q0;"},
	{in: "&q0"},
	{in: "&qu;"},
	{in: "&qu"},
	{in: "&quot;", length: 6, str: `"`},
	{in: "&#x;"},
	{in: "&#x"},
	{in: "&#X;"},
	{in: "&#X"},
	{in: "&#x0"},
	{in: "&#x0@;"},
	{in: "&#x0@"},
	{in: "&#X0"},
	{in: "&#x000000001;"},
	{in: "&#x000000001"},
	{in: "&#x09"},
	{in: "&#x09;", length: 6, str: "\t"},
	{in: "&#X09;", length: 6, str: "\t"},
	{in: "&#x0a;", length: 6, str: "\n"},
	{in: "&#x0A;", length: 6, str: "\n"},
	{in: "&#X0a;", length: 6, str: "\n"},
	{in: "&#X0A;", length: 6, str: "\n"},
	{in: "&#x22;", length: 6, str: `"`},
	{in: "&#x7fffffff;", length: 12, str: BadEntity},
	{in: "&#xa"},
	{in: "&#xA"},
	{in: "&#Xa"},
	{in: "&#XA"},
	{in: "&#xffffffff"},
	{in: "&#xfffffffff"},
	{in: "&#xffffffff;", length: 12, str: BadEntity},
	{in: "&#xG;"},
	{in: "&#xG"},
	{in: "&#XG;"},
	{in: "&#XG"},
}

var replaceEntitiesTests = []struct {
	in  string
	out string // empty means == in
}{
	{"", ""},
	{"a b c d e f g h i j k l m n o p q r s t u v w x y z", ""},
	{"&", ""},
	{"&nbsp; &amp; &copy; &quot;", `  & © "`},
	{"&#x22; &#X22; &#34;", `" " "`},
	{"&x; &#; &#x; &ThisIsWayTooLongToBeAnEntityIsntIt; &hi?; &copy &MadeUpEntity;", ""},
}

func TestParseEntity(t *testing.T) {
	for _, tc := range parseEntityTests {
		str, length := ParseEntity(tc.in)
		if str != tc.str || length != tc.length {
			t.Errorf("ParseEntity(%q): want %d,%q, got %d,%q", tc.in,
				tc.length, tc.str, length, str)
		}
	}
}

func TestUnescapeString(t *testing.T) {
	for _, tc := range replaceEntitiesTests {
		got := UnescapeString(tc.in)
		want := tc.out
		if want == "" {
			want = tc.in
		}
		if got != want {
			t.Errorf("UnescapeString(%q): want %q, got %q", tc.in, want, got)
		}
	}
}
