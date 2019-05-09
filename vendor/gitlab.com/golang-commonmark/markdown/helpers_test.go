// Copyright 2015 The Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package markdown

import "testing"

func TestParseLinkTitle(t *testing.T) {
	type testCase struct {
		in             string
		title          string
		nlines, endpos int
		ok             bool
	}
	testCases := []testCase{
		{"(multiline\ntitle)", "multiline\ntitle", 1, 17, true},
		{`"\"title\""`, `"title"`, 0, 11, true},
		{`"title"`, "title", 0, 7, true},
		{`"(title)"`, `(title)`, 0, 9, true},
		{`("title")`, `"title"`, 0, 9, true},
		{"", "", 0, 0, false},
		{`"`, "", 0, 0, false},
		{"(", "", 0, 0, false},
		{`(\`, "", 0, 0, false},
		{"(\\\n", "", 0, 0, false},
		{`"\"title\"`, "", 0, 0, false},
		{`"title`, "", 0, 0, false},
		{`("title"`, "", 0, 0, false},
		{"x", "", 0, 0, false},
	}
	for _, tc := range testCases {
		title, nlines, endpos, ok := parseLinkTitle(tc.in, 0, len(tc.in))
		if title != tc.title || nlines != tc.nlines || endpos != tc.endpos || ok != tc.ok {
			t.Errorf("parseLinkTitle(%q) = (%q, %d, %d, %v), want (%q, %d, %d, %v)", tc.in, title, nlines, endpos, ok, tc.title, tc.nlines, tc.endpos, tc.ok)
		}
	}
}

func TestParseLinkDestination(t *testing.T) {
	type testCase struct {
		in     string
		title  string
		endpos int
		ok     bool
	}
	testCases := []testCase{
		{"<\\>>", ">", 4, true},
		{"http://goo gle.com", "http://goo", 10, true},
		{"http://google.com", "http://google.com", 17, true},
		{"http://google.com/search?query=(1)", "http://google.com/search?query=(1)", 34, true},
		{"http://google.com/search?query=)1(", "http://google.com/search?query=", 31, true},
		{"http://google.com/search?query=((1))", "http://google.com/search?query=((1))", 36, true},
		{`http://google.com/search?query=a\ b\ c`, `http://google.com/search?query=a\ b\ c`, 38, true},
		{"http://goo\x00gle.com", "http://goo", 10, true},
		{"<link>", "link", 6, true},
		{"<", "", 0, false},
		{"<\\>", "", 0, false},
		{"<\\", "", 0, false},
		{"", "", 0, false},
		{"<>", "", 2, true},
		{"<link", "", 0, false},
		{"<\n", "", 0, false},
	}
	for _, tc := range testCases {
		title, _, endpos, ok := parseLinkDestination(tc.in, 0, len(tc.in))
		if title != tc.title || endpos != tc.endpos || ok != tc.ok {
			t.Errorf("parseLinkDestination(%q) = (%q, %d, %v), want (%q, %d, %v)", tc.in, title, endpos, ok, tc.title, tc.endpos, tc.ok)
		}
	}
}
