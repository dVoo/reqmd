package verify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// CTRF report shape. Only the fields reqmd needs are decoded; the per-test
// `extra` extension point carries the reqmd binding fields (id, case,
// verifies, description). `results` is kept as raw JSON at first so the
// top-level shape can be validated without decoding the (potentially very
// large) test list more than once.
//
// Reference: https://ctrf.io/docs/full-schema
type ctrfResults struct {
	Tests []ctrfTest `json:"tests"`
	Suite string     `json:"suite"`
}

type ctrfTest struct {
	Extra  map[string]any `json:"extra"`
	Name   string         `json:"name"`
	Status string         `json:"status"`
	Stop   int64          `json:"stop"`
	Start  int64          `json:"start"`
	Flaky  bool           `json:"flaky"`
}

// loadCTRFFile parses one CTRF JSON file into Results.
func loadCTRFFile(path string) ([]Result, []string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("reading %s: %w", path, err)
	}
	return loadCTRFData(data, path)
}

// loadCTRFData parses CTRF JSON bytes into Results. The path is
// used only for error messages and evidence attribution.
//
// Binding (extra.x-reqmd):
//   - id present → the result binds directly to that requirement ID
//     (patterns A/C); an explicit `case` is recorded on the result.
//   - no id but verifies present → a test case is synthesized under the
//     case key (explicit `case`, else normalized (suite, name)) with trace
//     edges to each verifies ref (patterns D/E).
//   - x-reqmd present but nothing bindable → an "unbound-result" warning.
//   - no x-reqmd block → skipped silently.
func loadCTRFData(data []byte, path string) ([]Result, []string, error) {
	var probe struct {
		Results json.RawMessage `json:"results"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return nil, nil, fmt.Errorf("parsing CTRF: %w", err)
	}
	if !isCTRFResultsObject(probe.Results) {
		return nil, nil, fmt.Errorf("not a CTRF report: missing top-level results object")
	}
	var report ctrfResults
	if err := json.Unmarshal(probe.Results, &report); err != nil {
		return nil, nil, fmt.Errorf("parsing CTRF: %w", err)
	}

	var results []Result
	var warnings []string
	for _, t := range report.Tests {
		fields := extractXReqmd(t.Extra)
		if !fields.present {
			// Uninstrumented test — silent skip (incremental adoption).
			continue
		}
		result := Result{
			Outcome:     ctrfStatusToOutcome(t.Status, t.Flaky),
			VerifiedAt:  msEpochToTime(t.Stop, t.Start),
			Evidence:    path,
			Source:      path,
			Name:        t.Name,
			Description: fields.Description,
			Verifies:    fields.Verifies,
		}
		switch {
		case fields.ID != "":
			result.MeasureID = fields.ID
			result.CaseKey = fields.Case
		case len(fields.Verifies) > 0:
			caseKey := fields.Case
			if caseKey == "" {
				caseKey = implicitCaseKey(report.Suite, t.Name)
			}
			result.CaseKey = caseKey
			result.MeasureID = testCasePrefix + caseKey
		default:
			warnings = append(warnings, fmt.Sprintf("unbound-result: CTRF test %q in %s declares x-reqmd but binds nothing (need id or verifies)", t.Name, path))
			continue
		}
		results = append(results, result)
	}
	return results, warnings, nil
}

// xreqmdFields holds the reqmd binding fields from a CTRF test's
// extra.x-reqmd block. present is true when the block exists.
type xreqmdFields struct {
	present     bool
	ID          string
	Case        string
	Verifies    []string
	Description string
}

// extractXReqmd reads the reqmd binding fields from a CTRF test's `extra`
// object. Returns present=false when there is no x-reqmd block.
func extractXReqmd(extra map[string]any) xreqmdFields {
	if extra == nil {
		return xreqmdFields{}
	}
	v, ok := extra["x-reqmd"]
	if !ok {
		return xreqmdFields{}
	}
	m, ok := v.(map[string]any)
	if !ok {
		return xreqmdFields{}
	}
	f := xreqmdFields{present: true}
	if id, ok := m["id"].(string); ok {
		f.ID = id
	}
	if c, ok := m["case"].(string); ok {
		f.Case = c
	}
	if d, ok := m["description"].(string); ok {
		f.Description = d
	}
	if verifies, ok := m["verifies"].([]any); ok {
		for _, item := range verifies {
			if s, ok := item.(string); ok {
				f.Verifies = append(f.Verifies, s)
			}
		}
	}
	return f
}

// implicitCaseKey derives a stable case key from the report suite and test
// name for verifies-only bindings (pattern E): suite + "/" + name, or the
// bare name when the report has no suite.
func implicitCaseKey(suite, name string) string {
	if suite == "" {
		return name
	}
	return suite + "/" + name
}

// ctrfStatusToOutcome maps CTRF status to reqmd outcome. A flaky test is
// inconclusive regardless of its (passing) final status.
func ctrfStatusToOutcome(status string, flaky bool) Outcome {
	if flaky {
		return OutcomeInconclusive
	}
	switch status {
	case "passed":
		return OutcomePass
	case "failed":
		return OutcomeFail
	case "skipped":
		return OutcomeSkipped
	default:
		// pending, other, or unknown
		return OutcomeInconclusive
	}
}

// msEpochToTime converts a CTRF millisecond epoch to time.Time, falling
// back to start when stop is 0, then to the zero value.
func msEpochToTime(stop, start int64) time.Time {
	if stop > 0 {
		return time.UnixMilli(stop).UTC()
	}
	if start > 0 {
		return time.UnixMilli(start).UTC()
	}
	return time.Time{}
}

// isCTRFFileData is a cheap shape check on already-read bytes.
// Used by the dir walker to distinguish CTRF `.json` from other
// `.json` (coverage reports, etc.) without reading the file twice.
func isCTRFFileData(data []byte) bool {
	return looksLikeCTRF(data)
}

// looksLikeCTRF reports whether the JSON bytes decode to an object with a
// top-level `results` object. The probe decodes only the top-level key and
// keeps `results` as raw bytes — the potentially huge test list is NOT
// decoded into Go values here (that happens once, in loadCTRFData).
func looksLikeCTRF(data []byte) bool {
	var probe struct {
		Results json.RawMessage `json:"results"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return false
	}
	return isCTRFResultsObject(probe.Results)
}

// isCTRFResultsObject reports whether the raw `results` value is a non-empty
// object (CTRF's results is an object containing `tests[]`, never an array).
// This is a byte-level check — no JSON decode of the contents.
func isCTRFResultsObject(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return false
	}
	// Fast path: CTRF reports don't put whitespace before the object.
	if raw[0] == '{' {
		return true
	}
	trimmed := bytes.TrimLeft(raw, " \t\r\n")
	return len(trimmed) > 0 && trimmed[0] == '{'
}
