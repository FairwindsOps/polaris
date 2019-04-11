package genny

import (
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_WetRunner(t *testing.T) {
	r := require.New(t)

	dir, err := ioutil.TempDir("", "")
	r.NoError(err)

	run := WetRunner(context.Background())
	run.Root = dir

	g := New()
	g.Command(exec.Command("echo", "hello"))
	g.File(NewFile("foo.txt", strings.NewReader("foo!")))
	dp := filepath.Join("a", "b", "c")
	g.File(NewDir(dp, 0755))
	run.With(g)

	r.NoError(run.Run())

	res := run.Results()
	r.Len(res.Commands, 1)

	c := res.Commands[0]
	r.Equal("echo hello", strings.Join(c.Args, " "))

	expected := []string{"a/b/c", "foo.txt"}

	for i, f := range res.Files {
		r.Equal(expected[i], f.Name())
	}

	r.Len(res.Files, len(expected))

	f, err := res.Find("foo.txt")
	r.NoError(err)
	r.Equal("foo.txt", f.Name())
	r.Equal("foo!", f.String())

	b, err := ioutil.ReadFile(filepath.Join(run.Root, "foo.txt"))
	r.NoError(err)
	r.Equal("foo!", string(b))

	_, err = res.Find("a/b/c")
	r.NoError(err)

	_, err = os.Stat(filepath.Join(run.Root, dp))
	r.NoError(err)
}
