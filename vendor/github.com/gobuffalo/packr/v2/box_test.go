package packr

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gobuffalo/packr/v2/file"
	"github.com/gobuffalo/packr/v2/file/resolver"
	"github.com/stretchr/testify/require"
)

func Test_New(t *testing.T) {
	r := require.New(t)

	box := New("Test_NewBox", filepath.Join("_fixtures", "list_test"))
	r.Len(box.List(), 4)

}
func Test_Box_AddString(t *testing.T) {
	r := require.New(t)

	box := New("Test_Box_AddString", "./templates")
	s, err := box.FindString("foo.txt")
	r.Error(err)
	r.Equal("", s)

	r.NoError(box.AddString("foo.txt", "foo!!"))
	s, err = box.FindString("foo.txt")
	r.NoError(err)
	r.Equal("foo!!", s)
}

func Test_Box_AddBytes(t *testing.T) {
	r := require.New(t)

	box := New("Test_Box_AddBytes", "")
	s, err := box.FindString("foo.txt")
	r.Error(err)
	r.Equal("", s)

	r.NoError(box.AddBytes("foo.txt", []byte("foo!!")))
	s, err = box.FindString("foo.txt")
	r.NoError(err)
	r.Equal("foo!!", s)
}

func Test_Box_String(t *testing.T) {
	r := require.New(t)

	box := New("Test_Box_String", "./templates")
	d := resolver.NewInMemory(map[string]file.File{
		"foo.txt": qfile("foo.txt", "foo!"),
	})
	box.SetResolver("foo.txt", d)

	s := box.String("foo.txt")
	r.Equal("foo!", s)

	s = box.String("idontexist")
	r.Equal("", s)
}

func Test_Box_String_Miss(t *testing.T) {
	r := require.New(t)

	box := New("Test_Box_String_Miss", filepath.Join("_fixtures", "templates"))

	s := box.String("foo.txt")
	r.Equal("FOO!!!", strings.TrimSpace(s))

	s = box.String("idontexist")
	r.Equal("", s)
}

func Test_Box_FindString(t *testing.T) {
	r := require.New(t)

	box := New("Test_Box_FindString", "./templates")
	d := resolver.NewInMemory(map[string]file.File{
		"foo.txt": qfile("foo.txt", "foo!"),
	})
	box.SetResolver("foo.txt", d)

	s, err := box.FindString("foo.txt")
	r.NoError(err)
	r.Equal("foo!", s)

	s, err = box.FindString("idontexist")
	r.Error(err)
	r.Equal("", s)
}

func Test_Box_FindString_Miss(t *testing.T) {
	r := require.New(t)

	box := New("Test_Box_FindString_Miss", filepath.Join("_fixtures", "templates"))

	s, err := box.FindString("foo.txt")
	r.NoError(err)
	r.Equal("FOO!!!", strings.TrimSpace(s))

	s, err = box.FindString("idontexist")
	r.Error(err)
	r.Equal("", s)
}

func Test_Box_Bytes(t *testing.T) {
	r := require.New(t)

	box := New("Test_Box_Bytes", "./templates")
	d := resolver.NewInMemory(map[string]file.File{
		"foo.txt": qfile("foo.txt", "foo!"),
	})
	box.SetResolver("foo.txt", d)

	s := box.Bytes("foo.txt")
	r.Equal([]byte("foo!"), s)

	s = box.Bytes("idontexist")
	r.Equal([]byte(""), s)
}

func Test_Box_Bytes_Miss(t *testing.T) {
	r := require.New(t)

	box := New("Test_Box_Bytes_Miss", filepath.Join("_fixtures", "templates"))

	s := box.Bytes("foo.txt")
	r.Equal([]byte("FOO!!!"), bytes.TrimSpace(s))

	s = box.Bytes("idontexist")
	r.Equal([]byte(""), s)
}

func Test_Box_Find(t *testing.T) {
	r := require.New(t)

	box := New("Test_Box_Find", "./templates")
	d := resolver.NewInMemory(map[string]file.File{
		"foo.txt": qfile("foo.txt", "foo!"),
	})
	box.SetResolver("foo.txt", d)

	s, err := box.Find("foo.txt")
	r.NoError(err)
	r.Equal("foo!", string(s))

	s, err = box.Find("idontexist")
	r.Error(err)
	r.Equal("", string(s))
}

func Test_Box_Find_Miss(t *testing.T) {
	r := require.New(t)

	box := New("Test_Box_Find_Miss", "./_fixtures/templates")
	s, err := box.Find("foo.txt")
	r.NoError(err)
	r.Equal("FOO!!!", strings.TrimSpace(string(s)))

	s, err = box.Find("idontexist")
	r.Error(err)
	r.Equal("", string(s))
}

func Test_Box_Has(t *testing.T) {
	r := require.New(t)

	box := New("Test_Box_Has", "./templates")
	d := resolver.NewInMemory(map[string]file.File{
		"foo.txt": qfile("foo.txt", "foo!"),
	})
	box.SetResolver("foo.txt", d)

	r.True(box.Has("foo.txt"))
	r.False(box.Has("idontexist"))
}

func Test_Box_Open(t *testing.T) {
	r := require.New(t)

	d := resolver.NewInMemory(map[string]file.File{
		"foo.txt":        qfile("foo.txt", "foo!"),
		"bar":            qfile("bar", "bar!"),
		"baz/index.html": qfile("baz", "baz!"),
	})
	box := New("Test_Box_Open", "./templates")

	box.DefaultResolver = d

	for _, x := range []string{"foo.txt", "/foo.txt", "bar", "/bar", "baz", "/baz"} {
		f, err := box.Open(x)
		r.NoError(err)
		r.NotZero(f)
	}

	f, err := box.Open("idontexist.txt")
	r.Error(err)
	r.Zero(f)
}

func Test_Box_List(t *testing.T) {
	r := require.New(t)

	box := New("Test_Box_List", filepath.Join("_fixtures", "list_test"))
	r.NoError(box.AddString(filepath.Join("d", "d.txt"), "D"))

	act := box.List()
	exp := []string{"a.txt", filepath.Join("b", "b.txt"), filepath.Join("b", "b2.txt"), filepath.Join("c", "c.txt"), filepath.Join("d", "d.txt")}
	r.Equal(exp, act)
}

func Test_Box_HasDir(t *testing.T) {
	r := require.New(t)

	box := New("Test_Box_HasDir", filepath.Join("_fixtures", "list_test"))
	r.NoError(box.AddString("d/e/f.txt", "D"))

	r.True(box.HasDir("d/e"))
	r.True(box.HasDir("d"))
	r.True(box.HasDir("c"))
	r.False(box.HasDir("a"))
}
