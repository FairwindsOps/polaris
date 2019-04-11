package new

import (
	"path/filepath"

	"github.com/gobuffalo/flect/name"
	"github.com/gobuffalo/genny"
	"github.com/gobuffalo/gogen"
	"github.com/gobuffalo/packr/v2"
)

func New(opts *Options) (*genny.Generator, error) {
	g := genny.New()

	if err := opts.Validate(); err != nil {
		return g, err
	}

	if err := g.Box(packr.New("github.com/gobuffalo/genny/genny/new", "../new/templates")); err != nil {
		return g, err
	}
	name := name.New(opts.Name)

	ctx := map[string]interface{}{
		"name":    name,
		"boxName": opts.BoxName,
	}
	g.Transformer(gogen.TemplateTransformer(ctx, nil))
	g.Transformer(genny.Replace("-name-", name.File().String()))
	g.Transformer(genny.Dot())
	g.Transformer(genny.NewTransformer("*", func(f genny.File) (genny.File, error) {
		f = genny.NewFile(filepath.Join(opts.Prefix, f.Name()), f)
		return f, nil
	}))
	return g, nil
}
