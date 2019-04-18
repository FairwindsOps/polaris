package gentest

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gobuffalo/genny"
	"github.com/gobuffalo/packd"
)

// CompareFiles compares a slice of expected filenames to a slice of
// genny.Files. Expected files can be listed in any order. An excellent choice
// for the actual files can be found in genny#Generator.Results under the Files
// attribute
func CompareFiles(exp []string, act []genny.File) error {
	if len(exp) != len(act) {
		return fmt.Errorf("len(exp) != len(act) [%d != %d]", len(exp), len(act))
	}

	var acts []string
	for _, f := range act {
		acts = append(acts, f.Name())
	}
	sort.Strings(exp)
	sort.Strings(acts)

	for i, n := range exp {
		if n != acts[i] {
			return fmt.Errorf("expected %q to match %q", acts, exp)
		}
	}
	return nil
}

// CompareBox compares a packd.Walkable box of files (usually fixtures)
// the results of a genny.Runner
func CompareBox(exp packd.Walkable, res genny.Results) error {
	return exp.Walk(func(path string, file packd.File) error {
		if filepath.Base(path) == ".DS_Store" {
			return nil
		}
		f, err := res.Find(path)
		if err != nil {
			return err
		}
		if file.String() != f.String() {
			return fmt.Errorf("[%s] expected %s to match %s", path, file, f)
		}
		return nil
	})
}

// CompareBoxStripped compares a packd.Walkable box of files (usually fixtures)
// the results of a genny.Runner by removing any whitespaces, tabs, or newlines.
func CompareBoxStripped(exp packd.Walkable, res genny.Results) error {
	return exp.Walk(func(path string, file packd.File) error {
		if filepath.Base(path) == ".DS_Store" {
			return nil
		}
		f, err := res.Find(path)
		if err != nil {
			return err
		}
		if clean(file.String()) != clean(f.String()) {
			return fmt.Errorf("[%s] expected %s to match %s", path, file, f)
		}
		return nil
	})
}

func clean(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Replace(s, "\n", "", -1)
	s = strings.Replace(s, "\r", "", -1)
	s = strings.Replace(s, "\t", "", -1)
	return s
}
