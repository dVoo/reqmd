// Package reporter formats reqmd validation results, per-document
// listings, and statistics as human-readable text or JSON.
package reporter

import (
	"encoding/json"
	"fmt"

	"reqmd/internal/graph"
	"reqmd/internal/model"
	"sort"
	"strings"
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

// ExitCode returns the process exit code this error carries.
func (e *ExitCodeError) ExitCode() int { return e.Code }

// DocHeader holds per-document metadata for the validate output header.
type DocHeader struct {
	Path     string
	Title    string
	ReqIDs   []string
	ReqCount int
}

// Report holds all results from a validation run.
type Report struct {
	valErrorsByReq   map[string][]ValidationError
	graphChecksByReq map[string][]graph.CheckResult
	Filter           string
	ValErrors        []ValidationError
	ParseErrors      []ParseError
	GraphChecks      []graph.CheckResult
	DocHeaders       []DocHeader
	ValidReqs        int
	TotalReqs        int
}

// NewIndex populates the pre-indexed maps from the flat slices.
// Bucketing by reqID makes the per-reqID lookups in Format()/FormatJSON() O(1)
// instead of scanning all errors/checks per requirement.
func (r *Report) NewIndex() {
	r.valErrorsByReq = make(map[string][]ValidationError)
	for _, ve := range r.ValErrors {
		r.valErrorsByReq[ve.ReqID] = append(r.valErrorsByReq[ve.ReqID], ve)
	}
	r.graphChecksByReq = make(map[string][]graph.CheckResult)
	for _, gc := range r.GraphChecks {
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
	_, errors := r.counts()
	if errors > 0 {
		return 1
	}
	if len(r.ValErrors) > 0 {
		return 1
	}
	return 0
}

// counts tallies WARNING and ERROR-level graph checks in a single pass.
// Shared by ExitCode, Format, and FormatJSON so the summary numbers are
// computed the same way everywhere.
func (r *Report) counts() (warnings, errors int) {
	for _, gc := range r.GraphChecks {
		switch gc.Level {
		case graph.LevelWarning:
			warnings++
		case graph.LevelError:
			errors++
		}
	}
	return warnings, errors
}

// Format produces the per-file, per-requirement validation report
// as specified in the spec tree under spec/.
func (r *Report) Format() string {
	var b strings.Builder

	// Parse errors first
	for _, pe := range r.ParseErrors {
		fmt.Fprintf(&b, "PARSE ERROR — %s\n  %s\n\n", pe.File, pe.Message)
	}

	// Per-doc sections
	for _, dh := range r.DocHeaders {
		b.WriteString("=\n")
		fmt.Fprintf(&b, "Schema : %s\n", dh.Title)
		fmt.Fprintf(&b, "File   : %s  (%d requirements)\n", dh.Path, dh.ReqCount)
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

			// Extract verdict info (if present) for the status line.
			verdict := ""
			verdictSource := ""
			for _, g := range graphMsgs {
				if g.Code == graph.CodeVerdict {
					verdict = g.Message
					verdictSource = g.File
					break
				}
			}

			// Render status line
			if pass1Err != "" {
				fmt.Fprintf(&b, "  ❌  %s  %s\n", reqID, pass1Err)
			} else {
				hasWarn := false
				hasError := false
				for _, g := range graphMsgs {
					switch g.Level {
					case graph.LevelWarning:
						hasWarn = true
					case graph.LevelError:
						hasError = true
					}
				}
				switch {
				case hasError:
					fmt.Fprintf(&b, "  ❌  %s  attributes valid (see trace checks below)\n", reqID)
				case hasWarn:
					fmt.Fprintf(&b, "  ⚠  %s  attributes valid (see trace checks below)\n", reqID)
				case verdict != "":
					if verdictSource != "" {
						fmt.Fprintf(&b, "  ✅  %s  all attributes valid (%s — %s)\n", reqID, verdict, verdictSource)
					} else {
						fmt.Fprintf(&b, "  ✅  %s  all attributes valid (%s)\n", reqID, verdict)
					}
				default:
					fmt.Fprintf(&b, "  ✅  %s  all attributes valid\n", reqID)
				}
			}

			// Render graph check detail lines (skip verdict INFO — already
			// shown in the status line annotation).
			for _, g := range graphMsgs {
				if g.Code == graph.CodeVerdict {
					continue
				}
				prefix := "⚠"
				switch g.Level {
				case graph.LevelError:
					prefix = "❌"
				case graph.LevelInfo:
					prefix = "✅"
				}
				fmt.Fprintf(&b, "  %s  %s  %s\n", prefix, reqID, g.Message)
			}
		}
		b.WriteString("\n")
	}

	// Summary line
	valErrCount := len(r.ValErrors)
	warnCount, graphErrCount := r.counts()

	fmt.Fprintf(&b, "Summary: %d total, %d valid, %d invalid, %d parse errors",
		r.TotalReqs, r.ValidReqs, valErrCount+graphErrCount, len(r.ParseErrors))
	if warnCount > 0 {
		fmt.Fprintf(&b, ", %d warnings", warnCount)
	}
	b.WriteString("\n")
	return b.String()
}

// FormatList produces a text table for `reqmd list`. Every node appears
// in document order with a leading Type column: requirements carry their
// ID (+ title) and attributes, containers and info items show their
// heading text with empty attribute cells.
func FormatList(docs []DocumentSummary) string {
	var b strings.Builder

	for _, doc := range docs {
		fmt.Fprintf(&b, "=== %s ===\n", doc.Path)
		// Header row
		fmt.Fprintf(&b, "%-10s", "Type")
		b.WriteString(" | ")
		fmt.Fprintf(&b, "%-24s", "ID")
		b.WriteString(" | ")
		for _, prop := range doc.Properties {
			fmt.Fprintf(&b, "%-12s", prop)
			b.WriteString(" | ")
		}
		b.WriteString("\n")
		b.WriteString(strings.Repeat("-", 10+3+24+3+len(doc.Properties)*15))
		b.WriteString("\n")

		for _, row := range doc.Rows {
			fmt.Fprintf(&b, "%-10s", row.Kind.String())
			b.WriteString(" | ")
			displayID := row.ID
			if row.Title != "" {
				if displayID != "" {
					displayID = row.ID + ": " + row.Title
				} else {
					displayID = row.Title
				}
			}
			fmt.Fprintf(&b, "%-24s", displayID)
			b.WriteString(" | ")
			for _, prop := range doc.Properties {
				val := formatAttrValue(row.Attrs[prop])
				fmt.Fprintf(&b, "%-12s", val)
				b.WriteString(" | ")
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	return b.String()
}

// FormatStats produces a stats breakdown. The per-document header shows
// requirement and item counts; a type breakdown section precedes the
// attribute-value breakdowns.
func FormatStats(docs []DocumentSummary, totalReqs int) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Requirements: %d\n", totalReqs)
	fmt.Fprintf(&b, "Documents:    %d\n\n", len(docs))

	for _, doc := range docs {
		fmt.Fprintf(&b, "=== %s (%d reqs", doc.Path, doc.ReqCount)
		if doc.ItemCount > 0 {
			label := "items"
			if doc.ItemCount == 1 {
				label = "item"
			}
			fmt.Fprintf(&b, ", %d %s", doc.ItemCount, label)
		}
		b.WriteString(") ===\n")

		if doc.ItemCount > 0 {
			counts := map[string]int{}
			for _, row := range doc.Rows {
				counts[row.Kind.String()]++
			}
			if len(counts) > 0 {
				b.WriteString("  type:\n")
				for _, val := range sortedKeys(counts) {
					fmt.Fprintf(&b, "    %-20s %d\n", val, counts[val])
				}
			}
		}

		for _, prop := range doc.Properties {
			counts := map[string]int{}
			for _, row := range doc.Rows {
				if val, ok := row.Attrs[prop].(string); ok {
					counts[val]++
				}
			}
			if len(counts) > 0 {
				fmt.Fprintf(&b, "  %s:\n", prop)
				for _, val := range sortedKeys(counts) {
					fmt.Fprintf(&b, "    %-20s %d\n", val, counts[val])
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
	Rows       []NodeRow // every node in document order (reqs + items)
	ReqCount   int
	ItemCount  int
}

// NodeRow is a single row in list/stats output: a requirement or an
// information item (container/info), in document order.
type NodeRow struct {
	Attrs map[string]any
	ID    string
	Title string
	Kind  model.Kind
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
	Documents   []jsonDocSection `json:"documents"`
	ParseErrors []jsonParseErr   `json:"parse_errors"`
	Summary     jsonSummary      `json:"summary"`
	Version     int              `json:"version"`
	ExitCode    int              `json:"exit_code"`
}

type jsonSummary struct {
	Filter      string `json:"filter,omitempty"`
	Total       int    `json:"total"`
	Valid       int    `json:"valid"`
	Invalid     int    `json:"invalid"`
	ParseErrors int    `json:"parse_errors"`
	Warnings    int    `json:"warnings"`
}

type jsonDocSection struct {
	Path     string          `json:"path"`
	Title    string          `json:"schema_title"`
	Reqs     []jsonReqResult `json:"requirements"`
	ReqCount int             `json:"req_count"`
}

type jsonReqResult struct {
	ID     string    `json:"id"`
	Checks []jsonChk `json:"checks,omitempty"`
	Valid  bool      `json:"valid"`
}

type jsonChk struct {
	Level     string `json:"level"`
	Code      string `json:"code,omitempty"`
	Direction string `json:"direction,omitempty"`
	Outcome   string `json:"outcome,omitempty"`
	Source    string `json:"source,omitempty"`
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
	Rows       []jsonListRow `json:"rows"`
}

type jsonListRow struct {
	Attrs map[string]any `json:"attrs,omitempty"`
	Type  string         `json:"type"`
	ID    string         `json:"id,omitempty"`
	Title string         `json:"title,omitempty"`
	Body  string         `json:"body,omitempty"`
}

type jsonStatsDoc struct {
	AttributeStats map[string]map[string]int `json:"attribute_stats"`
	Path           string                    `json:"path"`
	ReqCount       int                       `json:"req_count"`
	ItemCount      int                       `json:"item_count"`
}

// FormatJSON serializes the full check report as JSON.
func (r *Report) FormatJSON() string {
	warnCount, errorCount := r.counts()

	// Count invalid (pass 1 + graph ERROR)
	invalid := len(r.ValErrors) + errorCount

	jr := jsonReport{
		Version:  1,
		ExitCode: r.ExitCode(),
		Summary: jsonSummary{
			Total:       r.TotalReqs,
			Valid:       r.ValidReqs,
			Invalid:     invalid,
			ParseErrors: len(r.ParseErrors),
			Warnings:    warnCount,
			Filter:      r.Filter,
		},
		Documents:   make([]jsonDocSection, 0, len(r.DocHeaders)),
		ParseErrors: make([]jsonParseErr, 0, len(r.ParseErrors)),
	}

	for _, dh := range r.DocHeaders {
		doc := jsonDocSection{
			Path:     dh.Path,
			Title:    dh.Title,
			ReqCount: dh.ReqCount,
			Reqs:     make([]jsonReqResult, 0, len(dh.ReqIDs)),
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
					Outcome:   gc.Outcome,
					Source:    gc.File,
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
		jr.ParseErrors = append(jr.ParseErrors, jsonParseErr(pe))
	}

	b, err := json.MarshalIndent(jr, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error":"json marshal: %v"}`, err)
	}
	return string(b) + "\n"
}

// FormatListJSON serializes list output as JSON. All nodes appear in
// document order in the `rows` array with a `type` discriminator.
func FormatListJSON(docs []DocumentSummary) string {
	jr := jsonListReport{
		Documents: make([]jsonListDoc, 0, len(docs)),
	}
	for _, doc := range docs {
		jdoc := jsonListDoc{
			Path:       doc.Path,
			Properties: doc.Properties,
			Rows:       make([]jsonListRow, 0, len(doc.Rows)),
		}
		for _, row := range doc.Rows {
			jrow := jsonListRow{
				Type:  row.Kind.String(),
				ID:    row.ID,
				Title: row.Title,
			}
			if row.Kind == model.KindRequirement {
				jrow.Attrs = row.Attrs
			}
			jdoc.Rows = append(jdoc.Rows, jrow)
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
		Documents         []jsonStatsDoc `json:"documents"`
		TotalRequirements int            `json:"total_requirements"`
		TotalDocuments    int            `json:"total_documents"`
	}
	jr := jsonStatsOut{
		TotalRequirements: totalReqs,
		TotalDocuments:    len(docs),
	}
	for _, doc := range docs {
		jdoc := jsonStatsDoc{
			Path:           doc.Path,
			ReqCount:       doc.ReqCount,
			ItemCount:      doc.ItemCount,
			AttributeStats: make(map[string]map[string]int),
		}
		for _, prop := range doc.Properties {
			counts := map[string]int{}
			for _, row := range doc.Rows {
				if val, ok := row.Attrs[prop].(string); ok {
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
	switch v.(type) {
	case string, bool, int, int64, float64:
		return model.FormatScalar(v)
	case []any:
		b, _ := json.Marshal(v)
		return string(b)
	default:
		return ""
	}
}
