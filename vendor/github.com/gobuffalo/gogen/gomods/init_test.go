package gomods

import (
	"go/build"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/genny"
	"github.com/gobuffalo/genny/gentest"
	"github.com/stretchr/testify/require"
)

func Test_New_No_Modules(t *testing.T) {
	envy.Temp(func() {
		envy.MustSet(ENV, "off")

		r := require.New(t)

		run := gentest.NewRunner()
		gg, err := New("foo/bar", "")
		r.NoError(err)
		run.WithGroup(gg)
		r.NoError(run.Run())

		res := run.Results()
		r.Len(res.Files, 0)
		r.Len(res.Commands, 0)
	})
}

func Test_New_With_Modules(t *testing.T) {
	envy.Temp(func() {
		envy.MustSet(ENV, "on")

		type ts struct {
			name string
			path string
			mod  string
		}
		table := []ts{
			{"", "", "github.com/gobuffalo/gogen/gomods"},
			{"coke", "", "coke"},
			{"github.com\\gobuffalo\\coke", "", "github.com/gobuffalo/coke"},
			{"", "github.com\\gobuffalo\\coke", "github.com/gobuffalo/coke"},
		}

		c := build.Default

		for _, src := range c.SrcDirs() {
			table = append(table, ts{
				name: "coke",
				path: src,
				mod:  "coke",
			})
			table = append(table, ts{
				name: "coke",
				path: strings.ToLower(src),
				mod:  "coke",
			})
			table = append(table, ts{
				name: "",
				path: filepath.Join(src, "github.com", "gobuffalo", "coke"),
				mod:  "github.com/gobuffalo/coke",
			})
		}

		for _, tt := range table {
			t.Run(tt.name+tt.path+tt.mod, func(st *testing.T) {
				r := require.New(st)

				run := gentest.NewRunner()

				gg, err := New(tt.name, tt.path)
				r.NoError(err)
				run.WithGroup(gg)

				r.NoError(run.Run())
				res := run.Results()
				r.Len(res.Files, 0)
				r.Len(res.Commands, 2)

				c := res.Commands[0]
				args := []string{genny.GoBin(), "mod", "init"}
				if len(tt.mod) > 0 {
					args = append(args, tt.mod)
				}
				r.Equal(args, c.Args)

				c = res.Commands[1]
				r.Equal([]string{genny.GoBin(), "mod", "tidy"}, c.Args)
			})
		}

	})
}
