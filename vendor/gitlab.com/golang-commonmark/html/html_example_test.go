// Copyright 2015 The Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package html

import (
	"bytes"
	"fmt"
)

func ExampleEscapeString() {
	fmt.Println(EscapeString(`<a href="https://www.google.com/search?q=something&ie=utf-8">Google Search</a>`))

	// Output:
	// &lt;a href=&quot;https://www.google.com/search?q=something&amp;ie=utf-8&quot;&gt;Google Search&lt;/a&gt;
}

func ExampleWriteEscapedString() {
	var buf bytes.Buffer
	WriteEscapedString(&buf, `<a href="https://www.google.com/search?q=something&ie=utf-8">Google Search</a>`)
	fmt.Println(buf.String())

	// Output:
	// &lt;a href=&quot;https://www.google.com/search?q=something&amp;ie=utf-8&quot;&gt;Google Search&lt;/a&gt;
}

func ExampleParseEntity() {
	fmt.Println(ParseEntity("&quot;"))
	fmt.Println(ParseEntity("&#x22;"))
	fmt.Println(ParseEntity("&#34;"))
	fmt.Println(ParseEntity("&#x00000022;"))
	fmt.Println(ParseEntity("&#00000034;"))
	fmt.Println(ParseEntity("&#x000000022;"))
	fmt.Println(ParseEntity("&#000000034;"))
	fmt.Println(ParseEntity("&#98765432;"))

	// Output:
	// " 6
	// " 6
	// " 5
	// " 12
	// " 11
	//  0
	//  0
	// � 11
}

func ExampleUnescapeString() {
	fmt.Println(UnescapeString("&nbsp; &amp; &copy; &AElig; &Dcaron; &frac34; &HilbertSpace; &DifferentialD; &ClockwiseContourIntegral; &#35; &#1234; &#992; &#98765432; &#X22; &#XD06; &#xcab; &nbsp &x; &#; &#x; &ThisIsWayTooLongToBeAnEntityIsntIt; &hi?; &copy &MadeUpEntity;"))

	// Output:
	//   & © Æ Ď ¾ ℋ ⅆ ∲ # Ӓ Ϡ � " ആ ಫ &nbsp &x; &#; &#x; &ThisIsWayTooLongToBeAnEntityIsntIt; &hi?; &copy &MadeUpEntity;
}
