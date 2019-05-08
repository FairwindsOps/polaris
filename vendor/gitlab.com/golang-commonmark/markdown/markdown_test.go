// Copyright 2015 The Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package markdown

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"gitlab.com/opennota/wd"
)

type example struct {
	Filename string
	Num      int `json:"example"`
	Markdown string
	HTML     string
	Section  string
}

func loadExamplesFromJSON(fn string) []example {
	f, err := os.Open(fn)
	if err != nil {
		panic(err)
	}

	var examples []example
	err = json.NewDecoder(f).Decode(&examples)
	if err != nil {
		panic(err)
	}

	return examples
}

func render(src string, options ...option) (_ string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	md := New(options...)
	return md.RenderToString([]byte(src)), nil
}

func TestCommonMark(t *testing.T) {
	examples := loadExamplesFromJSON("spec/commonmark-0.28.json")
	for _, ex := range examples {
		ex := ex
		t.Run(fmt.Sprintf("example #%d (%s)", ex.Num, ex.Section), func(t *testing.T) {
			result, err := render(ex.Markdown, HTML(true), XHTMLOutput(true), Linkify(false), Typographer(false), LangPrefix("language-"))
			if err != nil {
				t.Errorf("#%d (%s): PANIC (%v)", ex.Num, ex.Section, err)
			} else if result != ex.HTML {
				d := wd.ColouredDiff(ex.HTML, result, false)
				d = wd.NumberLines(d)
				t.Errorf("#%d (%s):\n%s", ex.Num, ex.Section, d)
			}
		})
	}
}

func TestRenderSpec(t *testing.T) {
	data, err := ioutil.ReadFile("spec/spec-0.28.txt")
	if err != nil {
		t.Fatal(err)
	}

	md := New(HTML(true), XHTMLOutput(true))
	md.RenderToString(data)
}

func TestFrenchQuoteMarks(t *testing.T) {
	md := New(Quotes([]string{"« ", " »", "‹ ", " ›"}))
	got := md.RenderToString([]byte(`"Son 'explication' n'est qu'un mensonge", s'indigna le député.`))
	want := "<p>« Son ‹ explication › n’est qu’un mensonge », s’indigna le député.</p>\n"
	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestMultibyteCharAtTheEnd(t *testing.T) {
	md := New()
	got := md.RenderToString([]byte("“Test”\nTest\n“Test”"))
	want := "<p>“Test”\nTest\n“Test”</p>\n"
	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}
