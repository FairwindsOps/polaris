package packr_test

import (
	"testing"

	"github.com/gobuffalo/packr/v2/_fixtures/import_pkg"
	"github.com/stretchr/testify/require"
)

func Test_ImportWithBox(t *testing.T) {
	r := require.New(t)

	r.Len(import_pkg.BoxTestNew.List(), 2)

	r.Len(import_pkg.BoxTestNewBox.List(), 2)
}
