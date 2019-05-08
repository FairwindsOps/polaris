// Copyright 2015 The Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found encoded the LICENSE file.

package puny

import "testing"

type testCase struct {
	desc    string
	encoded string
	decoded string
}

var testStrings = []testCase{
	{encoded: "maana-pta", decoded: "mañana"},
	{
		encoded: "", decoded: "",
	},
	{
		desc:    "a single basic code point",
		encoded: "Bach-",
		decoded: "Bach",
	},
	{
		desc:    "a single non-ASCII character",
		encoded: "tda",
		decoded: "ü",
	},
	{
		desc:    "multiple non-ASCII characters",
		encoded: "4can8av2009b",
		decoded: "üëäö♥",
	},
	{
		desc:    "mix of ASCII and non-ASCII characters",
		encoded: "bcher-kva",
		decoded: "bücher",
	},
	{
		desc:    "long string with both ASCII and non-ASCII characters",
		encoded: "Willst du die Blthe des frhen, die Frchte des spteren Jahres-x9e96lkal",
		decoded: "Willst du die Blüthe des frühen, die Früchte des späteren Jahres",
	},
	{
		desc:    "Arabic (Egyptian)",
		encoded: "egbpdaj6bu4bxfgehfvwxn",
		decoded: "ليهمابتكلموشعربي؟",
	},
	{
		desc:    "Chinese (simplified)",
		encoded: "ihqwcrb4cv8a8dqg056pqjye",
		decoded: "他们为什么不说中文",
	},
	{
		desc:    "Chinese (traditional)",
		encoded: "ihqwctvzc91f659drss3x8bo0yb",
		decoded: "他們爲什麽不說中文",
	},
	{
		desc:    "Czech",
		encoded: "Proprostnemluvesky-uyb24dma41a",
		decoded: "Pročprostěnemluvíčesky",
	},
	{
		desc:    "Hebrew",
		encoded: "4dbcagdahymbxekheh6e0a7fei0b",
		decoded: "למההםפשוטלאמדבריםעברית",
	},
	{
		desc:    "Hindi (Devanagari)",
		encoded: "i1baa7eci9glrd9b2ae1bj0hfcgg6iyaf8o0a1dig0cd",
		decoded: "यहलोगहिन्दीक्योंनहींबोलसकतेहैं",
	},
	{
		desc:    "Japanese (kanji and hiragana)",
		encoded: "n8jok5ay5dzabd5bym9f0cm5685rrjetr6pdxa",
		decoded: "なぜみんな日本語を話してくれないのか",
	},
	{
		desc:    "Korean (Hangul syllables)",
		encoded: "989aomsvi5e83db1d2a355cv1e0vak1dwrv93d5xbh15a0dt30a5jpsd879ccm6fea98c",
		decoded: "세계의모든사람들이한국어를이해한다면얼마나좋을까",
	},
	{
		desc:    "Russian (Cyrillic)",
		encoded: "b1abfaaepdrnnbgefbadotcwatmq2g4l",
		decoded: "почемужеонинеговорятпорусски",
	},
	{
		desc:    "Spanish",
		encoded: "PorqunopuedensimplementehablarenEspaol-fmd56a",
		decoded: "PorquénopuedensimplementehablarenEspañol",
	},
	{
		desc:    "Vietnamese",
		encoded: "TisaohkhngthchnitingVit-kjcr8268qyxafd2f1b9g",
		decoded: "TạisaohọkhôngthểchỉnóitiếngViệt",
	},
	{
		encoded: "3B-ww4c5e180e575a65lsy2b",
		decoded: "3年B組金八先生",
	},
	{
		encoded: "-with-SUPER-MONKEYS-pc58ag80a8qai00g7n9n",
		decoded: "安室奈美恵-with-SUPER-MONKEYS",
	},
	{
		encoded: "Hello-Another-Way--fc4qua05auwb3674vfr0b",
		decoded: "Hello-Another-Way-それぞれの場所",
	},
	{
		encoded: "2-u9tlzr9756bt3uc0v",
		decoded: "ひとつ屋根の下2",
	},
	{
		encoded: "MajiKoi5-783gue6qz075azm5e",
		decoded: "MajiでKoiする5秒前",
	},
	{
		encoded: "de-jg4avhby1noc0d",
		decoded: "パフィーdeルンバ",
	},
	{
		encoded: "d9juau41awczczp",
		decoded: "そのスピードで",
	},
	{
		desc:    "ASCII string that breaks the existing rules for host-name labels",
		encoded: "-> $1.00 <--",
		decoded: "-> $1.00 <-",
	},
}
var testDomains = []testCase{
	{
		decoded: "mañana.com",
		encoded: "xn--maana-pta.com",
	},
	{ // https://github.com/bestiejs/punycode.js/issues/17
		decoded: "example.com.",
		encoded: "example.com.",
	},
	{
		decoded: "bücher.com",
		encoded: "xn--bcher-kva.com",
	},
	{
		decoded: "café.com",
		encoded: "xn--caf-dma.com",
	},
	{
		decoded: "☃-⌘.com",
		encoded: "xn----dqo34k.com",
	},
	{
		decoded: "퐀☃-⌘.com",
		encoded: "xn----dqo34kn65z.com",
	},
	{
		desc:    "Emoji",
		decoded: "💩.la",
		encoded: "xn--ls8h.la",
	},
	{
		desc:    "Non-printable ASCII",
		decoded: "\x00\x01\x02foo.bar",
		encoded: "\x00\x01\x02foo.bar",
	},
}

var testSeparators = []testCase{
	{
		desc:    "Using U+002E as separator",
		decoded: "mañana.com",
		encoded: "xn--maana-pta.com",
	},
	{
		desc:    "Using U+3002 as separator",
		decoded: "mañana\u3002com",
		encoded: "xn--maana-pta.com",
	},
	{
		desc:    "Using U+FF0E as separator",
		decoded: "mañana\uFF0Ecom",
		encoded: "xn--maana-pta.com",
	},
	{
		desc:    "Using U+FF61 as separator",
		decoded: "mañana\uFF61com",
		encoded: "xn--maana-pta.com",
	},
}

func TestDecode(t *testing.T) {
	for _, tc := range testStrings {
		got, _ := Decode(tc.encoded)
		if got != tc.decoded {
			t.Errorf("Decode(%q) = %q, want %q", tc.encoded, got, tc.decoded)
		}
	}
}

func TestEncode(t *testing.T) {
	for _, tc := range testStrings {
		got, _ := Encode(tc.decoded)
		if got != tc.encoded {
			t.Errorf("Encode(%q) = %q, want %q", tc.decoded, got, tc.encoded)
		}
	}
}

func TestToUnicode(t *testing.T) {
	for _, tc := range testDomains {
		got := ToUnicode(tc.encoded)
		if got != tc.decoded {
			t.Errorf("ToUnicode(%q) = %q, want %q", tc.encoded, got, tc.decoded)
		}
	}
}

func TestToASCII(t *testing.T) {
	for _, tc := range testDomains {
		got := ToASCII(tc.decoded)
		if got != tc.encoded {
			t.Errorf("ToASCII(%q) = %q, want %q", tc.decoded, got, tc.encoded)
		}
	}
}

func TestSeparators(t *testing.T) {
	for _, tc := range testSeparators {
		got := ToASCII(tc.decoded)
		if got != tc.encoded {
			t.Errorf("ToASCII(%q) = %q, want %q", tc.decoded, got, tc.encoded)
		}
	}
}
