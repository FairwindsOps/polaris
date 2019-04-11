package gogen

import (
	"strings"
	"testing"

	"github.com/gobuffalo/genny"
	"github.com/stretchr/testify/require"
)

func Test_PackageName(t *testing.T) {
	table := []struct {
		pass bool
		name string
		body string
		pkg  string
	}{
		{true, "version.go", "package foo\n", "foo"},
		{true, "foo/version.go", "package foo\n", "foo"},
		{true, "foo/version.go", "", "foo"},
		{false, "", "", ""},
	}

	for _, tt := range table {
		t.Run(tt.name+"|"+tt.body, func(st *testing.T) {
			r := require.New(st)
			f := genny.NewFile(tt.name, strings.NewReader(tt.body))

			pkg, err := PackageName(f)
			if tt.pass {
				r.NoError(err)
			} else {
				r.Error(err)
			}
			r.Equal(tt.pkg, pkg)
		})
	}

}
