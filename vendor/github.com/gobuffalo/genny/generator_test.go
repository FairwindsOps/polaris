package genny

import (
	"context"
	"os/exec"
	"strings"
	"testing"

	"github.com/gobuffalo/packd"
	"github.com/stretchr/testify/require"
)

var fixtures = func() packd.Box {
	box := packd.NewMemoryBox()
	box.AddString("foo.txt", "foo!")
	box.AddString("bar/baz.txt", "baz!")
	return box
}()

func Test_Generator_File(t *testing.T) {
	r := require.New(t)

	g := New()
	g.File(NewFile("foo.txt", strings.NewReader("hello")))

	run := DryRunner(context.Background())
	run.With(g)
	r.NoError(run.Run())

	res := run.Results()
	r.Len(res.Commands, 0)
	r.Len(res.Files, 1)

	f := res.Files[0]
	r.Equal("foo.txt", f.Name())
	r.Equal("hello", f.String())
}

func Test_Generator_Box(t *testing.T) {
	r := require.New(t)

	g := New()
	r.NoError(g.Box(fixtures))

	run := DryRunner(context.Background())
	run.With(g)
	r.NoError(run.Run())

	res := run.Results()
	r.Len(res.Commands, 0)
	r.Len(res.Files, 2)

	f := res.Files[0]
	r.Equal("bar/baz.txt", f.Name())
	r.Equal("baz!", f.String())

	f = res.Files[1]
	r.Equal("foo.txt", f.Name())
	r.Equal("foo!", f.String())
}

func Test_Command(t *testing.T) {
	r := require.New(t)

	g := New()
	g.Command(exec.Command("echo", "hello"))

	run := DryRunner(context.Background())
	run.With(g)
	r.NoError(run.Run())

	res := run.Results()
	r.Len(res.Commands, 1)
	r.Len(res.Files, 0)

	c := res.Commands[0]
	r.Equal("echo hello", strings.Join(c.Args, " "))
}

func Test_Merge(t *testing.T) {
	r := require.New(t)

	g1 := New()
	g1.Root = "one"
	g1.RunFn(func(r *Runner) error {
		return r.File(NewFileS("a.txt", "a"))
	})
	g1.RunFn(func(r *Runner) error {
		return r.File(NewFileS("b.txt", "b"))
	})
	g1.Transformer(NewTransformer("*", func(f File) (File, error) {
		return NewFileS(f.Name(), strings.ToUpper(f.String())), nil
	}))

	g2 := New()
	g2.Root = "two"
	g2.RunFn(func(r *Runner) error {
		return r.File(NewFileS("c.txt", "c"))
	})
	g2.Transformer(NewTransformer("*", func(f File) (File, error) {
		return NewFileS(f.Name(), f.String()+"g2"), nil
	}))

	g1.RunFn(func(r *Runner) error {
		return r.File(NewFileS("d.txt", "d"))
	})
	g1.Transformer(NewTransformer("*", func(f File) (File, error) {
		return NewFileS(f.Name(), f.String()+"g1"), nil
	}))

	g1.Merge(g2)

	run := DryRunner(context.Background())
	run.With(g1)
	r.NoError(run.Run())

	res := run.Results()
	r.Len(res.Files, 4)
}
