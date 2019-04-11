package genny

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Force_Exists(t *testing.T) {
	r := require.New(t)

	dir, err := ioutil.TempDir("", "test")
	r.NoError(err)
	f, err := os.Create(filepath.Join(dir, "foo.txt"))
	r.NoError(err)
	fmt.Fprint(f, "asd")
	f.Close()

	run := DryRunner(context.Background())
	run.WithRun(Force(dir, false))
	r.Error(run.Run())

	run = DryRunner(context.Background())
	run.WithRun(Force(dir, true))
	r.NoError(run.Run())
}

func Test_Force_Doesnt_Exists(t *testing.T) {
	r := require.New(t)

	dir := "i don't exist"

	run := DryRunner(context.Background())
	run.WithRun(Force(dir, false))
	r.NoError(run.Run())

	run = DryRunner(context.Background())
	run.WithRun(Force(dir, true))
	r.NoError(run.Run())
}
