package export

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"reqmd/internal/exporter"
	"reqmd/internal/filter"
	"reqmd/internal/graph"
	"reqmd/internal/parser"
	"reqmd/internal/verify"
)

func newHtmlCmd() *cobra.Command {
	var outputDir string
	var resultsPaths []string
	var filterExpr string

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

			// exportDocs is the set of documents to export. When --filter
			// is active, only matching requirements are exported; the graph
			// is still built from the full docs so trace links resolve.
			exportDocs := docs
			if filterExpr != "" {
				f, err := filter.Compile(filterExpr, filter.BuildValidAttrs(docs))
				if err != nil {
					return err
				}
				exportDocs, err = f.FilterDocs(docs)
				if err != nil {
					return err
				}
			}

			// Load ephemeral verification results when --results is supplied.
			// Results are synthesized into pseudo-requirements appended to
			// the doc slice so the graph builds result→measure edges and
			// outcome-gated checks run. The verdicts map is extracted for
			// rendering badges on measure cards.
			var verdicts map[string]exporter.VerdictInfo
			graphDocs := docs
			if len(resultsPaths) > 0 {
				merged, vVerdicts, _, err := verify.LoadVerdicts(resultsPaths)
				if err != nil {
					return fmt.Errorf("loading results: %w", err)
				}
				resultDoc := verify.Synthesize(merged)
				graphDocs = append(graphDocs, resultDoc)
				verdicts = make(map[string]exporter.VerdictInfo, len(vVerdicts))
				for id, v := range vVerdicts {
					verdicts[id] = exporter.VerdictInfo{Outcome: v.Outcome, Source: v.Source}
				}
			}

			// Build the trace graph for upstream/downstream links
			g, err := graph.New(graphDocs)
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
			exp.SetVerdicts(verdicts)

			// Ensure output directory exists.
			if outputDir != "" {
				if err := os.MkdirAll(outputDir, 0755); err != nil {
					return fmt.Errorf("creating output directory: %w", err)
				}
			}

			for _, doc := range exportDocs {
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
					return fmt.Errorf("exporting %s: %w", outPath, err)
				}
				f.Close()
				fmt.Fprintf(os.Stderr, "Wrote %s\n", outPath)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory for HTML files")
	cmd.Flags().StringArrayVar(&resultsPaths, "results", nil, "Load ephemeral verification results (CTRF .ctrf.json or manual-results dirs) to render verdict badges on measure cards. Repeatable.")
	cmd.Flags().StringVar(&filterExpr, "filter", "", "Filter requirements using an expr-lang expression. Only matching requirements are exported.")
	return cmd
}
