package safe

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Run(t *testing.T) {
	r := require.New(t)

	err := Run(func() {
		panic("foo")
	})
	r.Error(err)
	r.Equal("foo", err.Error())

	var x bool
	err = Run(func() {
		x = true
	})
	r.NoError(err)
	r.True(x)
}

func Test_RunE(t *testing.T) {
	r := require.New(t)

	err := RunE(func() error {
		panic("foo")
	})
	r.Error(err)
	r.Equal("foo", err.Error())

	var x bool
	err = RunE(func() error {
		x = true
		return nil
	})
	r.NoError(err)
	r.True(x)

	err = RunE(func() error {
		return errors.New("boom")
	})
	r.Error(err)
	r.Equal("boom", err.Error())
}
