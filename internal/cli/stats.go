package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"reqmd/internal/parser"
	"reqmd/internal/reporter"
)

func newStatsCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "stats <dir>",
		Short: "Show attribute-value statistics per document",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			docs, err := parser.Discover(args[0])
			if err != nil {
				return fmt.Errorf("discovering documents: %w", err)
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
	return cmd
}
