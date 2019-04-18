package gentest

import (
	"fmt"
	"os/exec"
	"strings"
)

// CompareCommands asserts that the expected commands match the actual commands
// executed. The expected commands should be the commands, with arguments, as
// you would type them on the command line. In typical usage, the actual
// commands can be extracted from a genny#Generator.Results under the Commands
// key
func CompareCommands(exp []string, act []*exec.Cmd) error {
	if len(exp) != len(act) {
		return fmt.Errorf("len(exp) != len(act) [%d != %d]", len(exp), len(act))
	}
	for i, c := range act {
		e := exp[i]
		a := strings.Join(c.Args, " ")
		if a != e {
			return fmt.Errorf("expect %q got %q", e, a)
		}
	}
	return nil
}
