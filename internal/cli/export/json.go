package export

import (
	"bufio"
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

// jsonTable names the four NDJSON files written by `export json`.
const (
	tableRequirements = "requirements.ndjson"
	tableTraces       = "traces.ndjson"
	tableResults      = "results.ndjson"
	tableFindings     = "findings.ndjson"
)

func newJSONCmd() *cobra.Command {
	var outputDir string
	var resultsPaths []string
	var filterExpr string
	var relaxedFilter bool
	var disjointChecks []string

	cmd := &cobra.Command{
		Use:   "json <dir>",
		Short: "Export requirements, traces, verdicts, and findings as NDJSON tables",
		Long: `Export the whole spec tree as flat NDJSON tables for reporting platforms
(Databricks, Snowflake, BigQuery, and similar) to ingest directly.

Four files are written, one JSON object per line:

  requirements.ndjson  one row per content node (req/container/info)
  traces.ndjson        one row per declared trace reference
  results.ndjson       one row per measure verdict (only with --results)
  findings.ndjson      one row per check finding

The output is a pure function of the tree plus any --results, so repeated
exports of the same tree are byte-identical.

With --filter, requirements.ndjson is limited to matching requirements while
traces.ndjson and findings.ndjson are still computed over the full graph (the
same semantics as the CSV and HTML exports).`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := args[0]

			docs, err := parser.Discover(root)
			if err != nil {
				return fmt.Errorf("discovering documents: %w", err)
			}

			// exportDocs holds the requirement rows to write. When --filter
			// is active only matching requirements are exported; the graph
			// is still built from the full docs so trace resolution and
			// verdict roll-ups see the whole tree (same as CSV/HTML).
			exportDocs := docs
			var filterSet map[string]struct{}
			if filterExpr != "" {
				f, warnings, compileErr := filter.CompileForDocs(docs, filterExpr, relaxedFilter)
				if compileErr != nil {
					return fmt.Errorf("compiling filter %q: %w", filterExpr, compileErr)
				}
				filter.WriteWarnings(os.Stderr, warnings)
				filterSet, err = f.MatchingIDs(docs)
				if err != nil {
					return fmt.Errorf("applying filter %q: %w", filterExpr, err)
				}
				exportDocs, err = f.FilterDocs(docs)
				if err != nil {
					return fmt.Errorf("applying filter %q: %w", filterExpr, err)
				}
			}

			// Load ephemeral verification results when --results is supplied.
			// The synthesized result document is appended to the graph docs
			// so verdicts roll up the same way the HTML/CSV exports do, but
			// it is skipped by the requirement and trace writers (Synthetic).
			graphDocs := docs
			var hasResults bool
			if len(resultsPaths) > 0 {
				merged, _, loadErr := verify.LoadMerged(resultsPaths)
				if loadErr != nil {
					return fmt.Errorf("loading verification results: %w", loadErr)
				}
				graphDocs = append(graphDocs, verify.Synthesize(merged))
				hasResults = true
			}

			// One graph build feeds all three derived tables: traces,
			// verdicts, and findings.
			g, err := graph.New(graphDocs)
			if err != nil {
				return fmt.Errorf("building trace graph: %w", err)
			}
			if attrs := graph.CollectDisjointAttrs(docs, disjointChecks); len(attrs) > 0 {
				g.SetDisjointAttrs(attrs)
			}
			if filterSet != nil {
				g.SetFilter(filterSet)
			}

			var exp exporter.JSON
			if hasResults {
				verdicts := make(map[string]exporter.VerdictInfo)
				for id, v := range g.MeasureVerdicts() {
					verdicts[id] = exporter.VerdictInfoFromGraph(v)
				}
				exp.SetVerdicts(verdicts)
			}

			outDir := outputDir
			if outDir == "" {
				outDir = root
			}
			if err := os.MkdirAll(outDir, outputDirPerm); err != nil {
				return fmt.Errorf("creating output directory: %w", err)
			}

			tables := []struct {
				write func(w *bufio.Writer) error
				name  string
			}{
				{name: tableRequirements, write: func(w *bufio.Writer) error {
					return exp.ExportRequirements(w, exportDocs)
				}},
				{name: tableTraces, write: func(w *bufio.Writer) error {
					return exp.ExportTraces(w, exportDocs, g)
				}},
				{name: tableResults, write: func(w *bufio.Writer) error {
					return exp.ExportResults(w, g)
				}},
				{name: tableFindings, write: func(w *bufio.Writer) error {
					return exp.ExportFindings(w, g)
				}},
			}

			for _, t := range tables {
				outPath := filepath.Join(outDir, t.name)

				f, err := os.Create(outPath)
				if err != nil {
					return fmt.Errorf("creating %s: %w", outPath, err)
				}
				w := bufio.NewWriter(f)
				if err := t.write(w); err != nil {
					_ = f.Close()
					return fmt.Errorf("exporting %s: %w", outPath, err)
				}
				if err := w.Flush(); err != nil {
					_ = f.Close()
					return fmt.Errorf("flushing %s: %w", outPath, err)
				}
				if err := f.Close(); err != nil {
					return fmt.Errorf("closing %s: %w", outPath, err)
				}
				fmt.Fprintf(os.Stderr, "Wrote %s\n", outPath)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory for the NDJSON tables (default: the exported tree root)")
	cmd.Flags().StringArrayVar(&resultsPaths, "results", nil, "Load ephemeral verification results (CTRF or manual) to emit results.ndjson. Repeatable.")
	cmd.Flags().StringVar(&filterExpr, "filter", "", "Filter requirements using an expr-lang expression. Only matching requirements are written to requirements.ndjson.")
	cmd.Flags().BoolVar(&relaxedFilter, "relaxed-filter", false, "Allow the filter to reference attributes not declared in any schema.yaml (warn instead of error).")
	cmd.Flags().StringArrayVar(&disjointChecks, "disjoint-check", nil, "Check that trace-linked requirements have overlapping values for the named array-typed attribute. Repeatable. Also settable via x-reqmd.disjoint-check in schema.yaml.")
	return cmd
}
