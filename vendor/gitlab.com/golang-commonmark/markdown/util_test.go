// Copyright 2015 The Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package markdown

import (
	"testing"
)

func TestUnescapeAll(t *testing.T) {
	type testCase struct {
		in   string
		want string
	}
	testCases := []testCase{
		{"", ""},
		{`foo \`, `foo \`},
		{`\f\o\o`, `\f\o\o`},
		{`\!\"\#\$\%\&\'\(\)\*\+\,\-\.\/\:\;\<\=\>\?\@\[\\\]\^\_\{\|\}\~`, "!\"#$%&'()*+,-./:;<=>?@[\\]^_{|}~"},
		{`\foo b\#a\r`, `\foo b#a\r`},
		{"&amp;", "&"},
		{`\&amp;`, "&amp;"},
		{"&quot;&amp;&quot;", `"&"`},
		{"&nLt;&nGt;", "\u226a\u20d2\u226b\u20d2"},
	}
	for _, tc := range testCases {
		got := unescapeAll(tc.in)
		if got != tc.want {
			t.Errorf("unescapeAll(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestNormalizeInlineCode(t *testing.T) {
	type testCase struct {
		in   string
		want string
	}
	testCases := []testCase{
		{"", ""},
		{"    ", ""},
		{"\n", ""},
		{"\n\n", ""},
		{"\nfoo\n", "foo"},
		{"foo bar", "foo bar"},
		{"foo\nbar", "foo bar"},
		{"foo\n\nbar", "foo bar"},
		{"foo  bar", "foo bar"},
		{"foo  \nbar", "foo bar"},
		{"  foo  ", "foo"},
		{"    foo   bar\n  baz\n  foo\n\n", "foo bar baz foo"},
	}
	for _, tc := range testCases {
		got := normalizeInlineCode(tc.in)
		if got != tc.want {
			t.Errorf("normalizeInlineCode(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestValidateLink(t *testing.T) {
	type testCase struct {
		in   string
		want bool
	}
	testCases := []testCase{
		{"", true},
		{"http://google.com", true},
		{"javascript:alert(1)", false},
		{"Javascript:alert(1)", false},
		{"  javascript:alert(1)", false},
		{"\x09javascript:alert(1)", false},
		{"\x00javascript:alert(1)", false},
		// {"&#9;javascript:alert(1)", false},
		{"&#1;javascript:alert(1)", true},
		// {"&#x6a;avascript:alert(1)", false},
		// {"&#x4a;avascript:alert(1)", false},
		{"&#x26;#x6a;avascript:alert(1)", true},
		// {"javascript&#x3a;alert(1)", false},
		{"vbscript:alert(1)", false},
		{"file:///home", false},
		{"data:text/html;base64,", false},
		{"data:image/jpeg;base64,", true},
		{"data:image/gif;base64,", true},
		{"data:image/png;base64,", true},
		{"data:image/webp;base64,", true},
	}
	for _, tc := range testCases {
		got := validateLink(tc.in)
		if got != tc.want {
			t.Errorf("validateLink(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestIsMarkdownPunct(t *testing.T) {
	type testCase struct {
		in   string
		want bool
	}
	testCases := []testCase{
		{"!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~", true},
		{" \x7f0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZабвгдеёжзийклмнопрстуфхцчшщьыъэюяАБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЬЫЪЭЮЯ", false},
	}
	for _, tc := range testCases {
		for _, r := range tc.in {
			got := isMdAsciiPunct(r)
			if got != tc.want {
				t.Errorf("isMarkdownPunct(%c) = %v, want %v", r, got, tc.want)
			}
		}
	}
}

func TestNormalizeLink(t *testing.T) {
	type testCase struct {
		in   string
		want string
	}
	testCases := []testCase{
		{"", ""},
		// {"http://google.com/s%65arch", "http://google.com/search"},
		{"http://google.com/search?query=%5B%5D", "http://google.com/search?query=%5B%5D"},
		{"//google.com", "//google.com"},
		{"http://%XX", "http://%25XX"},
	}
	for _, tc := range testCases {
		got := normalizeLink(tc.in)
		if got != tc.want {
			t.Errorf("normalizeLink(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestNormalizeLinkText(t *testing.T) {
	type testCase struct {
		in   string
		want string
	}
	testCases := []testCase{
		{"", ""},
		{"http://google.com/s%65arch", "http://google.com/search"},
		{"http://google.com/search%XX", "http://google.com/search%XX"},
		// {"http://google.com/search%80", "http://google.com/search%80"},
	}
	for _, tc := range testCases {
		got := normalizeLinkText(tc.in)
		if got != tc.want {
			t.Errorf("normalizeLinkText(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestNormalizeReference(t *testing.T) {
	type testCase struct {
		in   string
		want string
	}
	testCases := []testCase{
		{"", ""},
		{"  r  e  f  ", "r e f"},
		{"  ref  ", "ref"},
		{" ref ", "ref"},
		{"ref", "ref"},
		{"REF", "ref"},
		{"r\u00a0e\u00a0f", "r e f"},
	}
	for _, tc := range testCases {
		got := normalizeReference(tc.in)
		if got != tc.want {
			t.Errorf("normalizeReference(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestRemoveSpecial(t *testing.T) {
	type testCase struct {
		in   string
		want string
	}
	testCases := []testCase{
		{"", ""},
		{"javascript", "javascript"},
		{"javascript\x00", "javascript"},
		{"\x00javascript", "javascript"},
		{"\x00java\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x20script\x7f", "javascript"},
	}
	for _, tc := range testCases {
		got := removeSpecial(tc.in)
		if got != tc.want {
			t.Errorf("removeSpecial(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
