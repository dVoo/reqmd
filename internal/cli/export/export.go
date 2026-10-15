// Package export implements the reqmd `export` subcommands (CSV, HTML,
// NDJSON tables, and LadybugDB graph output).
package export

import (
	"github.com/spf13/cobra"
)

// NewExportCmd builds the `reqmd export` command with its subcommands.
func NewExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export requirements to CSV, HTML, JSON, or graph",
	}
	cmd.AddCommand(newCsvCmd())
	cmd.AddCommand(newHTMLCmd())
	cmd.AddCommand(newJSONCmd())
	cmd.AddCommand(newGraphCmd())
	return cmd
}
