package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"reqmd/internal/repin"
)

func newRepinCmd() *cobra.Command {
	var (
		yes             bool
		jsonOutput      bool
		promoteUnpinned bool
	)

	cmd := &cobra.Command{
		Use:   "repin <root>",
		Short: "Update version-pin (~N) trace references to the upstream's current version",
		Long: `Update version-pin (~N) trace references to the upstream's current version.

For every requirement whose trace ref has a version pin (e.g. UP-001~1)
and whose upstream's version attribute is higher (e.g. version: 3),
repin proposes changing the pin to the upstream's current version
(UP-001~3). By default the change list is printed and no file is
written; pass --yes to apply.

Predated findings (pin > upstream.version) are surfaced in the
output but never auto-fixed — they indicate a data integrity error
that requires manual review.

Pass --promote-unpinned to also propose pins for trace refs that
have no ~N suffix against a versioned upstream. This converts
"no claim" into "claimed at current version" and is opt-in.`,
		Example: `  # Dry-run: list proposed changes
  reqmd repin spec/

  # Apply
  reqmd repin spec/ --yes

  # Also pin refs that have no ~N
  reqmd repin spec/ --yes --promote-unpinned

  # Machine-readable output
  reqmd repin spec/ --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRepin(cmd, args[0], repinOptions{
				yes:             yes,
				jsonOutput:      jsonOutput,
				promoteUnpinned: promoteUnpinned,
			})
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Apply changes without prompting")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.Flags().BoolVar(&promoteUnpinned, "promote-unpinned", false, "Also pin trace refs that have no ~N against a versioned upstream")
	return cmd
}

type repinOptions struct {
	yes             bool
	jsonOutput      bool
	promoteUnpinned bool
}

func runRepin(cmd *cobra.Command, root string, opt repinOptions) error {
	res, err := repin.Build(root, opt.promoteUnpinned)
	if err != nil {
		return fmt.Errorf("repin: %w", err)
	}
	// No changes: short-circuit.
	if len(res.Deltas) == 0 {
		if opt.jsonOutput {
			return writeJSON(cmd, res)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "no version-pin changes needed")
		return nil
	}

	// Decide whether to apply. --yes applies unconditionally; in
	// interactive mode (stdin is a TTY) we prompt first; otherwise
	// dry-run.
	apply := opt.yes
	stdinNotTTY := false
	aborted := false
	if !apply {
		if !isTerminal(os.Stdin) {
			stdinNotTTY = true
		} else {
			ok, err := promptYes(cmd.ErrOrStderr(), os.Stdin, len(res.Deltas), len(res.ByFile))
			if err != nil {
				return err
			}
			apply = ok
			if !apply {
				aborted = true
				fmt.Fprintln(cmd.ErrOrStderr(), "aborted; no files written.")
			}
		}
	}

	if aborted {
		// User rejected the prompt: don't dump the change list.
		return nil
	}
	if !apply {
		if opt.jsonOutput {
			return writeJSON(cmd, res)
		}
		fmt.Fprint(cmd.OutOrStdout(), repin.FormatText(res))
		if stdinNotTTY {
			fmt.Fprintln(cmd.ErrOrStderr(), "hint: pass --yes to apply (stdin is not a TTY).")
		}
		return nil
	}
	return applyRepin(cmd, res, opt)
}
func applyRepin(cmd *cobra.Command, res repin.Result, opt repinOptions) error {
	rep, err := repin.Apply(res.Deltas)
	if err != nil {
		return fmt.Errorf("apply: %w", err)
	}
	if opt.jsonOutput {
		// Emit a combined report: dry-run Result + apply outcome.
		out := struct {
			repin.Result
			Apply repin.ApplyReport `json:"apply"`
		}{res, rep}
		return writeJSON(cmd, out)
	}
	fmt.Fprint(cmd.OutOrStdout(), repin.FormatTextApplied(rep))
	if len(rep.Skipped) > 0 {
		fmt.Fprintf(cmd.ErrOrStderr(), "skipped %d predated findings (data-integrity errors, fix manually):\n", len(rep.Skipped))
		for _, d := range rep.Skipped {
			fmt.Fprintf(cmd.ErrOrStderr(), "  %s  %s  %s (pin %d > upstream v%d)\n",
				d.File, d.ReqID, d.SourceRef, d.OldPin, d.NewVersion)
		}
	}
	return nil
}

func writeJSON(cmd *cobra.Command, v any) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func isTerminal(f *os.File) bool { return isatty.IsTerminal(f.Fd()) }

func promptYes(w io.Writer, r io.Reader, nDeltas, nFiles int) (bool, error) {
	fmt.Fprintf(w, "Apply %d changes across %d files? [y/N] ", nDeltas, nFiles)
	rdr := bufio.NewReader(r)
	line, err := rdr.ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}
	ans := strings.TrimSpace(line)
	return strings.EqualFold(ans, "y") || strings.EqualFold(ans, "yes"), nil
}
