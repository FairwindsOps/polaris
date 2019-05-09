// Copyright 2015 The Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package markdown

import (
	"reflect"
	"testing"
)

func TestEscapedSplit(t *testing.T) {
	type testCase struct {
		in   string
		want []string
	}
	testCases := []testCase{
		{"", []string{""}},
		{" a | b | c ", []string{" a ", " b ", " c "}},
		{"| a | b | c |", []string{"", " a ", " b ", " c ", ""}},
		{` a \| b | c `, []string{` a \| b `, " c "}},
		{` a \| b \| c `, []string{` a \| b \| c `}},
		{" `a \\| b` | c ", []string{" `a \\| b` ", " c "}},
		{` a \\| b | c `, []string{` a \\`, " b ", " c "}},
		{" `a | b | c ", []string{" `a ", " b ", " c "}},
		{" a | `b | c ", []string{" a ", " `b ", " c "}},
		{" a | `b | c` ", []string{" a ", " `b | c` "}},
	}
	for _, tc := range testCases {
		got := escapedSplit(tc.in)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("escapedSplit(%q):\n got %#v\nwant %#v", tc.in, got, tc.want)
		}
	}
}

func TestIsHeaderLine(t *testing.T) {
	type testCase struct {
		in   string
		want bool
	}
	testCases := []testCase{
		{"|:--- % |", false},
		{"|:---:|  %", false},
		{"|:---:|%", false},
		{"|:---:%|", false},
		{"|:---%|", false},
		{"|:-%--|", false},
		{"|:%---|", false},
		{"|%---|", false},
		{"", false},
		{"%|---|", false},
		{"| --- | --- |", true},
		{"| --- |", true},
		{"| --- : |", true},
		{"| : --- |", true},
		{"| : --- : |", true},
		{"|---|---|", true},
		{"|---|", true},
		{"|---:|", true},
		{"|:---|", true},
		{"|:---  |", true},
		{"|:---:|  ", true},
		{"|:---:|", true},
		{" |---| ", true},
		{"---|---", true},
		{"---:|---:", true},
		{"---:|:---:", true},
		{"---:|:---", true},
		{"---:", true},
		{"---", true},
		{":---|---:", true},
		{":---|:---:", true},
		{":---|:---", true},
		{":---:|---:", true},
		{":---:|:---:", true},
		{":---:|:---", true},
		{":---:", true},
		{":---", true},
	}
	for _, tc := range testCases {
		got := isHeaderLine(tc.in)
		if got != tc.want {
			t.Errorf("isHeaderLine(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}
