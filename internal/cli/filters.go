package cli

import "github.com/spf13/cobra"

// relaxedFilterUsage is the shared help text for --relaxed-filter, kept in
// one place so every command that accepts --filter documents it identically.
const relaxedFilterUsage = "Allow the filter to reference attributes not declared in any schema.yaml (warn instead of error). Attributes present on a requirement are always allowed."

// addRelaxedFilterFlag registers --relaxed-filter on cmd, storing the value in
// target. Commands pair it with their --filter flag; see filter.CompileForDocs.
func addRelaxedFilterFlag(cmd *cobra.Command, target *bool) {
	cmd.Flags().BoolVar(target, "relaxed-filter", false, relaxedFilterUsage)
}
