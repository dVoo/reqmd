package verify

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// CTRF top-level report shape. Only the fields reqmd needs are decoded;
// the per-test `extra` extension point is where `x-reqmd.id` lives.
//
// Reference: https://ctrf.io/docs/full-schema
type ctrfReport struct {
	Results ctrfResults `json:"results"`
}

type ctrfResults struct {
	Tests []ctrfTest `json:"tests"`
}

type ctrfTest struct {
	Name     string          `json:"name"`
	Status   string          `json:"status"`
	Stop     int64           `json:"stop"`     // ms-epoch
	Start    int64           `json:"start"`    // ms-epoch
	Flaky    bool            `json:"flaky"`
	Extra    map[string]any  `json:"extra"`    // extension point
	// RawExtra keeps the original extra object so future fields can be
	// harvested without re-decoding.
	RawExtra json.RawMessage `json:"-"`
}

// loadCTRFFile parses one CTRF JSON file into Results.
func loadCTRFFile(path string) ([]Result, []string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	return loadCTRFData(data, path)
}

// loadCTRFData parses CTRF JSON bytes into Results. The path is
// used only for error messages and evidence attribution.
func loadCTRFData(data []byte, path string) ([]Result, []string, error) {
	if !looksLikeCTRF(data) {
		return nil, nil, fmt.Errorf("not a CTRF report: missing top-level results object")
	}

	var report ctrfReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, nil, fmt.Errorf("parsing CTRF: %w", err)
	}

	var results []Result
	var warnings []string
	for _, t := range report.Results.Tests {
		id := extractXReqmdID(t.Extra)
		if id == "" {
			warnings = append(warnings, fmt.Sprintf("CTRF test %q in %s has no x-reqmd.id; skipped", t.Name, path))
			continue
		}
		results = append(results, Result{
			MeasureID:  id,
			Outcome:    ctrfStatusToOutcome(t.Status, t.Flaky),
			VerifiedAt: msEpochToTime(t.Stop, t.Start),
			Evidence:   path,
			Source:     path,
		})
	}
	return results, warnings, nil
}

// extractXReqmdID reads the `x-reqmd.id` field from a CTRF test's `extra`
// object. Returns "" when absent (the test is unmapped).
func extractXReqmdID(extra map[string]any) string {
	if extra == nil {
		return ""
	}
	v, ok := extra["x-reqmd"]
	if !ok {
		return ""
	}
	m, ok := v.(map[string]any)
	if !ok {
		return ""
	}
	id, _ := m["id"].(string)
	return id
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
// top-level `results` object. Used to skip non-CTRF `.json` files in a
// results dir without a brittle full-schema parse.
func looksLikeCTRF(data []byte) bool {
	var probe struct {
		Results json.RawMessage `json:"results"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return false
	}
	if len(probe.Results) == 0 {
		return false
	}
	// `results` must be an object, not an array — CTRF's results is an
	// object containing `tests[]`.
	var obj map[string]any
	return json.Unmarshal(probe.Results, &obj) == nil
}