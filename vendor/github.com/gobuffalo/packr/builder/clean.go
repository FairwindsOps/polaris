package builder

import (
	"os"
	"path/filepath"

	"github.com/gobuffalo/packr/v2/jam/parser"
	"github.com/gobuffalo/packr/v2/jam/store"
)

// Clean up an *-packr.go files
func Clean(root string) error {
	defer func() {
		packd := filepath.Join(root, "packrd")
		os.RemoveAll(packd)
	}()

	p, err := parser.NewFromRoots([]string{root}, &parser.RootsOptions{
		IgnoreImports: true,
	})
	if err != nil {
		return err
	}

	boxes, err := p.Run()
	if err != nil {
		return err
	}

	d := store.NewDisk("", "")
	for _, box := range boxes {
		if err := d.Clean(box); err != nil {
			return err
		}
	}
	return nil
}
