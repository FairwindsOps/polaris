package simple

import (
	"fmt"
	"os/exec"

	"github.com/gobuffalo/genny"
	"github.com/gobuffalo/gogen"
)

// New ...
func New() *genny.Generator {
	g := genny.New()

	// add a file
	g.File(genny.NewFileS("index.html", "Hello\n"))

	// execute a command
	g.Command(exec.Command("go", "env"))

	// run a function at run time
	g.RunFn(func(r *genny.Runner) error {
		// look for the `genny` executable
		if _, err := r.LookPath("genny"); err != nil {
			// it wasn't found, so install it
			c := gogen.Get("github.com/gobuffalo/genny/genny")
			if err := r.Exec(c); err != nil {
				return err
			}
		}
		// call the `genny` executable with the `-h` flag.
		return r.Exec(exec.Command("genny", "-h"))
	})

	g.RunFn(func(r *genny.Runner) error {
		// try to find main.go either in the virtual "disk"
		// or the physical one
		f, err := r.Disk.Find("main.go")
		if err != nil {
			return err
		}
		// print the contents of the file
		fmt.Println(f.String())
		return nil
	})

	return g
}
