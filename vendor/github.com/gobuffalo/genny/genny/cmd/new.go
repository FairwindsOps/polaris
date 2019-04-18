package cmd

import (
	"context"

	"github.com/gobuffalo/genny"
	"github.com/gobuffalo/genny/genny/new"
	"github.com/spf13/cobra"
)

var newOptions = struct {
	*new.Options
	dryRun bool
}{
	Options: &new.Options{},
}

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "generates a new genny stub",
	RunE: func(cmd *cobra.Command, args []string) error {
		r := genny.WetRunner(context.Background())

		if newOptions.dryRun {
			r = genny.DryRunner(context.Background())
		}

		opts := newOptions.Options
		var name string
		if len(args) > 0 {
			name = args[0]
		}
		opts.Name = name
		g, err := new.New(opts)
		if err != nil {
			return err
		}

		r.With(g)
		return r.Run()
	},
}

func init() {
	newCmd.Flags().BoolVarP(&newOptions.dryRun, "dry-run", "d", false, "run the generator without creating files or running commands")
	newCmd.Flags().StringVarP(&newOptions.Prefix, "prefix", "p", "", "path prefix for the generator")
	rootCmd.AddCommand(newCmd)
}
