package export

import (
	"fmt"
	"os"
	"path/filepath"
	"reqmd/internal/exporter"
	"reqmd/internal/filter"
	"reqmd/internal/parser"
	"reqmd/internal/verify"

	"github.com/spf13/cobra"
)

func newCsvCmd() *cobra.Command {
	var outputDir string
	var resultsPaths []string
	var filterExpr string

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

			if filterExpr != "" {
				var f *filter.Filter
				f, err = filter.CompileForDocs(docs, filterExpr)
				if err != nil {
					return fmt.Errorf("compiling filter %q: %w", filterExpr, err)
				}
				docs, err = f.FilterDocs(docs)
				if err != nil {
					return fmt.Errorf("applying filter %q: %w", filterExpr, err)
				}
			}

			// Load ephemeral verification results when --results is supplied.
			_, vVerdicts, _, err := verify.LoadVerdicts(resultsPaths)
			if err != nil {
				return fmt.Errorf("loading verification results: %w", err)
			}
			verdicts := make(map[string]exporter.VerdictInfo, len(vVerdicts))
			for id, v := range vVerdicts {
				verdicts[id] = exporter.VerdictInfo{Outcome: v.Outcome, Source: v.Source}
			}

			var exp exporter.CSV
			exp.SetVerdicts(verdicts)

			for _, doc := range docs {
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

				if err := exp.Export(f, doc, props); err != nil {
					f.Close()
					return fmt.Errorf("exporting %s: %w", doc.Path, err)
				}
				f.Close()
				fmt.Fprintf(os.Stderr, "Wrote %s\n", outPath)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory for CSV files")
	cmd.Flags().StringArrayVar(&resultsPaths, "results", nil, "Load ephemeral verification results (CTRF or manual) to add Verdict and Verdict Source columns. Repeatable.")
	cmd.Flags().StringVar(&filterExpr, "filter", "", "Filter requirements using an expr-lang expression. Only matching requirements are exported.")
	return cmd
}
