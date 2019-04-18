package gogen

import (
	"context"
	"strings"
	"testing"

	"github.com/gobuffalo/genny"
	"github.com/stretchr/testify/require"
)

func Test_GoFmt(t *testing.T) {
	r := require.New(t)

	run := genny.DryRunner(context.Background())

	g, err := Fmt("")
	r.NoError(err)
	run.With(g)

	r.NoError(run.Run())

}

func Test_FmtTransformer(t *testing.T) {
	r := require.New(t)

	f := genny.NewFile("foo.go", strings.NewReader(badFmt))

	ft := FmtTransformer()
	f, err := ft.Transform(f)
	r.NoError(err)

	fmted := f.String()
	r.NotEqual(badFmt, fmted)
	r.Equal(goodFmt, fmted)
}

const goodFmt = `package main

// comment

func fooo() {}
`

const badFmt = `package main
      // comment

		func   fooo(    ) {     }

`
