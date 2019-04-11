package genny_test

import (
	"context"
	"fmt"
	"go/build"
	"log"
	"os/exec"
	"strings"

	"github.com/gobuffalo/genny"
	"github.com/gobuffalo/genny/gentest"
)

// exampleLogger just cleans up variable log content
// such as GOPATH, step names, etc....
// without this Go Example tests won't work.
func exampleLogger(l *gentest.Logger) genny.Logger {
	l.CloseFn = func() error {
		s := l.Stream.String()
		c := build.Default
		for _, src := range c.SrcDirs() {
			s = strings.Replace(s, src, "/go/src", -1)
		}
		s = strings.Replace(s, "\\", "/", -1)

		for i, line := range strings.Split(s, "\n") {
			if strings.Contains(line, "Step:") {
				s = strings.Replace(s, line, fmt.Sprintf("[DEBU] Step: %d", i+1), 1)
			}
		}
		fmt.Print(s)
		return nil
	}
	return l
}

func ExampleGenerator_withCommand() {
	// create a new `*genny.Generator`
	g := genny.New()

	g.Command(exec.Command("go", "version"))

	// create a new `*genny.Runner`
	r := genny.NewRunner(context.Background())

	// add a new logger to clean and dump output
	// for the example tests
	r.Logger = exampleLogger(gentest.NewLogger())

	// add the generator to the `*genny.Runner`.
	r.With(g)

	// run the runner
	if err := r.Run(); err != nil {
		log.Fatal(err)
	}
	// Output:
	// [DEBU] Step: 1
	// [DEBU] Chdir: /go/src/github.com/gobuffalo/genny
	// [DEBU] Exec: go version
}

func ExampleGenerator_withFile() {
	// create a new `*genny.Generator`
	g := genny.New()

	// add a file named `index.html` that has a body of `Hello\n`
	// to the generator
	g.File(genny.NewFileS("index.html", "Hello\n"))

	// create a new `*genny.Runner`
	r := genny.NewRunner(context.Background())

	// add a new logger to clean and dump output
	// for the example tests
	r.Logger = exampleLogger(gentest.NewLogger())

	// add the generator to the `*genny.Runner`.
	r.With(g)

	// run the runner
	if err := r.Run(); err != nil {
		log.Fatal(err)
	}
	// Output:
	// [DEBU] Step: 1
	// [DEBU] Chdir: /go/src/github.com/gobuffalo/genny
	// [DEBU] File: /go/src/github.com/gobuffalo/genny/index.html
}

func ExampleRunner() {
	// create a new `*genny.Runner`
	r := genny.NewRunner(context.Background())

	// add a new logger to clean and dump output
	// for the example tests
	r.Logger = exampleLogger(gentest.NewLogger())

	// add the generator(s) to the `*genny.Runner`.
	// r.With(g)

	// run the runner
	if err := r.Run(); err != nil {
		log.Fatal(err)
	}
}
