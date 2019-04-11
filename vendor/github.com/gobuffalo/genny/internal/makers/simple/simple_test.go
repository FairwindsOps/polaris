package simple

import (
	"errors"
	"testing"

	"github.com/gobuffalo/genny"
	"github.com/gobuffalo/genny/gentest"
	"github.com/stretchr/testify/require"
)

func Test_Happy(t *testing.T) {
	r := require.New(t)

	run := gentest.NewRunner()
	run.Disk.Add(genny.NewFileS("main.go", "my main.go file"))

	g := New()
	run.With(g)

	r.NoError(run.Run())
	res := run.Results()

	cmds := []string{"go env", "genny -h"}
	r.NoError(gentest.CompareCommands(cmds, res.Commands))

	files := []string{"index.html", "main.go"}
	r.NoError(gentest.CompareFiles(files, res.Files))
}

func Test_Missing_Genny(t *testing.T) {
	r := require.New(t)

	run := gentest.NewRunner()
	run.Disk.Add(genny.NewFileS("main.go", "my main.go file"))

	g := New()
	run.With(g)

	// pretend we can't find genny
	run.LookPathFn = func(s string) (string, error) {
		if s == "genny" {
			return "", errors.New("can't find genny")
		}
		return s, nil
	}

	r.NoError(run.Run())
	res := run.Results()

	cmds := []string{"go env", "go get github.com/gobuffalo/genny/genny", "genny -h"}
	r.NoError(gentest.CompareCommands(cmds, res.Commands))

	files := []string{"index.html", "main.go"}
	r.NoError(gentest.CompareFiles(files, res.Files))
}
