package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"reqmd/internal/parser"
	"reqmd/internal/reporter"
)

func newListCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:     "ls <dir>",
		Aliases: []string{"list", "l"},
		Short:   "List all requirements as a table",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			docs, err := parser.Discover(args[0])
			if err != nil {
				return fmt.Errorf("discovering documents: %w", err)
			}

			summaries := buildDocSummaries(docs)

			if jsonOutput {
				fmt.Fprint(cmd.OutOrStdout(), reporter.FormatListJSON(summaries))
			} else {
				fmt.Fprint(cmd.OutOrStdout(), reporter.FormatList(summaries))
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}
