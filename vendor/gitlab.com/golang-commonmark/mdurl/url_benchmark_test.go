// Copyright 2015 The Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mdurl

import "testing"

var slashedProtocolArray = [...]string{
	"http", "https", "ftp", "gopher", "file",
}

var slashedProtocolSlice = []string{
	"http", "https", "ftp", "gopher", "file",
}

func BenchmarkMapAccess(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = slashedProtocol["http"]
		_ = slashedProtocol["file"]
	}
}

func BenchmarkArrayAccess(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, p := range slashedProtocolArray {
			if p == "http" {
				break
			}
		}
		for _, p := range slashedProtocolArray {
			if p == "file" {
				break
			}
		}
	}
}

func BenchmarkSliceAccess(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, p := range slashedProtocolSlice {
			if p == "http" {
				break
			}
		}
		for _, p := range slashedProtocolSlice {
			if p == "file" {
				break
			}
		}
	}
}

func (u *URL) StringWithAppend() string {
	size := len(u.Path)
	if u.Scheme != "" {
		size += len(u.Scheme) + 1
	}
	if u.Slashes {
		size += 2
	}
	if u.Auth != "" {
		size += len(u.Auth) + 1
	}
	if u.Host != "" {
		size += len(u.Host)
		if u.IPv6 {
			size += 2
		}
	}
	if u.Port != "" {
		size += len(u.Port) + 1
	}
	if u.RawQuery != "" {
		size += len(u.RawQuery) + 1
	}
	if u.Fragment != "" {
		size += len(u.Fragment) + 1
	}
	if size == 0 {
		return ""
	}

	buf := make([]byte, 0, size)
	if u.Scheme != "" {
		buf = append(buf, u.Scheme...)
		buf = append(buf, ':')
	}
	if u.Slashes {
		buf = append(buf, '/', '/')
	}
	if u.Auth != "" {
		buf = append(buf, u.Auth...)
		buf = append(buf, '@')
	}
	if u.Host != "" {
		if u.IPv6 {
			buf = append(buf, '[')
			buf = append(buf, u.Host...)
			buf = append(buf, ']')
		} else {
			buf = append(buf, u.Host...)
		}
	}
	if u.Port != "" {
		buf = append(buf, ':')
		buf = append(buf, u.Port...)
	}
	buf = append(buf, u.Path...)
	if u.RawQuery != "" {
		buf = append(buf, '?')
		buf = append(buf, u.RawQuery...)
	}
	if u.Fragment != "" {
		buf = append(buf, '#')
		buf = append(buf, u.Fragment...)
	}
	return string(buf)
}

func BenchmarkString(b *testing.B) {
	url := URL{
		Scheme:   "http",
		Slashes:  true,
		Auth:     "admin:password",
		Host:     "example.com",
		Port:     "80",
		Path:     "/path",
		RawQuery: "query",
		Fragment: "frag",
		IPv6:     false,
	}
	for i := 0; i < b.N; i++ {
		url.String()
	}
}

func BenchmarkStringWithAppend(b *testing.B) {
	url := URL{
		Scheme:   "http",
		Slashes:  true,
		Auth:     "admin:password",
		Host:     "example.com",
		Port:     "80",
		Path:     "/path",
		RawQuery: "query",
		Fragment: "frag",
		IPv6:     false,
	}
	for i := 0; i < b.N; i++ {
		url.StringWithAppend()
	}
}
