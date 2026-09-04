package export

import (
	"fmt"
	"os"
	"path/filepath"
	"reqmd/internal/exporter"
	"reqmd/internal/filter"
	"reqmd/internal/graph"
	"reqmd/internal/parser"
	"reqmd/internal/verify"

	"github.com/spf13/cobra"
)

func newHTMLCmd() *cobra.Command {
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
				var f *filter.Filter
				f, err = filter.CompileForDocs(docs, filterExpr)
				if err != nil {
					return fmt.Errorf("compiling filter %q: %w", filterExpr, err)
				}
				exportDocs, err = f.FilterDocs(docs)
				if err != nil {
					return fmt.Errorf("applying filter %q: %w", filterExpr, err)
				}
			}

			// Load ephemeral verification results when --results is supplied.
			// Results are synthesized into pseudo-requirements appended to
			// the doc slice so the graph builds result→measure edges and
			// outcome-gated checks run. Badge verdicts are derived from the
			// graph after construction (single source of truth) rather than
			// from the raw merged map.
			var hasResults bool
			graphDocs := docs
			if len(resultsPaths) > 0 {
				merged, _, err := verify.LoadMerged(resultsPaths)
				if err != nil {
					return fmt.Errorf("loading results: %w", err)
				}
				resultDoc := verify.Synthesize(merged)
				graphDocs = append(graphDocs, resultDoc)
				hasResults = true
			}

			// Build the trace graph for upstream/downstream links
			g, err := graph.New(graphDocs)
			if err != nil {
				return fmt.Errorf("building trace graph: %w", err)
			}

			var verdicts map[string]exporter.VerdictInfo
			if hasResults {
				verdicts = make(map[string]exporter.VerdictInfo)
				for id, v := range g.MeasureVerdicts() {
					verdicts[id] = exporter.VerdictInfoFromGraph(v)
				}
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
				if err := os.MkdirAll(outputDir, 0o755); err != nil {
					return fmt.Errorf("creating output directory: %w", err)
				}
			}

			// outPathFor mirrors the file-writing logic below: when
			// outputDir is set every page is written flat into it, otherwise
			// each page goes into its own document directory.
			outPathFor := func(docPath string) string {
				name := filepath.Base(docPath) + "-requirements.html"
				if outputDir != "" {
					return filepath.Join(outputDir, name)
				}
				return filepath.Join(docPath, name)
			}
			outPathByDir := make(map[string]string, len(docs))
			for _, d := range docs {
				outPathByDir[filepath.Base(d.Path)] = outPathFor(d.Path)
			}

			for _, doc := range exportDocs {
				props := doc.Properties
				outPath := outPathFor(doc.Path)

				exp.SetBoundary(boundaries[doc.Path])

				// Build doc-level Confluence-Flow trace graph, point each
				// card at the output file that doc is written to, then
				// relativize to this page's own directory.
				graph := rctx.BuildChainGraph(doc)
				exporter.ResolveChainOutputs(&graph, func(dirName string) string {
					return outPathByDir[dirName]
				})
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
