// Copyright 2015 The Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mdurl

import (
	"strings"
	"testing"
)

func TestFindScheme(t *testing.T) {
	type testCase struct {
		in   string
		want int
		err  error
	}
	testCases := []testCase{
		{in: "", want: 0},
		{in: ":ptth", err: ErrMissingScheme},
		{in: "0http://", want: 0},
		{in: "h!ttp://", want: 0},
		{in: "http", want: 0},
		{in: "http://", want: 4},
		{in: "ed2k://", want: 4},
	}
	for _, tc := range testCases {
		got, err := findScheme(tc.in)
		if err != nil {
			if err != tc.err {
				t.Errorf("findScheme(%q): want error %v, got %v", tc.in, tc.err, err)
			}
		} else {
			if got != tc.want {
				t.Errorf("findScheme(%q) = %d, want %d", tc.in, got, tc.want)
			}
		}
	}
}

func TestParse(t *testing.T) {
	type testCase struct {
		in   string
		want URL
		err  error
	}
	testCases := []testCase{
		{in: "", want: URL{}},
		{
			in: "http://example.com:",
			want: URL{
				Scheme:    "http",
				RawScheme: "http",
				Slashes:   true,
				Host:      "example.com",
				Path:      ":",
			},
		},
		{
			in: "git+ssh://git@github.com:npm/npm",
			want: URL{
				Scheme:    "git+ssh",
				RawScheme: "git+ssh",
				Slashes:   true,
				Auth:      "git",
				Host:      "github.com",
				Path:      ":npm/npm",
			},
		},
		{
			in: "coap:[fedc:ba98:7654:3210:fedc:ba98:7654:3210]:61616/s/stopButton",
			want: URL{
				Scheme:    "coap",
				RawScheme: "coap",
				Host:      "fedc:ba98:7654:3210:fedc:ba98:7654:3210",
				Port:      "61616",
				Path:      "/s/stopButton",
				IPv6:      true,
			},
		},
		{
			in: "file://foo/etc/passwd",
			want: URL{
				Scheme:    "file",
				RawScheme: "file",
				Slashes:   true,
				Host:      "foo",
				Path:      "/etc/passwd",
			},
		},
		{
			in: "/foo/bar?baz=quux#frag",
			want: URL{
				Path:        "/foo/bar",
				RawQuery:    "baz=quux",
				HasQuery:    true,
				Fragment:    "frag",
				HasFragment: true,
			},
		},
		{
			in: "javascript:alert(\"hello\");",
			want: URL{
				Scheme:    "javascript",
				RawScheme: "javascript",
				Path:      "alert(\"hello\");",
			},
		},
		{
			in: "www.example.com",
			want: URL{
				Path: "www.example.com",
			},
		},
		{
			in: "coap://[1080:0:0:0:8:800:200C:417A]:61616/",
			want: URL{
				Scheme:    "coap",
				RawScheme: "coap",
				Slashes:   true,
				Host:      "1080:0:0:0:8:800:200C:417A",
				Port:      "61616",
				Path:      "/",
				IPv6:      true,
			},
		},
		{
			in: "http://user@www.example.com/",
			want: URL{
				Scheme:    "http",
				RawScheme: "http",
				Slashes:   true,
				Auth:      "user",
				Host:      "www.example.com",
				Path:      "/",
			},
		},
		{
			in: "http://x.y.com+a/b/c",
			want: URL{
				Scheme:    "http",
				RawScheme: "http",
				Slashes:   true,
				Host:      "x.y.com+a",
				Path:      "/b/c",
			},
		},
		{
			in: "http://➡.ws/➡",
			want: URL{
				Scheme:    "http",
				RawScheme: "http",
				Slashes:   true,
				Host:      "➡.ws",
				Path:      "/➡",
			},
		},
		{
			in: "http://user:pass@-lovemonsterz.tumblr.com:80/rss",
			want: URL{
				Scheme:    "http",
				RawScheme: "http",
				Slashes:   true,
				Auth:      "user:pass",
				Host:      "-lovemonsterz.tumblr.com",
				Port:      "80",
				Path:      "/rss",
			},
		},
		{
			in: "http://a@b?@c",
			want: URL{
				Scheme:    "http",
				RawScheme: "http",
				Slashes:   true,
				Auth:      "a",
				Host:      "b",
				RawQuery:  "@c",
				HasQuery:  true,
			},
		},
		{
			in: "HtTp://x.y.cOm;A/b/c?d=e#f g<h>i",
			want: URL{
				Scheme:      "http",
				RawScheme:   "HtTp",
				Slashes:     true,
				Host:        "x.y.cOm",
				Path:        ";A/b/c",
				RawQuery:    "d=e",
				HasQuery:    true,
				Fragment:    "f g<h>i",
				HasFragment: true,
			},
		},
		{
			in: "file:///etc/node/",
			want: URL{
				Scheme:    "file",
				RawScheme: "file",
				Slashes:   true,
				Path:      "/etc/node/",
			},
		},
		{
			in: "http://atpass:foo%40bar@127.0.0.1:8080/path?search=foo#bar",
			want: URL{
				Scheme:      "http",
				RawScheme:   "http",
				Slashes:     true,
				Auth:        "atpass:foo%40bar",
				Host:        "127.0.0.1",
				Port:        "8080",
				Path:        "/path",
				RawQuery:    "search=foo",
				HasQuery:    true,
				Fragment:    "bar",
				HasFragment: true,
			},
		},
		{
			in: "svn+ssh://foo/bar",
			want: URL{
				Scheme:    "svn+ssh",
				RawScheme: "svn+ssh",
				Slashes:   true,
				Host:      "foo",
				Path:      "/bar",
			},
		},
		{
			in: "http://user:password@[3ffe:2a00:100:7031::1]:8080",
			want: URL{
				Scheme:    "http",
				RawScheme: "http",
				Slashes:   true,
				Auth:      "user:password",
				Host:      "3ffe:2a00:100:7031::1",
				Port:      "8080",
				IPv6:      true,
			},
		},
		{
			in: "mailto:foo@bar.com?subject=hello",
			want: URL{
				Scheme:    "mailto",
				RawScheme: "mailto",
				Auth:      "foo",
				Host:      "bar.com",
				RawQuery:  "subject=hello",
				HasQuery:  true,
			},
		},
		{
			in: "http://_jabber._tcp.google.com/test",
			want: URL{
				Scheme:    "http",
				RawScheme: "http",
				Slashes:   true,
				Host:      "_jabber._tcp.google.com",
				Path:      "/test",
			},
		},
		{
			in: "git+http://github.com/joyent/node.git",
			want: URL{
				Scheme:    "git+http",
				RawScheme: "git+http",
				Slashes:   true,
				Host:      "github.com",
				Path:      "/joyent/node.git",
			},
		},
		{
			in: "coap://[FEDC:BA98:7654:3210:FEDC:BA98:7654:3210]",
			want: URL{
				Scheme:    "coap",
				RawScheme: "coap",
				Slashes:   true,
				Host:      "FEDC:BA98:7654:3210:FEDC:BA98:7654:3210",
				IPv6:      true,
			},
		},
		{
			in: "http://example.com#frag=?bar/#frag",
			want: URL{
				Scheme:      "http",
				RawScheme:   "http",
				Slashes:     true,
				Host:        "example.com",
				Fragment:    "frag=?bar/#frag",
				HasFragment: true,
			},
		},
		{
			in: "http://example.com:?a=b",
			want: URL{
				Scheme:    "http",
				RawScheme: "http",
				Slashes:   true,
				Host:      "example.com",
				Path:      ":",
				RawQuery:  "a=b",
				HasQuery:  true,
			},
		},
		{
			in: "http://user:pass@_jabber._tcp.google.com:80/test",
			want: URL{
				Scheme:    "http",
				RawScheme: "http",
				Slashes:   true,
				Auth:      "user:pass",
				Host:      "_jabber._tcp.google.com",
				Port:      "80",
				Path:      "/test",
			},
		},
		{
			in: "http://example.com?foo=/bar/#frag",
			want: URL{
				Scheme:      "http",
				RawScheme:   "http",
				Slashes:     true,
				Host:        "example.com",
				RawQuery:    "foo=/bar/",
				HasQuery:    true,
				Fragment:    "frag",
				HasFragment: true,
			},
		},
		{
			in: "HTTP://www.example.com/",
			want: URL{
				Scheme:    "http",
				RawScheme: "HTTP",
				Slashes:   true,
				Host:      "www.example.com",
				Path:      "/",
			},
		},
		{
			in: "http://USER:PW@www.ExAmPlE.com/",
			want: URL{
				Scheme:    "http",
				RawScheme: "http",
				Slashes:   true,
				Auth:      "USER:PW",
				Host:      "www.ExAmPlE.com",
				Path:      "/",
			},
		},
		{
			in: "<http://goo.corn/bread> Is a URL!",
			want: URL{
				Path: "<http://goo.corn/bread> Is a URL!",
			},
		},
		{
			in: "file://foo/etc/node/",
			want: URL{
				Scheme:    "file",
				RawScheme: "file",
				Slashes:   true,
				Host:      "foo",
				Path:      "/etc/node/",
			},
		},
		{
			in: "dash-test://foo/bar",
			want: URL{
				Scheme:    "dash-test",
				RawScheme: "dash-test",
				Slashes:   true,
				Host:      "foo",
				Path:      "/bar",
			},
		},
		{
			in: "http://atpass:foo%40bar@127.0.0.1/",
			want: URL{
				Scheme:    "http",
				RawScheme: "http",
				Slashes:   true,
				Auth:      "atpass:foo%40bar",
				Host:      "127.0.0.1",
				Path:      "/",
			},
		},
		{
			in: "coap:u:p@[::1]:61616/.well-known/r?n=Temperature",
			want: URL{
				Scheme:    "coap",
				RawScheme: "coap",
				Auth:      "u:p",
				Host:      "::1",
				Port:      "61616",
				Path:      "/.well-known/r",
				RawQuery:  "n=Temperature",
				HasQuery:  true,
				IPv6:      true,
			},
		},
		{
			in: "http://user:pw@www.ExAmPlE.com/",
			want: URL{
				Scheme:    "http",
				RawScheme: "http",
				Slashes:   true,
				Auth:      "user:pw",
				Host:      "www.ExAmPlE.com",
				Path:      "/",
			},
		},
		{
			in: "http://example.Bücher.com/",
			want: URL{
				Scheme:    "http",
				RawScheme: "http",
				Slashes:   true,
				Host:      "example.Bücher.com",
				Path:      "/",
			},
		},
		{
			in: "[fe80::1]",
			want: URL{
				Path: "[fe80::1]",
			},
		},
		{
			in: "http://a.com/a/b/c?s#h",
			want: URL{
				Scheme:      "http",
				RawScheme:   "http",
				Slashes:     true,
				Host:        "a.com",
				Path:        "/a/b/c",
				RawQuery:    "s",
				HasQuery:    true,
				Fragment:    "h",
				HasFragment: true,
			},
		},
		{
			in: "http://ex.com/foo%3F100%m%23r?abc=the%231?&foo=bar#frag",
			want: URL{
				Scheme:      "http",
				RawScheme:   "http",
				Slashes:     true,
				Host:        "ex.com",
				Path:        "/foo%3F100%m%23r",
				RawQuery:    "abc=the%231?&foo=bar",
				HasQuery:    true,
				Fragment:    "frag",
				HasFragment: true,
			},
		},
		{
			in: "http://ex.com/fooA100%mBr?abc=the%231?&foo=bar#frag",
			want: URL{
				Scheme:      "http",
				RawScheme:   "http",
				Slashes:     true,
				Host:        "ex.com",
				Path:        "/fooA100%mBr",
				RawQuery:    "abc=the%231?&foo=bar",
				HasQuery:    true,
				Fragment:    "frag",
				HasFragment: true,
			},
		},
		{
			in: "http://x/p/\"quoted\"",
			want: URL{
				Scheme:    "http",
				RawScheme: "http",
				Slashes:   true,
				Host:      "x",
				Path:      "/p/\"quoted\"",
			},
		},
		{
			in: "dot.test://foo/bar",
			want: URL{
				Scheme:    "dot.test",
				RawScheme: "dot.test",
				Slashes:   true,
				Host:      "foo",
				Path:      "/bar",
			},
		},
		{
			in: "http://SÉLIER.COM/",
			want: URL{
				Scheme:    "http",
				RawScheme: "http",
				Slashes:   true,
				Host:      "SÉLIER.COM",
				Path:      "/",
			},
		},
		{
			in: "http://example.com?foo=bar#frag",
			want: URL{
				Scheme:      "http",
				RawScheme:   "http",
				Slashes:     true,
				Host:        "example.com",
				RawQuery:    "foo=bar",
				HasQuery:    true,
				Fragment:    "frag",
				HasFragment: true,
			},
		},
		{
			in: "http://google.com\" onload=\"alert(42)/",
			want: URL{
				Scheme:    "http",
				RawScheme: "http",
				Slashes:   true,
				Host:      "google.com",
				Path:      "\" onload=\"alert(42)/",
			},
		},
		{
			in: "http://a@b@c/",
			want: URL{
				Scheme:    "http",
				RawScheme: "http",
				Slashes:   true,
				Auth:      "a@b",
				Host:      "c",
				Path:      "/",
			},
		},
		{
			in: "http://example.com?foo=@bar#frag",
			want: URL{
				Scheme:      "http",
				RawScheme:   "http",
				Slashes:     true,
				Host:        "example.com",
				RawQuery:    "foo=@bar",
				HasQuery:    true,
				Fragment:    "frag",
				HasFragment: true,
			},
		},
		{
			in: "//some_path",
			want: URL{
				Slashes: true,
				Host:    "some_path",
			},
		},
		{
			in: "xmpp:isaacschlueter@jabber.org",
			want: URL{
				Scheme:    "xmpp",
				RawScheme: "xmpp",
				Auth:      "isaacschlueter",
				Host:      "jabber.org",
			},
		},
		{
			in: "dot.test:foo/bar",
			want: URL{
				Scheme:    "dot.test",
				RawScheme: "dot.test",
				Host:      "foo",
				Path:      "/bar",
			},
		},
		{
			in: "http://example.com:/a/b.html",
			want: URL{
				Scheme:    "http",
				RawScheme: "http",
				Slashes:   true,
				Host:      "example.com",
				Path:      ":/a/b.html",
			},
		},
		{
			in: "http://user:pass@_jabber._tcp.google.com/test",
			want: URL{
				Scheme:    "http",
				RawScheme: "http",
				Slashes:   true,
				Auth:      "user:pass",
				Host:      "_jabber._tcp.google.com",
				Path:      "/test",
			},
		},
		{
			in: "http://x.com/path?that\"s#all, folks",
			want: URL{
				Scheme:      "http",
				RawScheme:   "http",
				Slashes:     true,
				Host:        "x.com",
				Path:        "/path",
				RawQuery:    "that\"s",
				HasQuery:    true,
				Fragment:    "all, folks",
				HasFragment: true,
			},
		},
		{
			in: "http://www.narwhaljs.org/blog/categories?id=news",
			want: URL{
				Scheme:    "http",
				RawScheme: "http",
				Slashes:   true,
				Host:      "www.narwhaljs.org",
				Path:      "/blog/categories",
				RawQuery:  "id=news",
				HasQuery:  true,
			},
		},
		{
			in: "file://localhost/etc/node/",
			want: URL{
				Scheme:    "file",
				RawScheme: "file",
				Slashes:   true,
				Host:      "localhost",
				Path:      "/etc/node/",
			},
		},
		{
			in: "http:/baz/../foo/bar",
			want: URL{
				Scheme:    "http",
				RawScheme: "http",
				Path:      "/baz/../foo/bar",
			},
		},
		{
			in: "http://user:pass@-lovemonsterz.tumblr.com/rss",
			want: URL{
				Scheme:    "http",
				RawScheme: "http",
				Slashes:   true,
				Auth:      "user:pass",
				Host:      "-lovemonsterz.tumblr.com",
				Path:      "/rss",
			},
		},
		{
			in: "http://x:1/\" <>\"`/{}|\\^~`/",
			want: URL{
				Scheme:    "http",
				RawScheme: "http",
				Slashes:   true,
				Host:      "x",
				Port:      "1",
				Path:      "/\" <>\"`/{}|\\^~`/",
			},
		},
		{
			in: "http://www.ExAmPlE.com/",
			want: URL{
				Scheme:    "http",
				RawScheme: "http",
				Slashes:   true,
				Host:      "www.ExAmPlE.com",
				Path:      "/",
			},
		},
		{
			in: "http://user%3Apw@www.example.com/",
			want: URL{
				Scheme:    "http",
				RawScheme: "http",
				Slashes:   true,
				Auth:      "user%3Apw",
				Host:      "www.example.com",
				Path:      "/",
			},
		},
		{
			in: "http://user:pass@mt0.google.com/vt/lyrs=m@114???&hl=en&src=api&x=2&y=2&z=3&s=",
			want: URL{
				Scheme:    "http",
				RawScheme: "http",
				Slashes:   true,
				Auth:      "user:pass",
				Host:      "mt0.google.com",
				Path:      "/vt/lyrs=m@114",
				RawQuery:  "??&hl=en&src=api&x=2&y=2&z=3&s=",
				HasQuery:  true,
			},
		},
		{
			in: "http:/foo/bar?baz=quux#frag",
			want: URL{
				Scheme:      "http",
				RawScheme:   "http",
				Path:        "/foo/bar",
				RawQuery:    "baz=quux",
				HasQuery:    true,
				Fragment:    "frag",
				HasFragment: true,
			},
		},
		{
			in: "http://example.com:#abc",
			want: URL{
				Scheme:      "http",
				RawScheme:   "http",
				Slashes:     true,
				Host:        "example.com",
				Path:        ":",
				Fragment:    "abc",
				HasFragment: true,
			},
		},
		{
			in: "http://-lovemonsterz.tumblr.com:80/rss",
			want: URL{
				Scheme:    "http",
				RawScheme: "http",
				Slashes:   true,
				Host:      "-lovemonsterz.tumblr.com",
				Port:      "80",
				Path:      "/rss",
			},
		},
		{
			in: "http://example.com?foo=?bar/#frag",
			want: URL{
				Scheme:      "http",
				RawScheme:   "http",
				Slashes:     true,
				Host:        "example.com",
				RawQuery:    "foo=?bar/",
				HasQuery:    true,
				Fragment:    "frag",
				HasFragment: true,
			},
		},
		{
			in: "http://user:pass@example.com:8000/foo/bar?baz=quux#frag",
			want: URL{
				Scheme:      "http",
				RawScheme:   "http",
				Slashes:     true,
				Auth:        "user:pass",
				Host:        "example.com",
				Port:        "8000",
				Path:        "/foo/bar",
				RawQuery:    "baz=quux",
				HasQuery:    true,
				Fragment:    "frag",
				HasFragment: true,
			},
		},
		{
			in: "dash-test:foo/bar",
			want: URL{
				Scheme:    "dash-test",
				RawScheme: "dash-test",
				Host:      "foo",
				Path:      "/bar",
			},
		},
		{
			in: "http://bucket_name.s3.amazonaws.com/image.jpg",
			want: URL{
				Scheme:    "http",
				RawScheme: "http",
				Slashes:   true,
				Host:      "bucket_name.s3.amazonaws.com",
				Path:      "/image.jpg",
			},
		},
		{
			in: "local1@domain1",
			want: URL{
				Path: "local1@domain1",
			},
		},
		{
			in: "http://-lovemonsterz.tumblr.com/rss",
			want: URL{
				Scheme:    "http",
				RawScheme: "http",
				Slashes:   true,
				Host:      "-lovemonsterz.tumblr.com",
				Path:      "/rss",
			},
		},
		{
			in: "HTTP://www.example.com",
			want: URL{
				Scheme:    "http",
				RawScheme: "HTTP",
				Slashes:   true,
				Host:      "www.example.com",
			},
		},
		{
			in: "//user:pass@example.com:8000/foo/bar?baz=quux#frag",
			want: URL{
				Slashes:     true,
				Auth:        "user:pass",
				Host:        "example.com",
				Port:        "8000",
				Path:        "/foo/bar",
				RawQuery:    "baz=quux",
				HasQuery:    true,
				Fragment:    "frag",
				HasFragment: true,
			},
		},
		{
			in: "http://_jabber._tcp.google.com:80/test",
			want: URL{
				Scheme:    "http",
				RawScheme: "http",
				Slashes:   true,
				Host:      "_jabber._tcp.google.com",
				Port:      "80",
				Path:      "/test",
			},
		},
		{
			in: "http://a\r\" \t\n<\"b:b@c\r\nd/e?f",
			want: URL{
				Scheme:    "http",
				RawScheme: "http",
				Slashes:   true,
				Auth:      "a\r\" \t\n<\"b:b",
				Host:      "c",
				Path:      "\r\nd/e",
				RawQuery:  "f",
				HasQuery:  true,
			},
		},
		{
			in: "file://localhost/etc/passwd",
			want: URL{
				Scheme:    "file",
				RawScheme: "file",
				Slashes:   true,
				Host:      "localhost",
				Path:      "/etc/passwd",
			},
		},
		{
			in: "http://www.Äffchen.cOm;A/b/c?d=e#f g<h>i",
			want: URL{
				Scheme:      "http",
				RawScheme:   "http",
				Slashes:     true,
				Host:        "www.Äffchen.cOm",
				Path:        ";A/b/c",
				RawQuery:    "d=e",
				HasQuery:    true,
				Fragment:    "f g<h>i",
				HasFragment: true,
			},
		},
		{
			in: "http://ليهمابتكلموشعربي؟.ي؟/",
			want: URL{
				Scheme:    "http",
				RawScheme: "http",
				Slashes:   true,
				Host:      "ليهمابتكلموشعربي؟.ي؟",
				Path:      "/",
			},
		},
		{
			in: "HtTp://x.y.cOm;a/b/c?d=e#f g<h>i",
			want: URL{
				Scheme:      "http",
				RawScheme:   "HtTp",
				Slashes:     true,
				Host:        "x.y.cOm",
				Path:        ";a/b/c",
				RawQuery:    "d=e",
				HasQuery:    true,
				Fragment:    "f g<h>i",
				HasFragment: true,
			},
		},
		{
			in: "http://x...y...#p",
			want: URL{
				Scheme:      "http",
				RawScheme:   "http",
				Slashes:     true,
				Host:        "x...y...",
				Fragment:    "p",
				HasFragment: true,
			},
		},
		{
			in: "http://mt0.google.com/vt/lyrs=m@114&hl=en&src=api&x=2&y=2&z=3&s=",
			want: URL{
				Scheme:    "http",
				RawScheme: "http",
				Slashes:   true,
				Host:      "mt0.google.com",
				Path:      "/vt/lyrs=m@114&hl=en&src=api&x=2&y=2&z=3&s=",
			},
		},
		{
			in: "http://mt0.google.com/vt/lyrs=m@114???&hl=en&src=api&x=2&y=2&z=3&s=",
			want: URL{
				Scheme:    "http",
				RawScheme: "http",
				Slashes:   true,
				Host:      "mt0.google.com",
				Path:      "/vt/lyrs=m@114",
				RawQuery:  "??&hl=en&src=api&x=2&y=2&z=3&s=",
				HasQuery:  true,
			},
		},
		{
			in: "file:///etc/passwd",
			want: URL{
				Scheme:    "file",
				RawScheme: "file",
				Slashes:   true,
				Path:      "/etc/passwd",
			},
		},
		{
			in: "http://[fe80::1]:/a/b?a=b#abc",
			want: URL{
				Scheme:      "http",
				RawScheme:   "http",
				Slashes:     true,
				Host:        "fe80::1",
				Path:        ":/a/b",
				RawQuery:    "a=b",
				HasQuery:    true,
				Fragment:    "abc",
				HasFragment: true,
				IPv6:        true,
			},
		},
		{
			in: "HTTP://X.COM/Y",
			want: URL{
				Scheme:    "http",
				RawScheme: "HTTP",
				Slashes:   true,
				Host:      "X.COM",
				Path:      "/Y",
			},
		},
		{
			in: "http://www.日本語.com/",
			want: URL{
				Scheme:    "http",
				RawScheme: "http",
				Slashes:   true,
				Host:      "www.日本語.com",
				Path:      "/",
			},
		},
		{
			in: "http://www.Äffchen.com/",
			want: URL{
				Scheme:    "http",
				RawScheme: "http",
				Slashes:   true,
				Host:      "www.Äffchen.com",
				Path:      "/",
			},
		},
		{
			in: "coap://u:p@[::192.9.5.5]:61616/.well-known/r?n=Temperature",
			want: URL{
				Scheme:    "coap",
				RawScheme: "coap",
				Slashes:   true,
				Auth:      "u:p",
				Host:      "::192.9.5.5",
				Port:      "61616",
				Path:      "/.well-known/r",
				RawQuery:  "n=Temperature",
				HasQuery:  true,
				IPv6:      true,
			},
		},
		{
			in: "http://atslash%2F%40:%2F%40@foo/",
			want: URL{
				Scheme:    "http",
				RawScheme: "http",
				Slashes:   true,
				Auth:      "atslash%2F%40:%2F%40",
				Host:      "foo",
				Path:      "/",
			},
		},
		{
			in:  ":ptth",
			err: ErrMissingScheme,
		},
	}
	for _, tc := range testCases {
		got, err := Parse(tc.in)
		if err != nil {
			if err != tc.err {
				t.Errorf("Parse(%q): want error %v, got %v", tc.in, tc.err, err)
			}
		} else {
			if *got != tc.want {
				t.Errorf("Parse(%q):\n got %#v\nwant %#v", tc.in, *got, tc.want)
			}
		}
	}
}

func TestEncode(t *testing.T) {
	type testCase struct {
		in   string
		want string
	}
	testCases := []testCase{
		{"", ""},
		{";/?:@&=+$,-_.!~*'()#012azAZ", ";/?:@&=+$,-_.!~*'()#012azAZ"},
		{" `%^{}\"<>\\", "%20%60%25%5E%7B%7D%22%3C%3E%5C"},
		{"%7e%25%7E", "%7E%25%7E"},
		{"%1g%z2%", "%251g%25z2%25"},
		{"процесс", "%D0%BF%D1%80%D0%BE%D1%86%D0%B5%D1%81%D1%81"},
	}
	for _, tc := range testCases {
		got := Encode(tc.in)
		if got != tc.want {
			t.Errorf("Encode(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestDecode(t *testing.T) {
	type testCase struct {
		in   string
		want string
	}
	testCases := []testCase{
		{"", ""},
		{"a%2520b", "a%20b"},
		{"%7e", "~"},
		{"%1g%z2%", "%1g%z2%"},
		{"%D0%BF%D1%80%D0%BE%D1%86%D0%B5%D1%81%D1%81", "процесс"},
		{"%D0", "\xEF\xBF\xBD"},
		{"%D0%25", "\xEF\xBF\xBD%"},
		{"%D0%", "\xEF\xBF\xBD%"},
		{"%D0%1g", "\xEF\xBF\xBD%1g"},
		{"%D0%1g", "\xEF\xBF\xBD%1g"},
		{"%D0\xBF", "п"},
		{"%D0%BF", "п"},
		{"%D0\x25", "\xEF\xBF\xBD%"},
		{"%80%81%82%83%84%85%86%87%88%89%8A%8B%8C%8D%8E%8F%90%91%92%93%94%95%96%97%98%99%9A%9B%9C%9D%9E%9F%A0%A1%A2%A3%A4%A5%A6%A7%A8%A9%AA%AB%AC%AD%AE%AF%B0%B1%B2%B3%B4%B5%B6%B7%B8%B9%BA%BB%BC%BD%BE%BF%F8%F9%FA%FB%FC%FD%FE", strings.Repeat("\xEF\xBF\xBD", 71)},
		{"%e3%81%84%e3%81%be%e3%81%aa%e3%81%ab%e3%81%97%e3%81%a6%e3%82%8b", "いまなにしてる"},
		{"%f0%90%80%80", "\U00010000"},
	}
	for _, tc := range testCases {
		got := Decode(tc.in)
		if got != tc.want {
			t.Errorf("Decode(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestString(t *testing.T) {
	type testCase struct {
		in   URL
		want string
	}
	testCases := []testCase{
		{URL{}, ""},
		{
			URL{
				Scheme:      "http",
				RawScheme:   "http",
				Slashes:     true,
				Auth:        "admin:password",
				Host:        "example.com",
				Port:        "80",
				Path:        "/path",
				RawQuery:    "query",
				HasQuery:    true,
				Fragment:    "frag",
				HasFragment: true,
				IPv6:        false,
			},
			"http://admin:password@example.com:80/path?query#frag",
		},
		{
			URL{
				Host: "::1",
				IPv6: true,
			},
			"[::1]",
		},
	}
	for _, tc := range testCases {
		got := tc.in.String()
		if got != tc.want {
			t.Errorf("String(%#v):\n got %q\nwant %q", tc.in, got, tc.want)
		}
	}
}
