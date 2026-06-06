package export

import (
	"github.com/spf13/cobra"
)

func NewExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export requirements to CSV, HTML, or graph",
	}
	cmd.AddCommand(newCsvCmd())
	cmd.AddCommand(newHtmlCmd())
	cmd.AddCommand(newGraphCmd())
	return cmd
}
