package cli

import (
	"fmt"
	"os"

	"github.com/dVoo/reqmd/internal/filter"
	"github.com/dVoo/reqmd/internal/parser"
	"github.com/dVoo/reqmd/internal/reporter"

	"github.com/spf13/cobra"
)

func newStatsCmd() *cobra.Command {
	var jsonOutput bool
	var filterExpr string
	var relaxedFilter bool

	cmd := &cobra.Command{
		Use:   "stats <dir>",
		Short: "Show attribute-value statistics per document",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			docs, err := parser.Discover(args[0])
			if err != nil {
				return fmt.Errorf("discovering documents: %w", err)
			}

			if filterExpr != "" {
				f, warnings, err := filter.CompileForDocs(docs, filterExpr, relaxedFilter)
				if err != nil {
					return fmt.Errorf("compiling filter %q: %w", filterExpr, err)
				}
				filter.WriteWarnings(os.Stderr, warnings)
				docs, err = f.FilterDocs(docs)
				if err != nil {
					return fmt.Errorf("applying filter %q: %w", filterExpr, err)
				}
			}

			summaries := buildDocSummaries(docs)
			totalReqs := 0
			for _, s := range summaries {
				totalReqs += s.ReqCount
			}

			if jsonOutput {
				_, _ = fmt.Fprint(cmd.OutOrStdout(), reporter.FormatStatsJSON(summaries, totalReqs))
			} else {
				_, _ = fmt.Fprint(cmd.OutOrStdout(), reporter.FormatStats(summaries, totalReqs))
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.Flags().StringVar(&filterExpr, "filter", "", "Filter requirements using an expr-lang expression. Only matching requirements are counted in stats.")
	addRelaxedFilterFlag(cmd, &relaxedFilter)
	return cmd
}
