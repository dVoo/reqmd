// Package cli hosts the cobra command tree for the reqmd-import binary.
package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/dVoo/reqmd/reqmd-import/exchange"
	"github.com/dVoo/reqmd/reqmd-import/internal/writer"

	"github.com/spf13/cobra"
)

// NewRootCmd builds the reqmd-import command tree. Rendering an exchange
// bundle is the tool's only job, so it is the root command itself.
func NewRootCmd(version string) *cobra.Command {
	var outDir string
	var check bool
	var validateOnly bool
	var statusMapFlags []string

	cmd := &cobra.Command{
		Use:     "reqmd-import <bundle.json|->",
		Short:   "Render an exchange-format bundle into reqmd documents",
		Version: version,
		Long: `Reads an exchange-format JSON bundle and writes the reqmd documents it
describes (one schema.yaml per document, one .md for documents with items)
into the output directory.

This is the single renderer for every importer: reqmd-code and third-party
scripts only emit the exchange format, they never write markdown or YAML.

Examples:
  reqmd-code ./internal | reqmd-import - -o spec/03-software/imported
  reqmd-import doors-foo.json -o spec/00-doors
  reqmd-import bundle.json -o spec/00-door --check
  reqmd-import bundle.json --validate-only
  reqmd-import bundle.json -o spec/00-door --status-map "In Review=review"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd, runOptions{
				src:          args[0],
				outDir:       outDir,
				check:        check,
				validateOnly: validateOnly,
				statusMap:    statusMapFlags,
			})
		},
	}

	cmd.Flags().StringVarP(&outDir, "output", "o", ".", "directory to write the generated reqmd documents into")
	cmd.Flags().BoolVar(&check, "check", false, "run 'reqmd check' on the output directory after writing")
	cmd.Flags().BoolVar(&validateOnly, "validate-only", false, "validate the bundle and report every violation without writing files")
	cmd.Flags().StringArrayVar(&statusMapFlags, "status-map", nil, `map a source status onto a reqmd status ("In Review=review", repeatable)`)

	return cmd
}

// runOptions carries the flags of one invocation.
type runOptions struct {
	src          string
	outDir       string
	statusMap    []string
	check        bool
	validateOnly bool
}

// parseStatusMap turns "Source=reqmd" pairs into a lookup table.
func parseStatusMap(pairs []string) (map[string]string, error) {
	if len(pairs) == 0 {
		return nil, nil
	}
	out := make(map[string]string, len(pairs))
	for _, p := range pairs {
		from, to, ok := strings.Cut(p, "=")
		if !ok || from == "" || to == "" {
			return nil, fmt.Errorf("invalid --status-map %q, want \"Source=reqmd\"", p)
		}
		out[from] = to
	}
	return out, nil
}

func run(cmd *cobra.Command, opts runOptions) error {
	src, outDir, check := opts.src, opts.outDir, opts.check
	statusMap, err := parseStatusMap(opts.statusMap)
	if err != nil {
		return err
	}
	in := cmd.InOrStdin()
	if src != "-" {
		f, err := os.Open(src)
		if err != nil {
			return fmt.Errorf("open bundle %q: %w", src, err)
		}
		defer func() { _ = f.Close() }()
		in = f
	}

	bundle, err := exchange.Decode(in)
	if err != nil {
		return err
	}
	files, err := bundle.BuildWith(exchange.Options{StatusMap: statusMap})
	if err != nil {
		return err
	}
	if opts.validateOnly {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "reqmd-import: bundle is valid (%d document(s), %d file(s) would be written)\n",
			len(bundle.Documents), len(files))
		return nil
	}
	if err := writer.Write(outDir, files); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "reqmd-import: wrote %d file(s) to %s\n", len(files), outDir)

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
