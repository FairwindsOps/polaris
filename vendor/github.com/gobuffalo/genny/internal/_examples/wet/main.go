package main

import (
	"context"
	"log"

	"github.com/gobuffalo/genny"
	"github.com/gobuffalo/genny/internal/makers/simple"
)

func main() {
	run := genny.WetRunner(context.Background())

	g := simple.New()
	run.With(g)

	if err := run.Run(); err != nil {
		log.Fatal(err)
	}
}
