package genny

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_exts(t *testing.T) {
	r := require.New(t)

	e := exts(NewFile("foo.bar.baz.biz", nil))
	r.Len(e, 3)
	r.Equal([]string{".bar", ".baz", ".biz"}, e)
}

func Test_StripExt(t *testing.T) {
	r := require.New(t)

	f := NewFile("foo.bar.baz.biz", nil)
	f = StripExt(f, ".bar")
	r.Equal("foo.baz.biz", f.Name())
}

func Test_HasExt(t *testing.T) {
	r := require.New(t)

	f := NewFile("foo.bar.baz.biz", nil)
	r.True(HasExt(f, ".bar"))
	r.True(HasExt(f, ".baz"))
	r.True(HasExt(f, ".boz", ".baz", ".boz"))
	r.True(HasExt(f))
	r.False(HasExt(f, ".boz"))
	r.False(HasExt(f, ".boz", ".bzo"))
}
