package gomods

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Force(t *testing.T) {
	r := require.New(t)

	Force(true)
	r.True(On())
	Force(false)
	r.False(On())
}
