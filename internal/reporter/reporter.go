package reporter

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"reqmd/internal/graph"
)

// ValidationError describes a single requirement that failed validation.
type ValidationError struct {
	File    string
	ReqID   string
	Message string
}

// ParseError describes a structural problem (file unreadable, bad YAML, etc.).
type ParseError struct {
	File    string
	Message string
}

// ExitCodeError wraps a desired process exit code so that main.go
// can set exit code via cobra error handling instead of os.Exit in command handlers.
type ExitCodeError struct {
	Code int
}

func (e *ExitCodeError) Error() string {
	return fmt.Sprintf("reqmd exited with code %d", e.Code)
}

// DocHeader holds per-document metadata for the validate output header.
type DocHeader struct {
	Path        string
	SchemaTitle string
	ReqCount    int
	ReqIDs      []string
}

// Report holds all results from a validation run.
type Report struct {
	ValidReqs   int
	TotalReqs   int
	ValErrors   []ValidationError
	ParseErrors []ParseError
	GraphChecks []graph.CheckResult
	DocHeaders  []DocHeader

	// Pre-indexed maps (populated by NewIndex once at creation)
	valErrorsByDoc   map[string][]ValidationError   // docPath → errors
	graphChecksByDoc map[string][]graph.CheckResult // docPath → graph checks
	valErrorsByReq   map[string][]ValidationError   // reqID → errors
	graphChecksByReq map[string][]graph.CheckResult // reqID → graph checks
}

// NewIndex populates the pre-indexed maps from the flat slices.
// Bucketing by reqID makes the per-reqID lookups in Format()/FormatJSON() O(1)
// instead of scanning all errors/checks per requirement.
func (r *Report) NewIndex() {
	r.valErrorsByDoc = make(map[string][]ValidationError)
	r.valErrorsByReq = make(map[string][]ValidationError)
	for _, ve := range r.ValErrors {
		docPath := filepath.Dir(ve.File)
		r.valErrorsByDoc[docPath] = append(r.valErrorsByDoc[docPath], ve)
		r.valErrorsByReq[ve.ReqID] = append(r.valErrorsByReq[ve.ReqID], ve)
	}
	r.graphChecksByDoc = make(map[string][]graph.CheckResult)
	r.graphChecksByReq = make(map[string][]graph.CheckResult)
	for _, gc := range r.GraphChecks {
		docPath := filepath.Dir(gc.File)
		r.graphChecksByDoc[docPath] = append(r.graphChecksByDoc[docPath], gc)
		r.graphChecksByReq[gc.ReqID] = append(r.graphChecksByReq[gc.ReqID], gc)
	}
}

// ExitCode returns the appropriate exit code per the spec:
//
//	0 = all valid
//	1 = validation errors
//	2 = parse errors
func (r *Report) ExitCode() int {
	if len(r.ParseErrors) > 0 {
		return 2
	}
	// Also check if any GraphChecks are ERROR level
	for _, gc := range r.GraphChecks {
		if gc.Level == graph.LevelError {
			return 1
		}
	}
	if len(r.ValErrors) > 0 {
		return 1
	}
	return 0
}

// Format produces the per-file, per-requirement validation report
// as specified in the spec tree under spec/.
func (r *Report) Format() string {
	var b strings.Builder

	// Parse errors first
	for _, pe := range r.ParseErrors {
		b.WriteString(fmt.Sprintf("PARSE ERROR — %s\n  %s\n\n", pe.File, pe.Message))
	}

	// Per-doc sections
	for _, dh := range r.DocHeaders {
		b.WriteString("=\n")
		b.WriteString(fmt.Sprintf("Schema : %s\n", dh.SchemaTitle))
		b.WriteString(fmt.Sprintf("File   : %s  (%d requirements)\n", dh.Path, dh.ReqCount))
		b.WriteString("=\n")

		for _, reqID := range dh.ReqIDs {
			// Find matching validation error (first only)
			var pass1Err string
			for _, ve := range r.valErrorsByReq[reqID] {
				pass1Err = ve.Message
				break
			}

			// Collect matching graph checks
			graphMsgs := r.graphChecksByReq[reqID]

			// Render status line
			if pass1Err != "" {
				b.WriteString(fmt.Sprintf("  ❌  %s  %s\n", reqID, pass1Err))
			} else {
				hasWarn := false
				for _, g := range graphMsgs {
					if g.Level == graph.LevelWarning {
						hasWarn = true
						break
					}
				}
				if hasWarn {
					b.WriteString(fmt.Sprintf("  ⚠  %s  attributes valid (see trace checks below)\n", reqID))
				} else {
					b.WriteString(fmt.Sprintf("  ✅  %s  all attributes valid\n", reqID))
				}
			}

			// Render graph check detail lines
			for _, g := range graphMsgs {
				prefix := "⚠"
				if g.Level == graph.LevelError {
					prefix = "❌"
				}
				b.WriteString(fmt.Sprintf("  %s  %s  %s\n", prefix, reqID, g.Message))
			}
		}
		b.WriteString("\n")
	}

	// Summary line
	valErrCount := len(r.ValErrors)
	warnCount := 0
	graphErrCount := 0
	for _, gc := range r.GraphChecks {
		switch gc.Level {
		case graph.LevelWarning:
			warnCount++
		case graph.LevelError:
			graphErrCount++
		}
	}

	b.WriteString(fmt.Sprintf("Summary: %d total, %d valid, %d invalid, %d parse errors",
		r.TotalReqs, r.ValidReqs, valErrCount+graphErrCount, len(r.ParseErrors)))
	if warnCount > 0 {
		b.WriteString(fmt.Sprintf(", %d warnings", warnCount))
	}
	b.WriteString("\n")
	return b.String()
}

// FormatList produces a text table for `reqmd list`.
func FormatList(docs []DocumentSummary) string {
	var b strings.Builder

	for _, doc := range docs {
		b.WriteString(fmt.Sprintf("=== %s ===\n", doc.Path))
		// Header row
		b.WriteString(fmt.Sprintf("%-24s", "ID"))
		b.WriteString(" | ")
		for _, prop := range doc.Properties {
			b.WriteString(fmt.Sprintf("%-12s", prop))
			b.WriteString(" | ")
		}
		b.WriteString("\n")
		b.WriteString(strings.Repeat("-", 24+3+len(doc.Properties)*15))
		b.WriteString("\n")

		for _, req := range doc.Rows {
			displayID := req.ID
			if req.Title != "" {
				displayID = req.ID + ": " + req.Title
			}
			b.WriteString(fmt.Sprintf("%-24s", displayID))
			b.WriteString(" | ")
			for _, prop := range doc.Properties {
				val := formatAttrValue(req.Attrs[prop])
				b.WriteString(fmt.Sprintf("%-12s", val))
				b.WriteString(" | ")
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	return b.String()
}

// FormatStats produces a stats breakdown.
func FormatStats(docs []DocumentSummary, totalReqs int) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Requirements: %d\n", totalReqs))
	b.WriteString(fmt.Sprintf("Documents:    %d\n\n", len(docs)))

	for _, doc := range docs {
		b.WriteString(fmt.Sprintf("=== %s (%d reqs) ===\n", doc.Path, len(doc.Rows)))
		for _, prop := range doc.Properties {
			counts := map[string]int{}
			for _, req := range doc.Rows {
				if val, ok := req.Attrs[prop].(string); ok {
					counts[val]++
				}
			}
			if len(counts) > 0 {
				b.WriteString(fmt.Sprintf("  %s:\n", prop))
				for _, val := range sortedKeys(counts) {
					b.WriteString(fmt.Sprintf("    %-20s %d\n", val, counts[val]))
				}
			}
		}
		b.WriteString("\n")
	}
	return b.String()
}

// DocumentSummary is the data needed for list/stats display.
type DocumentSummary struct {
	Path       string
	Properties []string
	Rows       []ReqRow
}

// ReqRow is a single row in list/stats output.
type ReqRow struct {
	ID    string
	Title string
	Attrs map[string]any
}

func sortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// ---------------------------------------------------------------------------
// JSON output types (private — serialized via json.Marshal)
// ---------------------------------------------------------------------------

type jsonReport struct {
	Version     int              `json:"version"`
	ExitCode    int              `json:"exit_code"`
	Summary     jsonSummary      `json:"summary"`
	Documents   []jsonDocSection `json:"documents"`
	ParseErrors []jsonParseErr   `json:"parse_errors"`
}

type jsonSummary struct {
	Total       int `json:"total"`
	Valid       int `json:"valid"`
	Invalid     int `json:"invalid"`
	ParseErrors int `json:"parse_errors"`
	Warnings    int `json:"warnings"`
}

type jsonDocSection struct {
	Path        string          `json:"path"`
	SchemaTitle string          `json:"schema_title"`
	ReqCount    int             `json:"req_count"`
	Reqs        []jsonReqResult `json:"requirements"`
}

type jsonReqResult struct {
	ID     string    `json:"id"`
	Valid  bool      `json:"valid"`
	Checks []jsonChk `json:"checks,omitempty"`
}

type jsonChk struct {
	Level     string `json:"level"`
	Code      string `json:"code,omitempty"`
	Direction string `json:"direction,omitempty"`
	Message   string `json:"message"`
}

type jsonParseErr struct {
	File    string `json:"file"`
	Message string `json:"message"`
}

type jsonListReport struct {
	Documents []jsonListDoc `json:"documents"`
}

type jsonListDoc struct {
	Path       string        `json:"path"`
	Properties []string      `json:"properties"`
	Reqs       []jsonListReq `json:"requirements"`
}

type jsonListReq struct {
	ID    string         `json:"id"`
	Title string         `json:"title,omitempty"`
	Attrs map[string]any `json:"attrs"`
}

type jsonStatsReport struct {
	TotalRequirements int            `json:"total_requirements"`
	TotalDocuments    int            `json:"total_documents"`
	Documents         []jsonStatsDoc `json:"documents"`
}

type jsonStatsDoc struct {
	Path           string                    `json:"path"`
	ReqCount       int                       `json:"req_count"`
	AttributeStats map[string]map[string]int `json:"attribute_stats"`
}

// FormatJSON serializes the full check report as JSON.
func (r *Report) FormatJSON() string {
	// Count warnings
	warnCount := 0
	for _, gc := range r.GraphChecks {
		if gc.Level == graph.LevelWarning {
			warnCount++
		}
	}

	// Count invalid (pass 1 + graph ERROR)
	invalid := len(r.ValErrors)
	for _, gc := range r.GraphChecks {
		if gc.Level == graph.LevelError {
			invalid++
		}
	}

	jr := jsonReport{
		Version:  1,
		ExitCode: r.ExitCode(),
		Summary: jsonSummary{
			Total:       r.TotalReqs,
			Valid:       r.ValidReqs,
			Invalid:     invalid,
			ParseErrors: len(r.ParseErrors),
			Warnings:    warnCount,
		},
		Documents:   make([]jsonDocSection, 0, len(r.DocHeaders)),
		ParseErrors: make([]jsonParseErr, 0, len(r.ParseErrors)),
	}

	for _, dh := range r.DocHeaders {
		doc := jsonDocSection{
			Path:        dh.Path,
			SchemaTitle: dh.SchemaTitle,
			ReqCount:    dh.ReqCount,
			Reqs:        make([]jsonReqResult, 0, len(dh.ReqIDs)),
		}

		for _, reqID := range dh.ReqIDs {
			// Collect all checks for this req into a flat list
			var checks []jsonChk

			// Schema validation errors
			for _, ve := range r.valErrorsByReq[reqID] {
				checks = append(checks, jsonChk{Level: graph.LevelError, Message: ve.Message})
			}
			// Trace graph results
			for _, gc := range r.graphChecksByReq[reqID] {
				checks = append(checks, jsonChk{
					Level:     gc.Level,
					Code:      gc.Code,
					Direction: gc.Direction,
					Message:   gc.Message,
				})
			}

			// Valid = no ERROR-level checks
			valid := true
			for _, c := range checks {
				if c.Level == "ERROR" {
					valid = false
					break
				}
			}

			r := jsonReqResult{ID: reqID, Valid: valid}
			if len(checks) > 0 {
				r.Checks = checks
			}
			doc.Reqs = append(doc.Reqs, r)
		}
		jr.Documents = append(jr.Documents, doc)
	}

	for _, pe := range r.ParseErrors {
		jr.ParseErrors = append(jr.ParseErrors, jsonParseErr{
			File:    pe.File,
			Message: pe.Message,
		})
	}

	b, err := json.MarshalIndent(jr, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error":"json marshal: %v"}`, err)
	}
	return string(b) + "\n"
}

// FormatListJSON serializes list output as JSON.
func FormatListJSON(docs []DocumentSummary) string {
	jr := jsonListReport{
		Documents: make([]jsonListDoc, 0, len(docs)),
	}
	for _, doc := range docs {
		jdoc := jsonListDoc{
			Path:       doc.Path,
			Properties: doc.Properties,
			Reqs:       make([]jsonListReq, 0, len(doc.Rows)),
		}
		for _, row := range doc.Rows {
			jdoc.Reqs = append(jdoc.Reqs, jsonListReq{
				ID:    row.ID,
				Title: row.Title,
				Attrs: row.Attrs,
			})
		}
		jr.Documents = append(jr.Documents, jdoc)
	}
	b, err := json.MarshalIndent(jr, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error":"json marshal: %v"}`, err)
	}
	return string(b) + "\n"
}

// FormatStatsJSON serializes stats output as JSON.
func FormatStatsJSON(docs []DocumentSummary, totalReqs int) string {
	type jsonStatsOut struct {
		TotalRequirements int            `json:"total_requirements"`
		TotalDocuments    int            `json:"total_documents"`
		Documents         []jsonStatsDoc `json:"documents"`
	}
	jr := jsonStatsOut{
		TotalRequirements: totalReqs,
		TotalDocuments:    len(docs),
	}
	for _, doc := range docs {
		jdoc := jsonStatsDoc{
			Path:           doc.Path,
			ReqCount:       len(doc.Rows),
			AttributeStats: make(map[string]map[string]int),
		}
		for _, prop := range doc.Properties {
			counts := map[string]int{}
			for _, req := range doc.Rows {
				if val, ok := req.Attrs[prop].(string); ok {
					counts[val]++
				}
			}
			if len(counts) > 0 {
				jdoc.AttributeStats[prop] = counts
			}
		}
		jr.Documents = append(jr.Documents, jdoc)
	}
	b, err := json.MarshalIndent(jr, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error":"json marshal: %v"}`, err)
	}
	return string(b) + "\n"
}

// formatAttrValue formats an arbitrary attr value for display in text tables.
func formatAttrValue(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case bool:
		return strconv.FormatBool(val)
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case []any:
		b, _ := json.Marshal(val)
		return string(b)
	default:
		return ""
	}
}
