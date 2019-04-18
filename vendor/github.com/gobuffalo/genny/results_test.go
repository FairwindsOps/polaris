package genny

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Results_Find(t *testing.T) {
	res := Results{
		Files: []File{
			NewFileS("a.txt", "A"),
			NewFileS("b/b.txt", "B"),
		},
	}

	table := []struct {
		name string
		out  string
		err  bool
	}{
		{"a.txt", "A", false},
		{"b/b.txt", "B", false},
		{"c.txt", "", true},
	}

	for _, tt := range table {
		t.Run(tt.name, func(st *testing.T) {
			r := require.New(st)

			f, err := res.Find(tt.name)
			if tt.err {
				r.Error(err)
				return
			}
			r.Equal(tt.out, f.String())
		})
	}
}
