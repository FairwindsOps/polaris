package genny

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Replace(t *testing.T) {
	r := require.New(t)

	table := []struct {
		in      string
		out     string
		search  string
		replace string
	}{
		{in: "foo/-dot-git-keep", out: "foo/.git-keep", search: "-dot-", replace: "."},
		{in: "foo/dot-git-keep", out: "foo/dot-git-keep", search: "-dot-", replace: "."},
	}

	for _, tt := range table {
		in := NewFile(tt.in, nil)
		out, err := Replace(tt.search, tt.replace).Transform(in)
		r.NoError(err)
		r.Equal(tt.out, out.Name())
	}
}

func Test_Dot(t *testing.T) {
	r := require.New(t)

	f := NewFile("-dot-travis.yml", nil)
	f, err := Dot().Transform(f)
	r.NoError(err)
	r.Equal(".travis.yml", f.Name())
}
