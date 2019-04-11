package gentest

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_CompareCommands(t *testing.T) {
	r := require.New(t)

	r.NoError(CompareCommands([]string{"go version"}, []*exec.Cmd{exec.Command("go", "version")}))
	r.Error(CompareCommands([]string{"go version"}, []*exec.Cmd{exec.Command("go", "version"), exec.Command("go", "env")}))
}
