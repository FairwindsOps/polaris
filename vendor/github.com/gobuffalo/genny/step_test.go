package genny

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_NewStep(t *testing.T) {
	r := require.New(t)

	_, err := NewStep(nil, 0)
	r.Error(err)

	s, err := NewStep(New(), 0)
	r.NoError(err)
	r.NotZero(s)
}

func Test_StepBefore(t *testing.T) {
	r := require.New(t)

	var actual []string

	g := New()
	g.RunFn(func(r *Runner) error {
		actual = append(actual, "as")
		return nil
	})

	s, err := NewStep(g, 0)
	r.NoError(err)

	b := New()
	b.RunFn(func(r *Runner) error {
		actual = append(actual, "before")
		return nil
	})
	s.Before(b)

	run := DryRunner(context.Background())

	r.NoError(s.Run(run))

	r.Equal([]string{"before", "as"}, actual)
}

func Test_StepBefore_Delete(t *testing.T) {
	n := 5
	for i := 0; i < n; i++ {
		t.Run("deleting index "+strconv.Itoa(i), func(st *testing.T) {
			r := require.New(st)

			s, err := NewStep(New(), 0)
			r.NoError(err)

			var name string
			for x := 0; x < n; x++ {
				g := New()
				del := s.Before(g)
				if x == i {
					name = g.StepName
					del()
				}
			}
			r.Len(s.before, n-1)
			r.NotZero(name)
			for _, g := range s.before {
				r.NotEqual(name, g.StepName)
			}
		})
	}
}

func Test_StepAfter(t *testing.T) {
	r := require.New(t)

	var actual []string

	g := New()
	g.RunFn(func(r *Runner) error {
		actual = append(actual, "as")
		return nil
	})

	s, err := NewStep(g, 0)
	r.NoError(err)

	b := New()
	b.RunFn(func(r *Runner) error {
		actual = append(actual, "after")
		return nil
	})
	s.After(b)

	run := DryRunner(context.Background())

	r.NoError(s.Run(run))

	r.Equal([]string{"as", "after"}, actual)
}

func Test_StepAfter_Delete(t *testing.T) {
	n := 5
	for i := 0; i < n; i++ {
		t.Run("deleting index "+strconv.Itoa(i), func(st *testing.T) {
			r := require.New(st)

			s, err := NewStep(New(), 0)
			r.NoError(err)

			var name string
			for x := 0; x < n; x++ {
				g := New()
				del := s.After(g)
				if x == i {
					name = g.StepName
					del()
				}
			}
			r.Len(s.after, n-1)
			r.NotZero(name)
			for _, g := range s.after {
				r.NotEqual(name, g.StepName)
			}
		})
	}
}
