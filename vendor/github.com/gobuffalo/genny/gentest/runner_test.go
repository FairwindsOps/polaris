package gentest

import (
	"errors"
	"testing"

	"github.com/gobuffalo/genny"
	"github.com/stretchr/testify/require"
)

func Test_Run(t *testing.T) {
	r := require.New(t)

	g := genny.New()
	g.File(genny.NewFileS("a.txt", "A"))
	g.File(genny.NewFileS("b.txt", "B"))

	res, err := Run(g)
	r.NoError(err)
	r.Len(res.Files, 2)
}

func Test_RunNew(t *testing.T) {
	r := require.New(t)

	g := genny.New()
	g.File(genny.NewFileS("a.txt", "A"))
	g.File(genny.NewFileS("b.txt", "B"))

	res, err := RunNew(func() (*genny.Generator, error) {
		return g, nil
	}())
	r.NoError(err)
	r.Len(res.Files, 2)
}

func Test_RunNew_Error(t *testing.T) {
	r := require.New(t)

	res, err := RunNew(func() (*genny.Generator, error) {
		return nil, errors.New("boom")
	}())
	r.Error(err)
	r.Len(res.Files, 0)
}

func Test_RunGroup(t *testing.T) {
	r := require.New(t)

	g := genny.New()
	g.File(genny.NewFileS("a.txt", "A"))
	g.File(genny.NewFileS("b.txt", "B"))

	gg := &genny.Group{}
	gg.Add(g)

	res, err := RunGroup(gg)
	r.NoError(err)
	r.Len(res.Files, 2)
}
