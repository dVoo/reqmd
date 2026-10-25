package cli

import (
	"fmt"
	"os"

	"github.com/dVoo/reqmd/internal/filter"
	"github.com/dVoo/reqmd/internal/parser"
	"github.com/dVoo/reqmd/internal/reporter"

	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var jsonOutput bool
	var filterExpr string
	var relaxedFilter bool

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

			var output string
			if jsonOutput {
				output = reporter.FormatListJSON(summaries)
			} else {
				output = reporter.FormatList(summaries)
			}
			if _, err := fmt.Fprint(cmd.OutOrStdout(), output); err != nil {
				return fmt.Errorf("writing output: %w", err)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.Flags().StringVar(&filterExpr, "filter", "", "Filter requirements using an expr-lang expression. Only matching requirements are listed.")
	addRelaxedFilterFlag(cmd, &relaxedFilter)
	return cmd
}
