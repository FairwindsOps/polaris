package genny

import (
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_NewFile(t *testing.T) {
	r := require.New(t)

	input := "hi"
	f := NewFile("foo.txt", strings.NewReader(input))
	r.NotNil(f)
	r.Equal("foo.txt", f.Name())
	b, err := ioutil.ReadAll(f)
	r.NoError(err)
	r.Equal(input, string(b))
}

func Test_File_Writer(t *testing.T) {
	r := require.New(t)

	input := "hi"
	f := NewFile("foo.txt", strings.NewReader(input))
	r.Equal(input, f.String())
	i, err := io.Copy(f, strings.NewReader("HELLO"))
	r.NoError(err)
	r.Equal(int64(5), i)

	r.Equal("HELLO", f.String())
}
