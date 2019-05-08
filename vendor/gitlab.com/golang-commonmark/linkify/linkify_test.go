// Copyright 2015 The Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linkify

import (
	"reflect"
	"testing"
)

func TestLinks(t *testing.T) {
	type testCase struct {
		in   string
		want []Link
	}
	testCases := []testCase{
		{"", nil},

		// 1

		{"127.0.0.1:8080/path?query=xxx#fragment", []Link{{Scheme: "", Start: 0, End: 38}}},
		{"127.0.0.1", []Link{{Scheme: "", Start: 0, End: 9}}},
		{">127.0.0.1<", []Link{{Scheme: "", Start: 1, End: 10}}},
		{"xxx 127.0.0.1", []Link{{Scheme: "", Start: 4, End: 13}}},
		{"1.0001.0.1", nil},
		{"1234.0.0", nil},
		{"1.2345.0.1", nil},
		{"1.2.3.4x", nil},
		{"_127.0.0.1", nil},
		{"127.0.0_1", nil},
		{"127.0.0.1/%", nil},
		{"127.0.0.1%", nil},
		{"127.0.0._", nil},
		{"127.0.0", nil},
		{"127.0._.1", nil},
		{" _.127.0", nil},
		{" .127._", nil},
		{" .127.", nil},
		{" .127.X", nil},
		{":234.0.0", nil},
		{"256.0.0.1", nil},
		{" X.127.0", nil},
		{"x234.0.0", nil},

		// 2

		{"google.com!", []Link{{Scheme: "", Start: 0, End: 10}}},
		{"google.com...", []Link{{Scheme: "", Start: 0, End: 10}}},
		{"google.com.", []Link{{Scheme: "", Start: 0, End: 10}}},
		{">google.com<", []Link{{Scheme: "", Start: 1, End: 11}}},
		{"google.com/?search", []Link{{Scheme: "", Start: 0, End: 18}}},
		{"some.host.name.com.net.abracadabra", nil},
		{"some.host.name.com.net.org", []Link{{Scheme: "", Start: 0, End: 26}}},
		{"some.host.name.com.net.org/", []Link{{Scheme: "", Start: 0, End: 27}}},
		{"some.host.name.com.net.org/path?query=xxx#fragment", []Link{{Scheme: "", Start: 0, End: 50}}},
		{"ya.ru", []Link{{Scheme: "", Start: 0, End: 5}}},
		{"a.r", nil},
		{".com", nil},
		{".COM", nil},
		{"-google.com", nil},
		{"goo---gle.com", nil},
		{"google-.com", nil},
		{"google.com/%", nil},
		{"google.comXXX", nil},
		{"GOOGLE.COMXXX", nil},
		{"images-.google.com", nil},
		{"\ufffdgoogle.com", nil},
		{"\x00.com", nil},
		{"\x00.COM", nil},
		{"xxx.com.abracadabra", nil},
		{"XXX.ZZ", nil},

		// 3

		{"a@a.ru", []Link{{Scheme: "mailto:", Start: 0, End: 6}}},
		{"a.b@golang.org", []Link{{Scheme: "mailto:", Start: 0, End: 14}}},
		{"a..b@golang.org", []Link{{Scheme: "", Start: 5, End: 15}}},
		{"r@golang.org", []Link{{Scheme: "mailto:", Start: 0, End: 12}}},
		{".r@golang.org", []Link{{Scheme: "", Start: 3, End: 13}}},
		{"@r@golang.org", []Link{{Scheme: "", Start: 3, End: 13}}},
		{"r.@golang.org", []Link{{Scheme: "", Start: 3, End: 13}}},
		{"\ufffdr@golang.org", []Link{{Scheme: "", Start: 5, End: 15}}},
		{"XXX r@golang.org", []Link{{Scheme: "mailto:", Start: 4, End: 16}}},
		{"XXX .r@golang.org", []Link{{Scheme: "", Start: 7, End: 17}}},
		{"a@a.r", nil},
		{"A@A.R", nil},
		{"@golang", nil},
		{"@GOOGLE", nil},
		{"r@golang.", nil},
		{"r@golang", nil},
		{"r@golang.zzz", nil},
		{"r@", nil},
		{"ROOT@", nil},
		{"r@\x00golang", nil},
		{"r\x00@golang", nil},
		{"R@\x00GOLANG", nil},
		{"R\x00@GOLANG", nil},

		// 4

		{"//127.0.0.1:80/", []Link{{Scheme: "//", Start: 0, End: 15}}},
		{"//google.com.", []Link{{Scheme: "//", Start: 0, End: 12}}},
		{"//ya.ru", []Link{{Scheme: "//", Start: 0, End: 7}}},
		{"://google.com", nil},
		{"//google.com%", nil},
		{"//google", nil},
		{"//google.zzz", nil},
		{"//\x00google", nil},
		{"x//google.com", nil},

		// 5

		{"mailto:a.b.c@golang.org", []Link{{Scheme: "mailto:", Start: 0, End: 23}}},
		{"mailto:a..b.c@golang.org", []Link{{Scheme: "", Start: 14, End: 24}}},
		{"mailto:r@golang.org", []Link{{Scheme: "mailto:", Start: 0, End: 19}}},
		{"mailto:r.@golang.org", []Link{{Scheme: "", Start: 10, End: 20}}},
		{"mailto:xxxЖ@golang.org", []Link{{Scheme: "", Start: 13, End: 23}}},
		{"ailto:xxxx", nil},
		{"AILTO:XXXX", nil},
		{"mailto:a@a.a", nil},
		{"MAILTO:A@A.A", nil},
		{"mailto:a..b.c@golang", nil},
		{"mailto:a..b.cgolang.org", nil},
		{"mailto:r@golang", nil},
		{"mailto:rgolangorg", nil},
		{"mailto:r@golang.zzz", nil},
		{"mailto:r@gol", nil},
		{"mailto:r@", nil},
		{"mailto:r@xxx", nil},
		{"mailto:\x00myemail", nil},
		{"MAILTO:\x00MYEMAIL", nil},
		{"mailto:xxx@", nil},
		{"xmailto:myemail", nil},
		{"XMAILTO:MYEMAIL", nil},
		{"xxxxxo:myemail", nil},
		{"XXXXXO:MYEMAIL", nil},
		{"xailto:myemail", nil},

		// 6

		{"http://127.0.0.1:80/", []Link{{Scheme: "http:", Start: 0, End: 20}}},
		{"http://google.com/[1(2", []Link{{Scheme: "http:", Start: 0, End: 18}}},
		{"http://google.com/#[1(2", []Link{{Scheme: "http:", Start: 0, End: 19}}},
		{"http://google.com/[1](2)", []Link{{Scheme: "http:", Start: 0, End: 24}}},
		{"http://google.com/#[1](2)", []Link{{Scheme: "http:", Start: 0, End: 25}}},
		{"http://google.com/[1)", []Link{{Scheme: "http:", Start: 0, End: 18}}},
		{"http://google.com/#[1)", []Link{{Scheme: "http:", Start: 0, End: 19}}},
		{"http://google.com ", []Link{{Scheme: "http:", Start: 0, End: 17}}},
		{"http://google.com--", []Link{{Scheme: "http:", Start: 0, End: 17}}},
		{"http://google.com-", []Link{{Scheme: "http:", Start: 0, End: 17}}},
		{"http://google.com!", []Link{{Scheme: "http:", Start: 0, End: 17}}},
		{"http://google.com...", []Link{{Scheme: "http:", Start: 0, End: 17}}},
		{"http://google.com.", []Link{{Scheme: "http:", Start: 0, End: 17}}},
		{"HtTp://gOoGlE.CoM", []Link{{Scheme: "http:", Start: 0, End: 17}}},
		{"http://google.com/--", []Link{{Scheme: "http:", Start: 0, End: 18}}},
		{"http://google.com/", []Link{{Scheme: "http:", Start: 0, End: 18}}},
		{`"http://google.com"`, []Link{{Scheme: "http:", Start: 1, End: 18}}},
		{">http://google.com<", []Link{{Scheme: "http:", Start: 1, End: 18}}},
		{"(http://google.com)", []Link{{Scheme: "http:", Start: 1, End: 18}}},
		{"http://google.com\n", []Link{{Scheme: "http:", Start: 0, End: 17}}},
		{"http://google.com/?query=[1(2", []Link{{Scheme: "http:", Start: 0, End: 25}}},
		{"http://google.com/?query=[1](2)", []Link{{Scheme: "http:", Start: 0, End: 31}}},
		{"http://google.com/?query=[1)", []Link{{Scheme: "http:", Start: 0, End: 25}}},
		{"http://google.com/?query=)x(", []Link{{Scheme: "http:", Start: 0, End: 25}}},
		{"http://google.com/?query=]xxx", []Link{{Scheme: "http:", Start: 0, End: 25}}},
		{"http://google.com/search---", []Link{{Scheme: "http:", Start: 0, End: 24}}},
		{"http://google.com/search?query=xxx---", []Link{{Scheme: "http:", Start: 0, End: 34}}},
		{"http://google.com/#toc---", []Link{{Scheme: "http:", Start: 0, End: 22}}},
		{"http://google.com/]xxx", []Link{{Scheme: "http:", Start: 0, End: 18}}},
		{"http://google.com/#]xxx", []Link{{Scheme: "http:", Start: 0, End: 19}}},
		{"http://www.youtube.com/watch?v=EIBRdBVkDHQ.", []Link{{Scheme: "http:", Start: 0, End: 42}}},
		{"http://goo---gle.com", nil},
		{"http://google-.com", nil},
		{"http://google.com\\", nil},
		{"http:google.com", nil},
		{"http://google.com/search?q%0z", nil},
		{"http://google.com/search?q%1", nil},
		{"http://google.com/search?q%", nil},
		{"http://google.com/search?q\ufffd", nil},
		{"http://google.com/search?q=x%78x?/:@#toc\ufffd", nil},
		{"http://google.com/search?q%z0", nil},
		{"http://google", nil},
		{"http://goo\ufffd", nil},
		{"http://images.google.com.-net", nil},
		{"http://images.google.com.net.-org", nil},
		{"http://\x00google", nil},
		{"HTTP://\x00GOOGLE", nil},
		{"http://xxx.com.abracadabra", nil},
		{"http://xxx", nil},
		{"HTTP://XXX", nil},
		{"http:xxxxxx", nil},
		{"HTTP:XXXXXX", nil},
		{"TP:GOOGLE", nil},
		{"ttp://google", nil},
		{"TTP://GOOGLE", nil},
		{"xhttp://google", nil},
		{"XHTTP://GOOGLE", nil},
		{"xttp://google", nil},
		{"XTTP://GOOGLE", nil},

		// 7

		{"https://google.com/p%61th?%71uery#%66ragment", []Link{{Scheme: "https:", Start: 0, End: 44}}},
		{"https://\x00google", nil},
		{"HTTPS://\x00GOOGLE", nil},
		{"https:XXgoogle", nil},
		{"HTTPS:XXGOOGLE", nil},
		{"https://xxx", nil},
		{"HTTPS://XXX", nil},
		{"ttps://google", nil},
		{"TTPS://GOOGLE", nil},
		{"xhttps://google", nil},
		{"XHTTPS://GOOGLE", nil},
		{"xttps://google", nil},
		{"XTTPS://GOOGLE", nil},

		// 8

		{"ftp://google.com", []Link{{Scheme: "ftp:", Start: 0, End: 16}}},
		{"fXp://google", nil},
		{"FXP://GOOGLE", nil},
		{"tp:google", nil},
		{"xtp://google", nil},
		{"XTP://GOOGLE", nil},

		// 9

		{">http://example.com<", []Link{{Scheme: "http:", Start: 1, End: 19}}},
		{"http://localhost:80/", []Link{{Scheme: "http:", Start: 0, End: 20}}},
		{">http://localhost:80<", []Link{{Scheme: "http:", Start: 1, End: 20}}},
		{"localhost:3128/path?query=xxx#fragment", []Link{{Scheme: "", Start: 0, End: 38}}},
		{"//localhost:80/", []Link{{Scheme: "//", Start: 0, End: 15}}},
		{"localhost:80/!$&'()*+,;=/:@", []Link{{Scheme: "", Start: 0, End: 27}}},
		{"localhost:80/#?!$&'()*+,;=/:@", []Link{{Scheme: "", Start: 0, End: 29}}},
		{"localhos", nil},
		{"localhost:0", nil},
		{"localhost:65536", nil},
		{"localhost:80/#fragment%0", nil},
		{"localhost:80/#fragment%0z", nil},
		{"localhost:80/#fragment%", nil},
		{"localhost:80/#fragment%z0", nil},
		{"LOCALHOST:80", nil},
		{"localhost:80/path%0", nil},
		{"localhost:80/path%0z", nil},
		{"localhost:80/path%", nil},
		{"localhost:80/path%z0", nil},
		{"localhost:80/\ufffd", nil},
		{"localhost:80\ufffd", nil},
		{"localhost:", nil},
		{"localhost:x", nil},
		{"localhost", nil},
		{"localhostX:80", nil},
		{"Xlocalhost:80", nil},

		// 10

		{":00", nil},
		{"skype://nickname", nil},
		{"SKYPE://NICKNAME", nil},
		{"XXX http://google.com r@golang.org localhost:80 mailto:r@golang.org //google.com XXX", []Link{{Scheme: "http:", Start: 4, End: 21}, {Scheme: "mailto:", Start: 22, End: 34}, {Scheme: "", Start: 35, End: 47}, {Scheme: "mailto:", Start: 48, End: 67}, {Scheme: "//", Start: 68, End: 80}}},

		// from twitter-text

		{"badly formatted http://foo!bar.com", []Link{{Start: 27, End: 34}}},
		{"badly formatted http://foo_bar.com", nil},
		{"Check out http://example.com/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", []Link{{Scheme: "http:", Start: 10, End: 1529}}},
		{"Check out http://search.twitter.com/#!/search?q=avro&lang=en", []Link{{Scheme: "http:", Start: 10, End: 60}}},
		{"Czech out sweet deals at http://mrs.domain-dash.biz ok?", []Link{{Scheme: "http:", Start: 25, End: 51}}},
		{"Email support@example.com", []Link{{Scheme: "mailto:", Start: 6, End: 25}}},
		{"example: https://twitter.com/otm_m@\"onmousedown=\"alert('foo')\" style=background-color:yellow;color:yellow;\"/", []Link{{Scheme: "https:", Start: 9, End: 35}}},
		{"foo@bar.com", []Link{{Scheme: "mailto:", Start: 0, End: 11}}},
		{"Go to http://example.com/a+ or http://example.com/a-", []Link{{Scheme: "http:", Start: 6, End: 27}, {Scheme: "http:", Start: 31, End: 52}}},
		{"Go to http://example.com/a+?this=that or http://example.com/a-?this=that", []Link{{Scheme: "http:", Start: 6, End: 37}, {Scheme: "http:", Start: 41, End: 72}}},
		{"Go to http://example.com/view/slug-url-?foo=bar", []Link{{Scheme: "http:", Start: 6, End: 47}}},
		{"http://example.com https://sslexample.com http://sub.example.com", []Link{{Scheme: "http:", Start: 0, End: 18}, {Scheme: "https:", Start: 19, End: 41}, {Scheme: "http:", Start: 42, End: 64}}},
		{"http://example.mobi/path", []Link{{Scheme: "http:", Start: 0, End: 24}}},
		{"http://foo.com AND https://bar.com AND www.foobar.com", []Link{{Scheme: "http:", Start: 0, End: 14}, {Scheme: "https:", Start: 19, End: 34}, {Scheme: "", Start: 39, End: 53}}},
		{"http://foo.com/?#foo", []Link{{Scheme: "http:", Start: 0, End: 20}}},
		{"http://www.example.com/#answer", []Link{{Scheme: "http:", Start: 0, End: 30}}},
		{"http://www.flickr.com/photos/29674651@N00/4382024406 if you know what's good for you.", []Link{{Scheme: "http:", Start: 0, End: 52}}},
		{"http://www.flickr.com/photos/29674651@N00/4382024406", []Link{{Scheme: "http:", Start: 0, End: 52}}},
		{"http://www.flickr.com/photos/29674651@N00/foobar", []Link{{Scheme: "http:", Start: 0, End: 48}}},
		{"In http://example.com/test, Douglas explains 42.", []Link{{Scheme: "http:", Start: 3, End: 26}}},
		{"Is http://www.foo-bar.com a valid URL?", []Link{{Scheme: "http:", Start: 3, End: 25}}},
		{"Is www...foo a valid URL?", nil},
		{"Is www.foo-bar.com a valid URL?", []Link{{Scheme: "", Start: 3, End: 18}}},
		{"Is www.-foo.com a valid URL?", nil},
		{"I think it's proper to end sentences with a period http://tell.me.com. Even when they contain a URL.", []Link{{Scheme: "http:", Start: 51, End: 69}}},
		{"I think it's proper to end sentences with a period http://tell.me/why?=because.i.want.it. Even when they contain a URL.", []Link{{Scheme: "http:", Start: 51, End: 88}}},
		{"I think it's proper to end sentences with a period http://tell.me/why. Even when they contain a URL.", []Link{{Scheme: "http:", Start: 51, End: 69}}},
		{"<link rel='true'>http://example.com</link>", []Link{{Scheme: "http:", Start: 17, End: 35}}},
		{"Parenthetically bad http://example.com/i_has_a_) thing", []Link{{Scheme: "http:", Start: 20, End: 47}}},
		{"See: http://example.com/café", []Link{{Scheme: "http:", Start: 5, End: 29}}},
		{"See: http://example.com/@user", []Link{{Scheme: "http:", Start: 5, End: 29}}},
		{"See: http://example.com/@user/", []Link{{Scheme: "http:", Start: 5, End: 30}}},
		{"See: http://example.com/?@user=@user", []Link{{Scheme: "http:", Start: 5, End: 36}}},
		{"See: http://x.xx.com/@\"style=\"color:pink\"onmouseover=alert(1)//", []Link{{Scheme: "http:", Start: 5, End: 22}}},
		{"test http://www.example.co.jp", []Link{{Scheme: "http:", Start: 5, End: 29}}},
		{"test http://www.example.sx", []Link{{Scheme: "http:", Start: 5, End: 26}}},
		{"text http://example.com,", []Link{{Scheme: "http:", Start: 5, End: 23}}},
		{"text http://example.com;", []Link{{Scheme: "http:", Start: 5, End: 23}}},
		{"text http://example.com:", []Link{{Scheme: "http:", Start: 5, End: 23}}},
		{"text http://example.com!", []Link{{Scheme: "http:", Start: 5, End: 23}}},
		{"text http://example.com?", []Link{{Scheme: "http:", Start: 5, End: 23}}},
		{"text http://example.com.", []Link{{Scheme: "http:", Start: 5, End: 23}}},
		{"text http://example.com'", []Link{{Scheme: "http:", Start: 5, End: 23}}},
		{"text http://example.com", []Link{{Scheme: "http:", Start: 5, End: 23}}},
		{"text http://example.com)", []Link{{Scheme: "http:", Start: 5, End: 23}}},
		{"text http://example.com]", []Link{{Scheme: "http:", Start: 5, End: 23}}},
		{"text http://example.com}", []Link{{Scheme: "http:", Start: 5, End: 23}}},
		{"text:http://example.com", []Link{{Scheme: "http:", Start: 5, End: 23}}},
		{"text !http://example.com", []Link{{Scheme: "http:", Start: 6, End: 24}}},
		{"text /http://example.com", []Link{{Scheme: "http:", Start: 6, End: 24}}},
		{"text (http://example.com)", []Link{{Scheme: "http:", Start: 6, End: 24}}},
		{"text \"http://example.com\"", []Link{{Scheme: "http:", Start: 6, End: 24}}},
		{"text http://example.com more text", []Link{{Scheme: "http:", Start: 5, End: 23}}},
		{"text =http://example.com", nil},
		{"text (http://example.com/test)", []Link{{Scheme: "http:", Start: 6, End: 29}}},
		{`text http://foo.com/("onclick="alert(1)")`, []Link{{Scheme: "http:", Start: 5, End: 20}}},
		{"text http://msdn.com/S(deadbeef)/page.htm", []Link{{Scheme: "http:", Start: 5, End: 41}}},
		{"text http://msdn.microsoft.com/ja-jp/library/system.net.httpwebrequest(v=VS.100).aspx", []Link{{Scheme: "http:", Start: 5, End: 85}}},
		{"text https://rdio.com/artist/50_Cent/album/We_Up/track/We_Up_(Album_Version_(Edited", []Link{{Scheme: "https:", Start: 5, End: 61}}},
		{"text https://rdio.com/artist/50_Cent/album/We_Up/track/We_Up_(Album_Version_(Edited))/", []Link{{Scheme: "https:", Start: 5, End: 86}}},
		{"text https://rdio.com/artist/50_Cent/album/We_Up/track/We_Up(URL description with spaces and (parentheses))", []Link{{Scheme: "https:", Start: 5, End: 107}}},
		{"text http://t.co/gksG6xlq", []Link{{Scheme: "http:", Start: 5, End: 25}}},
		{"text http://t.co/gksG6xlq text #hashtag text @username", []Link{{Scheme: "http:", Start: 5, End: 25}}},
		{"text (URL in parentheses http://msdn.com/S(deadbeef))", []Link{{Scheme: "http:", Start: 25, End: 52}}},
		{"text (URL in parentheses https://rdio.com/artist/50_Cent/album/We_Up/track/We_Up_(Album_Version_(Edited))/)", []Link{{Scheme: "https:", Start: 25, End: 106}}},
		{"@user Try http:// example.com/path", []Link{{Scheme: "", Start: 18, End: 34}}},
		{"@user Try http:// example.com/path", []Link{{Scheme: "", Start: 19, End: 35}}},

		// punycode

		{"See also: http://xn--80abe5aohbnkjb.xn--p1ai/", []Link{{Scheme: "http:", Start: 10, End: 45}}},
		{"xn--80abe5aohbnkjb.xn--p1ai/", []Link{{Scheme: "", Start: 0, End: 28}}},
		{"admin@xn--80abe5aohbnkjb.xn--p1ai", []Link{{Scheme: "mailto:", Start: 0, End: 33}}},
		{"mailto:admin@xn--80abe5aohbnkjb.xn--p1ai", []Link{{Scheme: "mailto:", Start: 0, End: 40}}},

		//{"いまなにしてるhttp://example.comいまなにしてる", []Link{}},
		//{"I enjoy Macintosh Brand computers: http://✪df.ws/ejp", []Link{}},
		//{"See: http://t.co/abcde's page", []Link{{Scheme: "http:", Start: 5, End: 22}}},
		//{"text http://example.com=", []Link{{Scheme: "http:", Start: 5, End: 23}}},
		//{"text http://example.com/pipe|character?yes|pipe|character", []Link{{Scheme: "http:", Start: 5, End: 57}}},

		{"http://☃.net/", []Link{{Scheme: "http:", Start: 0, End: 15}}},
	}
	for i, tc := range testCases {
		got := Links(tc.in)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("#%d: Links(%q):\n got %#v\nwant %#v", i, tc.in, got, tc.want)
		}
	}
}

func TestParseIPv4(t *testing.T) {
	type testCase struct {
		in     string
		length int
		ok     bool
	}
	testCases := []testCase{
		{"", 0, false},
		{"8.8.8.8", 7, true},
		{"8.8.8.", 0, false},
		{"8.8.8.xxx", 0, false},
		{"8.8.8.8xxx", 7, true},
		{"256.0.0.1", 0, false},
		{"001.001.001.001", 15, true},
	}
	for _, tc := range testCases {
		length, ok := skipIPv4(tc.in)
		if ok != tc.ok {
			s := "failed"
			if !tc.ok {
				s = "unexpectedly succeeded"
			}
			t.Errorf("skipIPv4(%q) %s", tc.in, s)
		} else if length != tc.length {
			t.Errorf("skipIPv4(%q) returned length %d, want %d", tc.in, length, tc.length)
		}
	}
}
