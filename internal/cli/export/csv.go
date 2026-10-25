package export

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dVoo/reqmd/internal/exporter"
	"github.com/dVoo/reqmd/internal/filter"
	"github.com/dVoo/reqmd/internal/graph"
	"github.com/dVoo/reqmd/internal/parser"
	"github.com/dVoo/reqmd/internal/verify"

	"github.com/spf13/cobra"
)

func newCsvCmd() *cobra.Command {
	var outputDir string
	var resultsPaths []string
	var filterExpr string
	var relaxedFilter bool

	cmd := &cobra.Command{
		Use:   "csv <dir>",
		Short: "Export requirements to CSV format",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := args[0]

			docs, err := parser.Discover(root)
			if err != nil {
				return fmt.Errorf("discovering documents: %w", err)
			}

			// exportDocs is the set of documents to export. When --filter
			// is active, only matching requirements are exported; the graph
			// is still built from the full docs so trace links and verdict
			// roll-ups resolve against the unfiltered tree (same as the
			// HTML export).
			exportDocs := docs
			if filterExpr != "" {
				f, warnings, ferr := filter.CompileForDocs(docs, filterExpr, relaxedFilter)
				if ferr != nil {
					return fmt.Errorf("compiling filter %q: %w", filterExpr, ferr)
				}
				filter.WriteWarnings(os.Stderr, warnings)
				exportDocs, err = f.FilterDocs(docs)
				if err != nil {
					return fmt.Errorf("applying filter %q: %w", filterExpr, err)
				}
			}

			// Load ephemeral verification results when --results is supplied.
			// Verdicts (rolled up over downstream test cases) come from the
			// graph, which is built like the HTML export does: from the
			// unfiltered docs plus synthesized results.
			var hasResults bool
			graphDocs := docs
			if len(resultsPaths) > 0 {
				merged, _, err := verify.LoadMerged(resultsPaths)
				if err != nil {
					return fmt.Errorf("loading verification results: %w", err)
				}
				graphDocs = append(graphDocs, verify.Synthesize(merged))
				hasResults = true
			}

			var verdicts map[string]exporter.VerdictInfo
			if hasResults {
				g, err := graph.New(graphDocs)
				if err != nil {
					return fmt.Errorf("building trace graph: %w", err)
				}
				verdicts = make(map[string]exporter.VerdictInfo)
				for id, v := range g.MeasureVerdicts() {
					verdicts[id] = exporter.VerdictInfoFromGraph(v)
				}
			}

			var exp exporter.CSV
			exp.SetVerdicts(verdicts)

			if outputDir != "" {
				if err := os.MkdirAll(outputDir, outputDirPerm); err != nil {
					return fmt.Errorf("creating output directory: %w", err)
				}
			}

			for _, doc := range exportDocs {
				props := doc.Properties
				dirName := filepath.Base(doc.Path)
				outPath := dirName + "-requirements.csv"

				if outputDir != "" {
					outPath = filepath.Join(outputDir, outPath)
				} else {
					outPath = filepath.Join(doc.Path, outPath)
				}

				f, err := os.Create(outPath)
				if err != nil {
					return fmt.Errorf("creating %s: %w", outPath, err)
				}

				// Close reports the last write error, so both are collected before
				// deciding what to report: a truncated file must not look exported.
				writeErr := exp.Export(f, doc, props)
				closeErr := f.Close()
				if writeErr != nil {
					return fmt.Errorf("exporting %s: %w", doc.Path, writeErr)
				}
				if closeErr != nil {
					return fmt.Errorf("closing %s: %w", outPath, closeErr)
				}
				fmt.Fprintf(os.Stderr, "Wrote %s\n", outPath)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory for CSV files")
	cmd.Flags().StringArrayVar(&resultsPaths, "results", nil, "Load ephemeral verification results (CTRF or manual) to add Verdict, Verdict Source, and Verdict Cases columns. Repeatable.")
	cmd.Flags().StringVar(&filterExpr, "filter", "", "Filter requirements using an expr-lang expression. Only matching requirements are exported.")
	cmd.Flags().BoolVar(&relaxedFilter, "relaxed-filter", false, "Allow the filter to reference attributes not declared in any schema.yaml (warn instead of error).")
	return cmd
}
