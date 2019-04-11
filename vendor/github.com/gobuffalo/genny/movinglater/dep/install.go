package dep

import (
	"os/exec"

	"github.com/gobuffalo/genny"
)

// InstallDep is deprecated. Use github.com/gobuffalo/depgen#InstallDep instead.
func InstallDep(args ...string) genny.RunFn {
	return func(r *genny.Runner) error {
		if _, err := r.LookPath("dep"); err == nil {
			return nil
		}

		args = append([]string{"get"}, args...)
		args = append(args, "github.com/golang/dep/cmd/dep")
		c := exec.Command(genny.GoBin(), args...)
		return r.Exec(c)
	}

}
