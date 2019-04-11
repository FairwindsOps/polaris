package gogen

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/gobuffalo/genny"
	"github.com/stretchr/testify/require"
)

func Test_TemplateTransformer(t *testing.T) {
	r := require.New(t)

	f := genny.NewFile("foo.tmpl.txt", strings.NewReader("Hello {{.}}"))

	tr := TemplateTransformer("mark", nil)
	f, err := tr.Transform(f)
	r.NoError(err)
	r.Equal("foo.txt", f.Name())

	b, err := ioutil.ReadAll(f)
	r.NoError(err)
	r.Equal("Hello mark", string(b))
}
