//go:build ladybug

package export

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"reqmd/internal/exporter"
	"reqmd/internal/parser"
)

func newGraphCmd() *cobra.Command {
	var outputDir string

	cmd := &cobra.Command{
		Use:   "graph <dir>",
		Short: "Export requirements to a LadybugDB graph database for querying",
		Long: `Export requirements to a LadybugDB graph database.

Creates a LadybugDB-backed graph database at the specified output directory.
The database contains one node per requirement and TracesTo edges for every
trace link.

The resulting database can be browsed or queried with the LadybugDB CLI:

  lbug <outDir>/reqmd-graph.lbug`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := args[0]

			docs, err := parser.Discover(root)
			if err != nil {
				return fmt.Errorf("discovering documents: %w", err)
			}

			outPath := outputDir
			if outPath == "" {
				outPath = filepath.Join(root, "reqmd-graph")
			}

			if err := exporter.ExportGraph(docs, outPath); err != nil {
				return fmt.Errorf("exporting graph: %w", err)
			}

			fmt.Fprintf(os.Stderr, "Graph exported to %s\n", outPath)
			fmt.Fprintf(os.Stderr, "Query with: lbug %s\n", filepath.Join(outPath, "reqmd-graph.lbug"))
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory for the graph database")
	return cmd
}
