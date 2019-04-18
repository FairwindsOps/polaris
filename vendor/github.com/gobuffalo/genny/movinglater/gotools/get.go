package gotools

import (
	"github.com/gobuffalo/genny"
	"github.com/gobuffalo/gogen"
)

// Get is deprecated. Use github.com/gobuffalo/gogen#Get instead.
func Get(pkg string, args ...string) genny.RunFn {
	return func(r *genny.Runner) error {
		return r.Exec(gogen.Get(pkg, args...))
	}
}
