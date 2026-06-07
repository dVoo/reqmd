package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"reqmd/internal/graph"
	"reqmd/internal/model"
	"reqmd/internal/parser"
	"reqmd/internal/reporter"
	"reqmd/internal/schema"
)

func newCheckCmd() *cobra.Command {
	var schemaPath string
	var jsonOutput bool
	var relaxedVersions bool

	cmd := &cobra.Command{
		Use:     "check <dir>",
		Aliases: []string{"validate", "v"},
		Short:   "Check (validate) requirements against schema",
		Long: `Check (validate) requirements against their schema.

Given a directory, recursively finds all schema.yaml files and
validates every .md file with requirement attr blocks against the
corresponding schema. In single-file mode (-s), validates a single
.md file against an explicit schema.`,
		Example: `  reqmd check example/
  reqmd check path/to/file.md -s path/to/schema.yaml`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			// Single-file mode
			if schemaPath != "" {
				output, err := validateSingleFile(path, schemaPath, jsonOutput, relaxedVersions)
				if output != "" {
					fmt.Fprint(cmd.OutOrStdout(), output)
				}
				return err
			}

			output, err := validateDir(path, jsonOutput, relaxedVersions)
			if output != "" {
				fmt.Fprint(cmd.OutOrStdout(), output)
			}
			return err
		},
	}

	cmd.Flags().StringVarP(&schemaPath, "schema", "s", "", "Schema file for single-file validation")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.Flags().BoolVar(&relaxedVersions, "relaxed-versions", false, "Demote outdated version-pin findings from ERROR to WARNING. Predated (pin ahead of upstream) stays ERROR.")
	return cmd
}

// runValidationPipeline runs the common validation pipeline for a set of documents.
// It builds doc headers, runs Pass 1 JSON Schema validation, and (when docs are from
// a discovery walk) runs Pass 2 graph trace checks.
//
// relaxedVersions, when true, demotes "outdated" version-pin findings from
// ERROR to WARNING. Predated findings (pin ahead of upstream) stay ERROR
// because they are a data integrity issue, not a process issue.
func runValidationPipeline(docs []model.Document, root string, jsonOutput, relaxedVersions bool) (string, error) {
	report := &reporter.Report{}

	// Build doc headers
	for _, doc := range docs {
		title := schema.SchemaTitle(doc.Schema)
		ids := make([]string, 0, len(doc.Requirements))
		for _, req := range doc.Requirements {
			ids = append(ids, req.ID)
		}
		report.DocHeaders = append(report.DocHeaders, reporter.DocHeader{
			Path:        doc.Path,
			SchemaTitle: title,
			ReqCount:    len(doc.Requirements),
			ReqIDs:      ids,
		})
	}

	// Pass 1: JSON Schema validation
	for _, doc := range docs {
		compiled, err := schema.Compile(doc.Schema, doc.Path)
		if err != nil {
			report.ParseErrors = append(report.ParseErrors, reporter.ParseError{
				File:    doc.Path,
				Message: fmt.Sprintf("compiling schema: %v", err),
			})
			continue
		}
		for _, req := range doc.Requirements {
			report.TotalReqs++
			if err := compiled.Validate(req.Attrs); err != nil {
				report.ValErrors = append(report.ValErrors, reporter.ValidationError{
					File:    req.Source,
					ReqID:   req.ID,
					Message: err.Error(),
				})
			} else {
				report.ValidReqs++
			}
		}
	}

	// Pass 2: Graph build + trace validation (root is non-empty only for dir mode)
	if root != "" {
		g, err := graph.New(docs)
		if err != nil {
			report.ParseErrors = append(report.ParseErrors, reporter.ParseError{
				File:    root,
				Message: fmt.Sprintf("building trace graph: %v", err),
			})
		} else {
			report.GraphChecks = g.CheckResults()
			if relaxedVersions {
				demoteOutdatedVersionPins(report.GraphChecks)
			}
		}
	}

	report.NewIndex()
	if code := report.ExitCode(); code != 0 {
		return formatReport(report, jsonOutput), &reporter.ExitCodeError{Code: code}
	}
	return formatReport(report, jsonOutput), nil
}

// demoteOutdatedVersionPins demotes Code=="version-pin" + Direction=="outdated"
// findings from ERROR to WARNING. Predated findings (Direction=="predated")
// are left at ERROR — they indicate a downstream pinning a version that
// does not exist on the upstream and are a data integrity issue.
func demoteOutdatedVersionPins(checks []graph.CheckResult) {
	for i := range checks {
		c := &checks[i]
		if c.Code == "version-pin" && c.Direction == "outdated" && c.Level == graph.LevelError {
			c.Level = graph.LevelWarning
		}
	}
}

func validateDir(root string, jsonOutput, relaxedVersions bool) (string, error) {
	docs, err := parser.Discover(root)
	if err != nil {
		return "", fmt.Errorf("discovering documents: %w", err)
	}
	return runValidationPipeline(docs, root, jsonOutput, relaxedVersions)
}

func validateSingleFile(filePath, schemaPath string, jsonOutput, relaxedVersions bool) (string, error) {
	schemaRaw, err := os.ReadFile(schemaPath)
	if err != nil {
		return "", fmt.Errorf("reading schema %s: %w", schemaPath, err)
	}
	var schemaAny any
	if err := yaml.Unmarshal(schemaRaw, &schemaAny); err != nil {
		return "", fmt.Errorf("parsing schema: %w", err)
	}

	reqs, err := parser.ParseSingleFile(filePath)
	if err != nil {
		return "", fmt.Errorf("parsing %s: %w", filePath, err)
	}

	// Wrap as a single Document for the shared pipeline
	doc := model.Document{
		Path:         filepath.Dir(filePath),
		Schema:       schemaAny,
		Requirements: reqs,
	}

	return runValidationPipeline([]model.Document{doc}, "", jsonOutput, relaxedVersions)
}

// formatReport conditionally returns the report in text or JSON format.
func formatReport(r *reporter.Report, jsonOutput bool) string {
	if jsonOutput {
		return r.FormatJSON()
	}
	return r.Format()
}
