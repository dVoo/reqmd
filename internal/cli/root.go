package cli

import (
	"github.com/dVoo/reqmd/internal/cli/export"

	"github.com/spf13/cobra"
)

// NewRootCmd builds the root cobra command. version is injected at build time
// via -ldflags "-X main.version=..."; when empty, --version is omitted.
func NewRootCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reqmd",
		Short: "Requirement specification tool — check, ls, stats, export, serve, baseline, repin, init",
		Long: `reqmd — requirement specification tool

reqmd is a CLI tool for authoring, validating, and exporting
requirement specifications stored in Markdown files with embedded
attr blocks validated against JSON Schema (YAML-serialized).

Run "reqmd serve <dir>" for a live-reloading HTML preview, or
"reqmd export html <dir>" to write standalone HTML pages.`,
		Example: `  reqmd check spec/
  reqmd ls spec/
  reqmd export csv spec/`,
		Version:       version,
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
	cmd.AddCommand(newRepinCmd())

	return cmd
}
