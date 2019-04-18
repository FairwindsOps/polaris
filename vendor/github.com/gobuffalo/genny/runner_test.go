package genny

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Runner_Run(t *testing.T) {
	r := require.New(t)

	g := New()

	g.Command(exec.Command("foo", "bar"))
	g.File(NewFile("foo.txt", strings.NewReader("Hello mark")))

	run := DryRunner(context.Background())
	run.With(g)

	r.NoError(run.Run())

	res := run.Results()
	r.Len(res.Commands, 1)
	r.Len(res.Files, 1)

	c := res.Commands[0]
	r.Equal("foo bar", strings.Join(c.Args, " "))

	f := res.Files[0]
	r.Equal("foo.txt", f.Name())
	r.Equal("Hello mark", f.String())
}

func Test_Runner_FindFile(t *testing.T) {
	r := require.New(t)

	g := New()
	g.File(NewFile("foo.txt", strings.NewReader("Hello mark")))
	g.File(NewFile("foo.txt", strings.NewReader("Hello world")))

	run := DryRunner(context.Background())
	run.With(g)
	r.NoError(run.Run())

	res := run.Results()
	r.Len(res.Files, 1)

	f, err := run.FindFile("foo.txt")
	r.NoError(err)
	r.Equal(res.Files[0], f)
}

func Test_Runner_FindFile_FromDisk(t *testing.T) {
	r := require.New(t)

	run := DryRunner(context.Background())

	exp, err := ioutil.ReadFile("./fixtures/foo.txt")
	r.NoError(err)

	f, err := run.FindFile("fixtures/foo.txt")
	r.NoError(err)
	act, err := ioutil.ReadAll(f)
	r.NoError(err)

	r.Equal(string(exp), string(act))
}

func Test_Runner_FindFile_DoesntExist(t *testing.T) {
	r := require.New(t)

	run := DryRunner(context.Background())

	_, err := run.FindFile("idontexist")
	r.Error(err)
}

func Test_Runner_Request(t *testing.T) {
	table := []struct {
		code   int
		method string
		path   string
		boom   bool
	}{
		{200, "GET", "/a", false},
		{200, "POST", "/b", false},
		{399, "PATCH", "/c", false},
		{401, "GET", "/d", true},
		{500, "POST", "/e", true},
	}

	for _, tt := range table {
		t.Run(tt.method+tt.path, func(st *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
				res.WriteHeader(tt.code)
			}))
			defer ts.Close()

			r := require.New(st)
			p := ts.URL + tt.path
			req, err := http.NewRequest(tt.method, p, nil)
			r.NoError(err)

			run := WetRunner(context.Background())

			res, err := run.Request(req)
			if tt.boom {
				r.Error(err)
			} else {
				r.NoError(err)
			}

			results := run.Results()
			r.Len(results.Requests, 1)

			rr := results.Requests[0]
			r.Equal(tt.path, rr.Request.URL.Path)

			if res != nil {
				r.Equal(tt.code, res.StatusCode)
			}

		})
	}
}

func Test_Runner_FindStep(t *testing.T) {
	r := require.New(t)

	run := DryRunner(context.Background())

	for i := 0; i < 3; i++ {
		s, err := NewStep(New(), i)
		r.NoError(err)
		run.WithStep("step "+strconv.Itoa(i+1), s)
	}

	s, err := run.FindStep("step 2")
	r.NoError(err)
	r.NotZero(s)
}

func Test_Runner_ReplaceStep(t *testing.T) {
	r := require.New(t)

	run := DryRunner(context.Background())

	for i := 0; i < 3; i++ {
		g := New()
		g.File(NewFileS(fmt.Sprintf("%d.txt", i), strconv.Itoa(i)))
		s, err := NewStep(g, i)
		r.NoError(err)
		run.WithStep("step "+strconv.Itoa(i), s)
	}

	gx := New()
	gx.File(NewFileS("2.txt", "replaced"))
	s, err := NewStep(gx, 0)
	r.NoError(err)
	err = run.ReplaceStep("step 2", s)
	r.NoError(err)

	r.NoError(run.Run())
	res := run.Results()
	f, err := res.Find("2.txt")
	r.NoError(err)
	r.Equal("replaced", f.String())
}
