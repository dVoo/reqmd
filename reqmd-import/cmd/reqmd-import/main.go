package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"reqmd-import/internal/cli"
	// Blank-import language plugins so their init() functions self-register
	// into the lang registry. Without these, `reqmd-import extract` would
	// walk the source tree and skip every file (no plugin claims any
	// extension). Add new plugins here as they land.
	_ "reqmd-import/internal/lang/go"
	_ "reqmd-import/internal/lang/python"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "reqmd-import",
		Short: "Extract source code into reqmd requirement files",
		Long:  "reqmd-import scans source files (Go, Python) and generates ephemeral .md requirement files for downstream reqmd validation.",
		// When invoked with no subcommand, print usage to stderr and exit non-zero.
		// Without this, cobra would print nothing and exit 0, which is misleading:
		// a user running `reqmd-import` with no args sees a silent success and may
		// think the tool is broken.
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Usage()
		},
	}

	// All subcommands live in the cli package. Register() wires them onto
	// the root command; future subcommands (list-languages, version, etc.)
	// will be added there too.
	cli.Register(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		// cmd.Usage() already wrote to stderr; just print the error
		// summary and exit.
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
