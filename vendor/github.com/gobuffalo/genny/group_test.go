package genny

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Group_Add(t *testing.T) {
	r := require.New(t)

	gg := &Group{}
	gg.Add(&Generator{})
	gg.Add(&Generator{})

	r.Len(gg.Generators, 2)
}

func Test_Group_With(t *testing.T) {
	r := require.New(t)

	gg := &Group{}
	gg.Add(&Generator{StepName: "one"})
	gg.Add(&Generator{StepName: "two"})

	r.Len(gg.Generators, 2)

	run := DryRunner(context.Background())
	r.Len(run.Steps(), 0)
	gg.With(run)
	r.Len(run.Steps(), 2)
}
