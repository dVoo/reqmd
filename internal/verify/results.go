// Package verify loads ephemeral verification results (CTRF reports and
// manual markdown results) and synthesizes them into model.Document
// instances so the existing graph pipeline can run outcome-gated checks.
//
// Results are not persisted in the spec repo — they are loaded per
// `reqmd check --results <path>` invocation, merged into a measureID →
// latestResult map, and injected into the graph as pseudo-requirements.
package verify

import (
	"fmt"
	"os"
	"path/filepath"
	"reqmd/internal/model"
	"reqmd/internal/parser"
	"strings"
	"time"
)

// Outcome is the reqmd verdict axis derived from a verification result.
type Outcome string

// Verification result outcomes.
const (
	OutcomePass         Outcome = "pass"
	OutcomeFail         Outcome = "fail"
	OutcomeSkipped      Outcome = "skipped"
	OutcomeInconclusive Outcome = "inconclusive"
)

// Verdict is the outcome + source for a single measure, ready for
// consumption by exporters and reporters. It's the transport type
// between verify (which loads results) and exporter (which renders
// them), avoiding a direct dependency from exporter on verify.
type Verdict struct {
	Outcome string
	Source  string
}

// Result is one verification result, already mapped to a measure ID (and,
// for case-keyed runs, to a test-case identity). Multiple results for the
// same measure are merged by VerifiedAt (latest wins) before synthesis.
//
// The identity fields (Name, CaseKey, Verifies, Description) back the
// requirement → test-case → result chain model. Today only MeasureID is
// populated and Name is captured from CTRF; CaseKey/Verifies/Description
// are the v2 fields parsed in the binding phase that introduces synthesized
// test cases.
type Result struct {
	// MeasureID is the verification-measure requirement ID this result
	// applies to. May include a `~N` version pin (e.g. "TST-UNT-001~3").
	MeasureID  string
	Outcome    Outcome
	VerifiedAt time.Time
	// Evidence is a URI/path to the source report or record carrying the
	// "corresponding verification measure data".
	Evidence string
	// Verifier is the person/role/system that produced the result.
	Verifier string
	// Source is the file path the result was loaded from (for error
	// messages and graph node File field).
	Source string
	// Name is the human-readable test/case name from the source report
	// (e.g. CTRF tests[].name). Used as the synthesized case title.
	Name string
	// CaseKey is the stable test-case identity when the run is case-keyed
	// (explicit extra.x-reqmd.case, or the normalized (suite, name)).
	// Empty for the degenerate single-result-per-measure model (pattern A).
	CaseKey string
	// Verifies lists the upstream requirement IDs this result's test case
	// exercises (extra.x-reqmd.verifies), with optional `~N` pins. Empty
	// when the result binds directly to MeasureID.
	Verifies []string
	// Description is the Markdown body of the synthesized test case
	// (extra.x-reqmd.description). Used only when the result binds to a
	// synthesized case, never when it binds to an authored node.
	Description string
}

// MergeLatest returns, for each measure ID, the result with the latest
// VerifiedAt. Ties are broken by lexical order of Source for determinism.
// The `~N` pin is stripped from the key so results pinned to different
// versions of the same measure collapse to one entry (the pin is preserved
// on the synthesized trace edge, not the result identity).
func MergeLatest(results []Result) map[string]Result {
	byID := make(map[string]Result)
	for _, r := range results {
		id, _, _ := model.StripPin(r.MeasureID)
		if existing, ok := byID[id]; ok {
			if r.VerifiedAt.After(existing.VerifiedAt) {
				byID[id] = r
			} else if r.VerifiedAt.Equal(existing.VerifiedAt) && r.Source < existing.Source {
				byID[id] = r
			}
		} else {
			byID[id] = r
		}
	}
	return byID
}

// LoadMerged is a convenience wrapper that loads results from the
// given paths and merges them by measure ID (latest verdict wins).
// Returns the merged map, any warnings, and an error. When paths
// is empty/nil, returns (nil, nil, nil) — callers should check for
// nil before calling Synthesize.
func LoadMerged(paths []string) (map[string]Result, []string, error) {
	if len(paths) == 0 {
		return nil, nil, nil
	}
	results, warnings, err := LoadResults(paths)
	if err != nil {
		return nil, nil, err
	}
	return MergeLatest(results), warnings, nil
}

// LoadVerdicts loads results, merges by measure ID, and converts to
// a Verdict map. This is the shared results-loading path for check,
// serve, and export csv/html/graph. Returns all nil when paths is
// empty.
func LoadVerdicts(paths []string) (merged map[string]Result, verdicts map[string]Verdict, warnings []string, err error) {
	merged, warnings, err = LoadMerged(paths)
	if err != nil || merged == nil {
		return nil, nil, nil, err
	}
	verdicts = make(map[string]Verdict, len(merged))
	for measureID, r := range merged {
		verdicts[measureID] = Verdict{
			Outcome: string(r.Outcome),
			Source:  r.Source,
		}
	}
	return merged, verdicts, warnings, nil
}

// LoadResults walks the given paths and loads verification results from
// each. A path may be:
//   - a directory: walked; `.ctrf.json` and CTRF-shaped `.json` files are
//     parsed as CTRF, directories containing `schema.yaml` are loaded as
//     manual results via parser.Discover, other files are skipped.
//   - a file: parsed as CTRF if `.ctrf.json` or CTRF-shaped `.json`.
//
// Returns the merged set of results and any warnings (e.g. unmapped CTRF
// tests). Warnings are returned as CheckResult-shaped messages via the
// returned warning slice; errors abort loading.
func LoadResults(paths []string) ([]Result, []string, error) {
	var results []Result
	var warnings []string

	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return nil, nil, fmt.Errorf("results path %s: %w", p, err)
		}

		if info.IsDir() {
			rs, ws, err := loadDir(p)
			if err != nil {
				return nil, nil, fmt.Errorf("results dir %s: %w", p, err)
			}
			results = append(results, rs...)
			warnings = append(warnings, ws...)
		} else {
			// Single file: must be CTRF.
			rs, ws, err := loadCTRFFile(p)
			if err != nil {
				return nil, nil, fmt.Errorf("results file %s: %w", p, err)
			}
			results = append(results, rs...)
			warnings = append(warnings, ws...)
		}
	}

	return results, warnings, nil
}

// loadDir walks a results directory. CTRF files (`.ctrf.json` or
// CTRF-shaped `.json`) are parsed directly. Subdirectories with a
// `schema.yaml` are loaded as manual-results document dirs via
// parser.Discover. Non-CTRF files (coverage.json, junit.xml, etc.) are
// skipped silently.
//
// Each `.json` file is read once: the bytes are used for the CTRF
// shape check and, if it passes, for the full parse. This avoids
// the double-read that occurred when isCTRFFile and loadCTRFFile
// each opened the file independently.
func loadDir(dir string) ([]Result, []string, error) {
	var results []Result
	var warnings []string

	// CTRF files pending load: (path, data) pairs. The data is
	// already read during the walk for the shape check, so we pass
	// it to loadCTRFData to avoid a second read.
	type ctrfPending struct {
		path string
		data []byte
	}
	var ctrfFiles []ctrfPending
	var manualDirs []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Check for schema.yaml → manual-results doc dir.
			if _, err := os.Stat(filepath.Join(path, "schema.yaml")); err == nil {
				manualDirs = append(manualDirs, path)
				return filepath.SkipDir // don't descend into schema dirs
			}
			return nil
		}
		// CTRF by extension: no probe needed.
		if strings.HasSuffix(info.Name(), ".ctrf.json") {
			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("reading %s: %w", path, err)
			}
			ctrfFiles = append(ctrfFiles, ctrfPending{path, data})
			return nil
		}
		// Plain .json: read once, shape-check, cache if CTRF.
		if strings.HasSuffix(info.Name(), ".json") {
			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("reading %s: %w", path, err)
			}
			if isCTRFFileData(data) {
				ctrfFiles = append(ctrfFiles, ctrfPending{path, data})
			}
			// Non-CTRF .json (coverage.json, etc.) → skip silently.
		}
		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("walking %s: %w", dir, err)
	}

	for _, cf := range ctrfFiles {
		rs, ws, err := loadCTRFData(cf.data, cf.path)
		if err != nil {
			return nil, nil, fmt.Errorf("CTRF %s: %w", cf.path, err)
		}
		results = append(results, rs...)
		warnings = append(warnings, ws...)
	}

	for _, d := range manualDirs {
		rs, err := loadManualDir(d)
		if err != nil {
			return nil, nil, fmt.Errorf("manual results %s: %w", d, err)
		}
		results = append(results, rs...)
	}

	return results, warnings, nil
}

// loadManualDir parses a manual-results document dir via the standard
// parser pipeline and extracts result attrs from each requirement.
func loadManualDir(dir string) ([]Result, error) {
	docs, err := parser.Discover(dir)
	if err != nil {
		return nil, fmt.Errorf("discovering %s: %w", dir, err)
	}
	var results []Result
	for _, doc := range docs {
		for _, req := range doc.Requirements() {
			r, ok := resultFromReq(req)
			if !ok {
				continue
			}
			results = append(results, r)
		}
	}
	return results, nil
}

// resultFromReq builds a Result from a parsed manual-results requirement.
// Returns ok=false if the requirement has no `outcome` attr (not a result).
func resultFromReq(req *model.Node) (Result, bool) {
	outcomeRaw, ok := req.Attrs["outcome"]
	if !ok {
		return Result{}, false
	}
	outcomeStr, ok := outcomeRaw.(string)
	if !ok {
		return Result{}, false
	}
	r := Result{
		Outcome: Outcome(outcomeStr),
		Source:  req.Source,
	}
	// trace → measure ID (first ref; may include ~N pin).
	if trace, ok := req.Attrs[model.AttrTrace].([]any); ok && len(trace) > 0 {
		if s, ok := trace[0].(string); ok {
			r.MeasureID = s
		}
	}
	if r.MeasureID == "" {
		r.MeasureID = req.ID
	}
	if ev, ok := req.Attrs["evidence"].(string); ok {
		r.Evidence = ev
	}
	if v, ok := req.Attrs["verifier"].(string); ok {
		r.Verifier = v
	}
	if va, ok := req.Attrs["verified-at"].(string); ok {
		if t, err := time.Parse("2006-01-02", va); err == nil {
			r.VerifiedAt = t
		}
	}
	return r, true
}
