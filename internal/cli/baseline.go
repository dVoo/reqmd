package cli

import (
	"fmt"
	"os"
	"os/exec"
	"reqmd/internal/diff"
	"reqmd/internal/filter"
	"reqmd/internal/model"
	"reqmd/internal/parser"
	"reqmd/internal/reporter"

	"github.com/spf13/cobra"
)

func newBaselineCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "baseline",
		Short: "Baseline management",
	}
	cmd.AddCommand(newBaselineDiffCmd())
	return cmd
}

func newBaselineDiffCmd() *cobra.Command {
	var jsonOutput bool
	var filterExpr string
	var filterA string
	var filterB string

	cmd := &cobra.Command{
		Use:   "diff <tag1> <tag2>",
		Short: "Compare requirements between two git tags",
		Long: `Compare requirement specifications between two git tags.

Extracts the spec tree at each tag, parses all requirements,
and produces a semantic diff showing added, removed, and
modified requirements with attribute-level detail.

With --filter, both snapshots are scoped to matching requirements
before diffing. With --filter-a and --filter-b, a single snapshot
(at the given ref or HEAD) is compared two ways — useful for
"what does Premium add over Base" reports from a single commit.`,
		Example: `  reqmd baseline diff v1.0 v2.0
  reqmd baseline diff v1.0 v2.0 --json
  reqmd baseline diff v1.0 v2.0 --filter '"Base" in variant'
  reqmd baseline diff --filter-a 'variant == nil' --filter-b '"Premium" in variant' HEAD`,
		Args: cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBaselineDiff(cmd, args, jsonOutput, filterExpr, filterA, filterB)
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.Flags().StringVar(&filterExpr, "filter", "", "Filter both snapshots using an expr-lang expression before diffing.")
	cmd.Flags().StringVar(&filterA, "filter-a", "", "First view filter for same-commit comparison (requires --filter-b).")
	cmd.Flags().StringVar(&filterB, "filter-b", "", "Second view filter for same-commit comparison (requires --filter-a).")
	// Same-commit comparison requires both view filters; --filter (both
	// snapshots) and --filter-a/--filter-b (one snapshot, two views) are
	// mutually exclusive modes.
	// --filter (whole-tree filter) is mutually exclusive with the view pair;
	// each is registered as its own group so --filter-a --filter-b together
	// remain valid. These are programming errors if the flag names drift, so
	// panic rather than silently mis-behave.
	cmd.MarkFlagsRequiredTogether("filter-a", "filter-b")
	cmd.MarkFlagsMutuallyExclusive("filter", "filter-a")
	cmd.MarkFlagsMutuallyExclusive("filter", "filter-b")
	return cmd
}

func runBaselineDiff(cmd *cobra.Command, args []string, jsonOutput bool, filterExpr, filterA, filterB string) error {
	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	// Validate we're in a git repo.
	if err = exec.Command("git", "-C", root, "rev-parse", "--is-inside-work-tree").Run(); err != nil {
		return fmt.Errorf("baseline diff requires a git repository")
	}

	// Same-commit comparison mode: --filter-a and --filter-b both supplied.
	if filterA != "" && filterB != "" {
		return runBaselineDiffFilteredViews(cmd, root, args, jsonOutput, filterA, filterB)
	}

	// Normal two-tag mode.
	if len(args) < 2 {
		return fmt.Errorf("baseline diff requires two git tags (or use --filter-a/--filter-b for same-commit comparison)")
	}
	tag1, tag2 := args[0], args[1]

	docs1, schemas1, err := parser.DiscoverAtTagWithSchemas(root, tag1)
	if err != nil {
		return fmt.Errorf("loading %s: %w", tag1, err)
	}

	docs2, schemas2, err := parser.DiscoverAtTagWithSchemas(root, tag2)
	if err != nil {
		return fmt.Errorf("loading %s: %w", tag2, err)
	}

	// Apply --filter to both snapshots if active.
	if filterExpr != "" {
		docs1, err = applyFilter(docs1, filterExpr)
		if err != nil {
			return err
		}
		docs2, err = applyFilter(docs2, filterExpr)
		if err != nil {
			return err
		}
	}

	// Diff.
	result := diff.Diff(docs1, docs2, schemas1, schemas2, tag1, tag2)

	// Detect submodule changes (Option B: call separately, don't grow Diff signature)
	subs1, err := parser.ListSubmodulesAtTag(root, tag1)
	if err != nil {
		return fmt.Errorf("listing submodules at %s: %w", tag1, err)
	}

	subs2, err := parser.ListSubmodulesAtTag(root, tag2)
	if err != nil {
		return fmt.Errorf("listing submodules at %s: %w", tag2, err)
	}

	result.Submodules = diff.Submodules(subs1, subs2)

	// Format output.
	var output string
	if jsonOutput {
		output, err = reporter.FormatDiffJSON(result)
		if err != nil {
			return fmt.Errorf("formatting JSON: %w", err)
		}
	} else {
		output = reporter.FormatDiff(result)
	}

	if _, err := fmt.Fprint(cmd.OutOrStdout(), output); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}
	return nil
}

// runBaselineDiffFilteredViews loads a single snapshot and compares two
// filtered views of it (RFC §3.4 — "what does Premium add over Base" from
// a single commit).
func runBaselineDiffFilteredViews(cmd *cobra.Command, root string, args []string, jsonOutput bool, filterA, filterB string) error {
	ref := "HEAD"
	if len(args) > 0 {
		ref = args[0]
	}

	docs, schemas, err := parser.DiscoverAtTagWithSchemas(root, ref)
	if err != nil {
		return fmt.Errorf("loading %s: %w", ref, err)
	}

	docsA, err := applyFilter(docs, filterA)
	if err != nil {
		return err
	}
	docsB, err := applyFilter(docs, filterB)
	if err != nil {
		return err
	}

	result := diff.Diff(docsA, docsB, schemas, schemas, filterA, filterB)

	var output string
	if jsonOutput {
		output, err = reporter.FormatDiffJSON(result)
		if err != nil {
			return fmt.Errorf("formatting JSON: %w", err)
		}
	} else {
		output = reporter.FormatDiff(result)
	}

	if _, err := fmt.Fprint(cmd.OutOrStdout(), output); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}
	return nil
}

// applyFilter compiles and applies a filter expression to a doc slice,
// returning the filtered docs. Shared by both baseline diff modes.
func applyFilter(docs []model.Document, expr string) ([]model.Document, error) {
	f, err := filter.CompileForDocs(docs, expr)
	if err != nil {
		return nil, fmt.Errorf("compiling filter %q: %w", expr, err)
	}
	filtered, err := f.FilterDocs(docs)
	if err != nil {
		return nil, fmt.Errorf("applying filter %q: %w", expr, err)
	}
	return filtered, nil
}
