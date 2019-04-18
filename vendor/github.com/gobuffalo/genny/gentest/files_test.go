package gentest

import (
	"testing"

	"github.com/gobuffalo/genny"
	"github.com/gobuffalo/packd"
	"github.com/stretchr/testify/require"
)

func Test_CompareFiles(t *testing.T) {
	r := require.New(t)

	exp := []string{"b.html", "a.html"}
	act := []genny.File{
		genny.NewFileS("a.html", "A"),
		genny.NewFileS("b.html", "B"),
	}
	r.NoError(CompareFiles(exp, act))
}

func Test_CompareBox(t *testing.T) {
	r := require.New(t)

	box := packd.NewMemoryBox()
	box.AddString("a.html", "A")
	box.AddString("b.html", "B")

	res := genny.Results{
		Files: []genny.File{
			genny.NewFileS("a.html", "A"),
			genny.NewFileS("b.html", "B"),
		},
	}

	r.NoError(CompareBox(box, res))
}

func Test_CompareBox_Missing(t *testing.T) {
	r := require.New(t)

	box := packd.NewMemoryBox()
	box.AddString("a.html", "A")
	box.AddString("b.html", "B")

	res := genny.Results{
		Files: []genny.File{
			genny.NewFileS("b.html", "b"),
		},
	}

	r.Error(CompareBox(box, res))
}

func Test_CompareBox_Stripped(t *testing.T) {
	r := require.New(t)

	box := packd.NewMemoryBox()
	box.AddString("a.html", "A\nx")
	box.AddString("b.html", "B")

	res := genny.Results{
		Files: []genny.File{
			genny.NewFileS("a.html", "    A\n\rx"),
			genny.NewFileS("b.html", "B"),
		},
	}

	r.NoError(CompareBoxStripped(box, res))
}
