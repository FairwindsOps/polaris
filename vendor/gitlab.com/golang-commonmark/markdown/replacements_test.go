// Copyright 2015 The Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package markdown

import "testing"

func TestPerformReplacements(t *testing.T) {
	type testCase struct {
		in   string
		want string
	}
	testCases := []testCase{
		{"", ""},
		{"(c)", "©"},
		{"(r)", "®"},
		{"(tm)", "™"},
		{"(p)", "§"},
		{"(C)", "©"},
		{"(R)", "®"},
		{"(TM)", "™"},
		{"(P)", "§"},
		{"+-", "±"},
		{"..", "…"},
		{"...", "…"},
		{"....", "…"},
		{"!..", "!.."},
		{"?..", "?.."},
		{"!...", "!.."},
		{"?...", "?.."},
		{",,", ","},
		{",,,", ","},
		{"--", "–"},
		{"---", "—"},
		{"----", "----"},
		{"?!!!!!......", "?!!.."},
		{" (c) ", " © "},
		{" (r) ", " ® "},
		{" (tm) ", " ™ "},
		{" (p) ", " § "},
		{" +- ", " ± "},
		{" .. ", " … "},
		{" ... ", " … "},
		{" .... ", " … "},
		{" !.. ", " !.. "},
		{" ?.. ", " ?.. "},
		{" !... ", " !.. "},
		{" ?... ", " ?.. "},
		{" ,, ", " , "},
		{" ,,, ", " , "},
		{" -- ", " – "},
		{" --- ", " — "},
		{" ---- ", " ---- "},
		{"(", "("},
		{"(c", "(c"},
		{"(r", "(r"},
		{"(p", "(p"},
		{"(t", "(t"},
		{"(tm", "(tm"},
		{"(c]", "(c]"},
		{"(r]", "(r]"},
		{"(p]", "(p]"},
		{"(tm]", "(tm]"},
		{"+", "+"},
		{"++", "++"},
		{".", "."},
		{".2", ".2"},
		{",", ","},
		{", ", ", "},
		{"-", "-"},
		{"- ", "- "},
		{"(z)", "(z)"},
		{"(c)(r)(tm)(p)..?!!!!!.....,,-- --- -- ----", "©®™§…?!!..,– — – ----"},
	}
	for _, tc := range testCases {
		got := performReplacements(tc.in)
		if got != tc.want {
			t.Errorf("performReplacements(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
