package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"reqmd/internal/filter"
	"reqmd/internal/parser"
	"reqmd/internal/reporter"
)

func newStatsCmd() *cobra.Command {
	var jsonOutput bool
	var filterExpr string

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
				f, err := filter.Compile(filterExpr, filter.BuildValidAttrs(docs))
				if err != nil {
					return err
				}
				docs, err = f.FilterDocs(docs)
				if err != nil {
					return err
				}
			}

			summaries := buildDocSummaries(docs)
			totalReqs := 0
			for _, s := range summaries {
				totalReqs += len(s.Rows)
			}

			if jsonOutput {
				fmt.Fprint(cmd.OutOrStdout(), reporter.FormatStatsJSON(summaries, totalReqs))
			} else {
				fmt.Fprint(cmd.OutOrStdout(), reporter.FormatStats(summaries, totalReqs))
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.Flags().StringVar(&filterExpr, "filter", "", "Filter requirements using an expr-lang expression. Only matching requirements are counted in stats.")
	return cmd
}
