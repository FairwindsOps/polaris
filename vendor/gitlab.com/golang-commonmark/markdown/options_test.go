// Copyright 2015 The Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package markdown

import (
	"strings"
	"testing"
)

func TestQuote(t *testing.T) {
	quotes := "‘’“”"
	md := New(Quotes(quotes))
	if strings.Join(md.Quotes[:], "") != quotes {
		t.Errorf("expected %q, got %q", quotes, md.Quotes)
	}
}
