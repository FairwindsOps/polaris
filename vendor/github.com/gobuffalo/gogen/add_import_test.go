package gogen

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gobuffalo/genny"
	"github.com/stretchr/testify/require"
)

func Test_AddImport(t *testing.T) {
	r := require.New(t)

	path := filepath.Join("actions", "app.go")
	f := genny.NewFile(path, strings.NewReader(importBefore))

	f, err := AddImport(f, "foo/bar", "foo/baz")
	r.NoError(err)

	b, err := ioutil.ReadAll(f)
	r.NoError(err)

	r.Equal(importAfter, string(b))
}

const importBefore = `package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
)
`

const importAfter = `package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"


	"foo/bar"
	"foo/baz"
)
`
