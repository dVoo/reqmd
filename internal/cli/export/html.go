package export

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"reqmd/internal/exporter"
	"reqmd/internal/graph"
	"reqmd/internal/parser"
)

func newHtmlCmd() *cobra.Command {
	var outputDir string

	cmd := &cobra.Command{
		Use:   "html <dir>",
		Short: "Export requirements to HTML format with trace links",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := args[0]

			docs, err := parser.Discover(root)
			if err != nil {
				return fmt.Errorf("discovering documents: %w", err)
			}

			// Build the trace graph for upstream/downstream links
			g, err := graph.New(docs)
			if err != nil {
				return fmt.Errorf("building trace graph: %w", err)
			}

			titleMap := exporter.BuildTitleMap(docs)

			// Build render context for doc chain and cross-file links
			rctx, err := exporter.NewRenderContext(docs, root)
			if err != nil {
				return fmt.Errorf("building render context: %w", err)
			}

			boundaries := exporter.ComputeDocBoundaries(docs)

			var exp exporter.HTML

			// Ensure output directory exists.
			if outputDir != "" {
				if err := os.MkdirAll(outputDir, 0755); err != nil {
					return fmt.Errorf("creating output directory: %w", err)
				}
			}

			for _, doc := range docs {
				props := doc.Properties
				dirName := filepath.Base(doc.Path)
				outPath := dirName + "-requirements.html"

				if outputDir != "" {
					outPath = filepath.Join(outputDir, outPath)
				} else {
					outPath = filepath.Join(doc.Path, outPath)
				}

				exp.SetBoundary(boundaries[doc.Path])

				// Build doc-level Confluence-Flow trace graph and rewrite card
				// paths to be relative to the output file we're about to write.
				graph := rctx.BuildChainGraph(doc)
				exporter.RelativizeChainGraph(&graph, filepath.Dir(outPath))
				exp.SetDocChainGraph(graph)

				f, err := os.Create(outPath)
				if err != nil {
					return fmt.Errorf("creating %s: %w", outPath, err)
				}

				tc := exporter.TraceResolver{
					Upstream:    g.UpstreamNeighbors,
					Downstream:  g.DownstreamNeighbors,
					ResolveLink: rctx.ResolveLink(outPath),
					TitleOf:     func(id string) string { return titleMap[id] },
				}
				err = exp.ExportWithTraces(f, doc, props, tc)
				if err != nil {
					f.Close()
					return fmt.Errorf("exporting %s: %w", doc.Path, err)
				}
				f.Close()
				fmt.Fprintf(os.Stderr, "Wrote %s\n", outPath)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory for HTML files")
	return cmd
}
