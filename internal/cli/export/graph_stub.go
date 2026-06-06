//go:build !ladybug

package export

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newGraphCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "graph <dir>",
		Short: "Export requirements to a LadybugDB graph database for querying",
		Long: `Export requirements to a LadybugDB graph database.

This build of reqmd does not include LadybugDB support.
Rebuild with: go build -tags ladybug`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("graph export requires a build with LadybugDB support; rebuild with: go build -tags ladybug")
		},
	}
}
