//go:build !ladybug

package export

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newGraphCmd() *cobra.Command {
	var outputDir string
	var resultsPaths []string

	cmd := &cobra.Command{
		Use:   "graph <dir>",
		Short: "Export requirements to a LadybugDB graph database for querying",
		Long: `Export requirements to a LadybugDB graph database.

This build of reqmd does not include LadybugDB support.
Rebuild with: go build -tags ladybug`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("graph export requires a build with LadybugDB support; rebuild with: go build -tags ladybug")
		},
	}

	// Register the same flags as the LadybugDB build so a documented command
	// (`export graph <dir> -o <dir> --results <path>`) reaches the helpful
	// build-tag error instead of failing on an unknown flag.
	cmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory for the graph database")
	cmd.Flags().StringArrayVar(&resultsPaths, "results", nil, "Load ephemeral verification results (CTRF or manual) to include result nodes with outcome properties in the graph. Repeatable.")
	return cmd
}
