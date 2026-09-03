// Package cli implements the reqmd cobra command tree.
package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"reqmd/internal/filter"
	"reqmd/internal/graph"
	"reqmd/internal/model"
	"reqmd/internal/parser"
	"reqmd/internal/reporter"
	"reqmd/internal/schema"
	"reqmd/internal/verify"
	"runtime"
	"sync"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newCheckCmd() *cobra.Command {
	var schemaPath string
	var jsonOutput bool
	var relaxedVersions bool
	var resultsPaths []string
	var filterExpr string
	var disjointChecks []string

	cmd := &cobra.Command{
		Use:     "check <dir>",
		Aliases: []string{"validate", "v"},
		Short:   "Check (validate) requirements against schema",
		Long: `Check (validate) requirements against their schema.

Given a directory, recursively finds all schema.yaml files and
validates every .md file with requirement attr blocks against the
corresponding schema. In single-file mode (-s), validates a single
.md file against an explicit schema.

With --results, loads ephemeral verification results (CTRF reports
and/or manual markdown results from outside the spec root) and runs
outcome-gated checks (missing-verdict, failing-verdict) in addition to
the standard trace checks. Results are not persisted; they live for the
duration of this run.`,
		Example: `  reqmd check example/
  reqmd check path/to/file.md -s path/to/schema.yaml
  reqmd check spec/ --results ./ci-out/ --results ./reviews/`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			// Single-file mode
			if schemaPath != "" {
				output, err := validateSingleFile(path, schemaPath, jsonOutput, relaxedVersions)
				if output != "" {
					if _, werr := fmt.Fprint(cmd.OutOrStdout(), output); werr != nil {
						return fmt.Errorf("writing output: %w", werr)
					}
				}
				return err
			}
			output, err := validateDir(path, jsonOutput, relaxedVersions, resultsPaths, filterExpr, disjointChecks)
			if output != "" {
				if _, werr := fmt.Fprint(cmd.OutOrStdout(), output); werr != nil {
					return fmt.Errorf("writing output: %w", werr)
				}
			}
			return err
		},
	}

	cmd.Flags().StringVarP(&schemaPath, "schema", "s", "", "Schema file for single-file validation")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.Flags().BoolVar(&relaxedVersions, "relaxed-versions", false, "Demote outdated version-pin findings from ERROR to WARNING. Predated (pin ahead of upstream) stays ERROR.")
	cmd.Flags().StringArrayVar(&resultsPaths, "results", nil, "Load ephemeral verification results (CTRF .ctrf.json or manual-results dirs with schema.yaml) and run outcome-gated checks. Repeatable. May be a dir (walked, auto-detected) or a single CTRF file.")
	cmd.Flags().StringVar(&filterExpr, "filter", "", "Filter requirements using an expr-lang expression (e.g. '\"Premium\" in variant'). Only matching requirements are checked and reported.")
	cmd.Flags().StringArrayVar(&disjointChecks, "disjoint-check", nil, "Check that trace-linked requirements have overlapping values for the named array-typed attribute (e.g. variant). Repeatable. Also settable via x-reqmd.disjoint-check in schema.yaml.")
	return cmd
}

// runValidationPipeline runs the common validation pipeline for a set of documents.
// It builds doc headers, runs Pass 1 JSON Schema validation, and (when docs are from
// a discovery walk) runs Pass 2 graph trace checks.
//
// relaxedVersions, when true, demotes "outdated" version-pin findings from
// ERROR to WARNING. Predated findings (pin ahead of upstream) stay ERROR
// because they are a data integrity issue, not a process issue.
func runValidationPipeline(docs []model.Document, root string, jsonOutput, relaxedVersions bool, resultsPaths []string, filterExpr string, disjointChecks []string) (string, error) {
	report := &reporter.Report{}

	// Compile the filter expression (if any) once at startup.
	// Fail fast on syntax errors or references to unknown attributes.
	// The filter scopes authored requirements only: synthesized result
	// pseudo-requirements (from --results) carry no schema attributes and
	// are intentionally excluded from the matching set. Verdict checks
	// read result→measure edges directly, so excluding result nodes here
	// does not affect outcome-gated reporting.
	var filterSet map[string]struct{}
	if filterExpr != "" {
		f, err := filter.CompileForDocs(docs, filterExpr)
		if err != nil {
			return "", fmt.Errorf("compiling filter %q: %w", filterExpr, err)
		}
		filterSet, err = f.MatchingIDs(docs)
		if err != nil {
			return "", fmt.Errorf("collecting matching IDs for filter %q: %w", filterExpr, err)
		}
		report.Filter = filterExpr
	}

	// Load ephemeral verification results (CTRF / manual) when --results
	// is supplied. Results are synthesized into a pseudo-document and
	// appended before graph build so the existing trace machinery and
	// the outcome-gated checks (missing-verdict, failing-verdict) apply.
	var resultDoc model.Document
	var resultWarnings []string
	if len(resultsPaths) > 0 {
		merged, _, warnings, err := verify.LoadVerdicts(resultsPaths)
		if err != nil {
			return "", fmt.Errorf("loading verification results: %w", err)
		}
		resultWarnings = warnings
		resultDoc = verify.Synthesize(merged)
	}

	// Precompute the flat requirement list per doc (derived from the
	// content tree; collected once for all passes below).
	reqsByDoc := make([][]*model.Node, len(docs))
	for i := range docs {
		reqsByDoc[i] = docs[i].Requirements()
	}

	// Build doc headers (skip the synthetic results doc — it has no
	// schema and is not a user-facing document).
	for i, doc := range docs {
		title := schema.Title(doc.Schema)
		ids := make([]string, 0, len(reqsByDoc[i]))
		for _, req := range reqsByDoc[i] {
			if !filterID(filterSet, req.ID) {
				continue
			}
			ids = append(ids, req.ID)
		}
		report.DocHeaders = append(report.DocHeaders, reporter.DocHeader{
			Path:     doc.Path,
			Title:    title,
			ReqCount: len(ids),
			ReqIDs:   ids,
		})
	}

	// Pass 1: JSON Schema validation (spec docs only; result doc has no schema)
	// Schema compilation and per-requirement validation are both parallelized.
	// Compile is per-doc-dir (independent schemas), validate is per-requirement
	// (the compiled schema is read-only after construction).

	// Phase A: compile all schemas in parallel.
	type compileResult struct {
		compiled *schema.Compiled
		err      error
	}
	compiles := make([]compileResult, len(docs))
	var compileWG sync.WaitGroup
	compileSem := make(chan struct{}, runtime.NumCPU())
	for i, doc := range docs {
		compileSem <- struct{}{}
		compileWG.Go(func() {
			defer func() { <-compileSem }()
			c, err := schema.Compile(doc.Schema, doc.Path)
			compiles[i] = compileResult{compiled: c, err: err}
		})
	}
	compileWG.Wait()

	// Collect compile errors and count total reqs serially (cheap).
	for i, doc := range docs {
		if cr := compiles[i]; cr.err != nil {
			report.ParseErrors = append(report.ParseErrors, reporter.ParseError{
				File:    doc.Path,
				Message: fmt.Sprintf("compiling schema: %v", cr.err),
			})
			continue
		}
		for _, req := range reqsByDoc[i] {
			if !filterID(filterSet, req.ID) {
				continue
			}
			report.TotalReqs++
		}
	}

	// Phase B: validate all requirements in parallel across all docs
	// that compiled successfully. Each Validate call is independent —
	// the Compiled struct is read-only and each req's attrs map is separate.
	type valResult struct {
		err error
		ok  bool
	}

	// Flatten the work list: (jobIdx, docIdx, reqIdx) triples for docs
	// with valid schemas.
	type valJob struct {
		jobIdx int
		docIdx int
		reqIdx int
	}
	var jobs []valJob
	for i, cr := range compiles {
		if cr.err != nil || cr.compiled == nil {
			continue
		}
		for j := range reqsByDoc[i] {
			if !filterID(filterSet, reqsByDoc[i][j].ID) {
				continue
			}
			jobs = append(jobs, valJob{jobIdx: len(jobs), docIdx: i, reqIdx: j})
		}
	}

	// Worker pool sized to NumCPU: one goroutine per requirement would
	// spawn thousands of short-lived goroutines on large trees. Workers
	// pull jobs from a channel and write results by index, so the
	// collected order matches the job order deterministically.
	valResults := make([]valResult, len(jobs))
	workers := min(runtime.NumCPU(), len(jobs))
	var valWG sync.WaitGroup
	jobCh := make(chan valJob)
	for range workers {
		valWG.Go(func() {
			for job := range jobCh {
				req := reqsByDoc[job.docIdx][job.reqIdx]
				err := compiles[job.docIdx].compiled.Validate(req.Attrs)
				valResults[job.jobIdx] = valResult{err: err, ok: err == nil}
			}
		})
	}
	for _, job := range jobs {
		jobCh <- job
	}
	close(jobCh)
	valWG.Wait()

	// Collect validation results in order.
	for k, job := range jobs {
		req := reqsByDoc[job.docIdx][job.reqIdx]
		if valResults[k].ok {
			report.ValidReqs++
		} else {
			report.ValErrors = append(report.ValErrors, reporter.ValidationError{
				File:    req.Source,
				ReqID:   req.ID,
				Message: valResults[k].err.Error(),
			})
		}
	}

	// Pass 2: Graph build + trace validation (root is non-empty only for dir mode).
	// When results were loaded, the synthesized result doc is appended so
	// outcome-gated checks run.
	if root != "" {
		graphDocs := docs
		if len(resultsPaths) > 0 {
			graphDocs = append(graphDocs, resultDoc)
		}
		g, err := graph.New(graphDocs)
		if err != nil {
			report.ParseErrors = append(report.ParseErrors, reporter.ParseError{
				File:    root,
				Message: fmt.Sprintf("building trace graph: %v", err),
			})
		} else {
			// Collect disjoint-check attrs from CLI flags + schema declarations.
			allDisjointAttrs := collectDisjointAttrs(docs, disjointChecks)
			if len(allDisjointAttrs) > 0 {
				g.SetDisjointAttrs(allDisjointAttrs)
			}
			if filterSet != nil {
				g.SetFilter(filterSet)
			}
			report.GraphChecks = g.CheckResults()
			if relaxedVersions {
				demoteOutdatedVersionPins(report.GraphChecks)
			}
			// Surface unmapped-CTRF-test warnings as WARNING graph
			// checks so they appear in both text and JSON output.
			for _, w := range resultWarnings {
				report.GraphChecks = append(report.GraphChecks, graph.CheckResult{
					Level:   graph.LevelWarning,
					Message: w,
				})
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
		if c.Code == graph.CodeVersionPin && c.Direction == graph.DirOutdated && c.Level == graph.LevelError {
			c.Level = graph.LevelWarning
		}
	}
}

func validateDir(root string, jsonOutput, relaxedVersions bool, resultsPaths []string, filterExpr string, disjointChecks []string) (string, error) {
	docs, err := parser.Discover(root)
	if err != nil {
		return "", fmt.Errorf("discovering documents: %w", err)
	}
	return runValidationPipeline(docs, root, jsonOutput, relaxedVersions, resultsPaths, filterExpr, disjointChecks)
}

func validateSingleFile(filePath, schemaPath string, jsonOutput, relaxedVersions bool) (string, error) {
	schemaRaw, err := os.ReadFile(schemaPath)
	if err != nil {
		return "", fmt.Errorf("reading schema %s: %w", schemaPath, err)
	}
	var schemaAny any
	if err = yaml.Unmarshal(schemaRaw, &schemaAny); err != nil {
		return "", fmt.Errorf("parsing schema: %w", err)
	}

	reqs, err := parser.ParseSingleFile(filePath)
	if err != nil {
		return "", fmt.Errorf("parsing %s: %w", filePath, err)
	}

	// Wrap as a single Document for the shared pipeline
	doc := model.Document{
		Path:   filepath.Dir(filePath),
		Schema: schemaAny,
		Nodes:  reqs,
	}

	return runValidationPipeline([]model.Document{doc}, "", jsonOutput, relaxedVersions, nil, "", nil)
}

// collectDisjointAttrs merges disjoint-check attribute names from CLI flags
// and x-reqmd.disjoint-check schema declarations into a deduplicated slice.
func collectDisjointAttrs(docs []model.Document, cliFlags []string) []string {
	seen := make(map[string]struct{})
	for _, attr := range cliFlags {
		if attr != "" {
			seen[attr] = struct{}{}
		}
	}
	for _, doc := range docs {
		if doc.XReqmd != nil {
			for _, attr := range doc.XReqmd.DisjointCheck {
				if attr != "" {
					seen[attr] = struct{}{}
				}
			}
		}
	}
	if len(seen) == 0 {
		return nil
	}
	out := make([]string, 0, len(seen))
	for attr := range seen {
		out = append(out, attr)
	}
	return out
}

// formatReport conditionally returns the report in text or JSON format.
func formatReport(r *reporter.Report, jsonOutput bool) string {
	if jsonOutput {
		return r.FormatJSON()
	}
	return r.Format()
}

// filterID reports whether reqID passes the active filter set. A nil set
// means no filter is active — every requirement passes.
func filterID(set map[string]struct{}, reqID string) bool {
	if set == nil {
		return true
	}
	_, ok := set[reqID]
	return ok
}
