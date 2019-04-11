package new

import (
	"testing"

	"github.com/gobuffalo/gogen/gomods"
	"github.com/stretchr/testify/require"
)

func Test_Options(t *testing.T) {
	gomods.Disable(func() error {
		r := require.New(t)

		opts := &Options{}
		r.Error(opts.Validate())

		opts.Name = "foo"
		r.NoError(opts.Validate())

		r.Equal("github.com/gobuffalo/genny/genny/new/foo/templates", opts.BoxName)
		return nil
	})
}
