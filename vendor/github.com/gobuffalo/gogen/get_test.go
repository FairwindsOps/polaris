package gogen

import (
	"context"
	"strings"
	"testing"

	"github.com/gobuffalo/genny"
	"github.com/stretchr/testify/require"
)

func Test_Get(t *testing.T) {
	r := require.New(t)

	run := genny.DryRunner(context.Background())

	get := Get("github.com/gobuffalo/buffalo", "-v", "-u")
	r.NoError(run.Exec(get))

	res := run.Results()
	r.Len(res.Commands, 1)

	c := res.Commands[0]
	r.Equal(genny.GoBin()+" get -v -u github.com/gobuffalo/buffalo", strings.Join(c.Args, " "))

	r.Len(res.Files, 0)

}
