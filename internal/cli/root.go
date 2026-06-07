package cli

import (
	"github.com/spf13/cobra"

	"reqmd/internal/cli/export"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reqmd",
		Short: "Requirement specification tool — check, ls, stats, export, serve, baseline, init",
		Long: `reqmd — requirement specification tool

reqmd is a CLI tool for authoring, validating, and exporting
requirement specifications stored in Markdown files with embedded
attr blocks validated against JSON Schema (YAML-serialized).`,
		Example: `  reqmd check example/
  reqmd ls example/
  reqmd export csv example/`,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.AddCommand(newCheckCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newStatsCmd())
	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newServeCmd())
	cmd.AddCommand(newBaselineCmd())
	cmd.AddCommand(export.NewExportCmd())

	return cmd
}
