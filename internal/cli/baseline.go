package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"reqmd/internal/diff"
	"reqmd/internal/parser"
	"reqmd/internal/reporter"
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

	cmd := &cobra.Command{
		Use:   "diff <tag1> <tag2>",
		Short: "Compare requirements between two git tags",
		Long: `Compare requirement specifications between two git tags.

Extracts the spec tree at each tag, parses all requirements,
and produces a semantic diff showing added, removed, and
modified requirements with attribute-level detail.`,
		Example: `  reqmd baseline diff v1.0 v2.0
  reqmd baseline diff v1.0 v2.0 --json
  reqmd baseline diff HEAD~10 HEAD`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBaselineDiff(cmd, args[0], args[1], jsonOutput)
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	return cmd
}

func runBaselineDiff(cmd *cobra.Command, tag1, tag2 string, jsonOutput bool) error {
	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	// Validate we're in a git repo.
	if err := exec.Command("git", "-C", root, "rev-parse", "--is-inside-work-tree").Run(); err != nil {
		return fmt.Errorf("baseline diff requires a git repository")
	}

	// Parse at both tags.
	docs1, schemas1, err := parser.DiscoverAtTagWithSchemas(root, tag1)
	if err != nil {
		return fmt.Errorf("loading %s: %w", tag1, err)
	}

	docs2, schemas2, err := parser.DiscoverAtTagWithSchemas(root, tag2)
	if err != nil {
		return fmt.Errorf("loading %s: %w", tag2, err)
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

	result.Submodules = diff.DiffSubmodules(subs1, subs2)

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

	fmt.Fprint(cmd.OutOrStdout(), output)
	return nil
}
