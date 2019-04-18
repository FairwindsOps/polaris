package main

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/gobuffalo/genny"
	"github.com/gobuffalo/genny/internal/makers/simple"
)

func main() {
	run := genny.DryRunner(context.Background())
	run.FileFn = func(f genny.File) (genny.File, error) {
		io.Copy(os.Stdout, f)
		return f, nil
	}

	g := simple.New()
	run.With(g)

	if err := run.Run(); err != nil {
		log.Fatal(err)
	}
}
