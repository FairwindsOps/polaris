package import_pkg

import (
	"testing"

	"github.com/gobuffalo/packr/v2"
	"github.com/stretchr/testify/require"
)

func Test_NewBox(t *testing.T) {
	r := require.New(t)

	box := packr.NewBox("./pkg_test")
	r.Len(box.List(), 2)
}

func Test_New(t *testing.T) {
	r := require.New(t)

	box := packr.New("pkg_test", "./pkg_test")
	r.Len(box.List(), 2)
}
