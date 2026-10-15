// Package cli hosts the cobra command tree for the reqmd-import binary.
package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"reqmd-import/exchange"
	"reqmd-import/internal/writer"
)

// NewRootCmd builds the reqmd-import command tree. Rendering an exchange
// bundle is the tool's only job, so it is the root command itself.
func NewRootCmd(version string) *cobra.Command {
	var outDir string
	var check bool

	cmd := &cobra.Command{
		Use:     "reqmd-import <bundle.json|->",
		Short:   "Render a reqmd-import/v1 exchange bundle into reqmd documents",
		Version: version,
		Long: `Reads a reqmd-import/v1 JSON bundle and writes the reqmd documents it
describes (one schema.yaml per document, one .md for documents with
requirements) into the output directory.

This is the single renderer for every importer: reqmd-code and third-party
scripts only emit the exchange format, they never write markdown or YAML.

Examples:
  reqmd-code ./internal | reqmd-import - -o spec/03-software/imported
  reqmd-import doors-foo.json -o spec/00-doors
  reqmd-import bundle.json -o spec/00-door --check`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd, args[0], outDir, check)
		},
	}

	cmd.Flags().StringVarP(&outDir, "output", "o", ".", "directory to write the generated reqmd documents into")
	cmd.Flags().BoolVar(&check, "check", false, "run 'reqmd check' on the output directory after writing")

	return cmd
}

func run(cmd *cobra.Command, src, outDir string, check bool) error {
	var in io.Reader = cmd.InOrStdin()
	if src != "-" {
		f, err := os.Open(src)
		if err != nil {
			return fmt.Errorf("open bundle %q: %w", src, err)
		}
		defer f.Close()
		in = f
	}

	bundle, err := exchange.Decode(in)
	if err != nil {
		return err
	}
	files, err := bundle.Build()
	if err != nil {
		return err
	}
	if err := writer.Write(outDir, files); err != nil {
		return err
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "reqmd-import: wrote %d file(s) to %s\n", len(files), outDir)

	if !check {
		return nil
	}
	c := exec.Command("reqmd", "check", outDir)
	c.Stdin, c.Stdout, c.Stderr = cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr()
	if err := c.Run(); err != nil {
		return fmt.Errorf("reqmd check %s: %w", outDir, err)
	}
	return nil
}
