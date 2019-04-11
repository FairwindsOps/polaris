package gogen

import (
	"strings"
	"testing"

	"github.com/gobuffalo/genny"
	"github.com/stretchr/testify/require"
)

func Test_RewriteImports(t *testing.T) {
	r := require.New(t)

	f := genny.NewFile("main.go", strings.NewReader(cItmpl))

	f, err := RewriteImports(f, map[string]string{
		"1/2": "2/1",
		"3/4": "4/3",
	})
	r.NoError(err)
	r.Equal(cIexp, f.String())
}

const cIexp = `package main

import (
	"2/1"
	"4/3"
	"fmt"
	"log"
)

func main() {}
`

const cItmpl = `package main

import (
	"fmt"
	"log"
	"1/2"
	"3/4"
)

func main() {}
`
