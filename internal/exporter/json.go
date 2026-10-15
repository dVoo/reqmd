// Package exporter's json.go writes the whole spec tree as flat NDJSON
// tables for external reporting platforms (Databricks, Snowflake, BigQuery).
//
// The output is a pure function of (tree + results): no state, no cursor,
// no live link. Every table is a projection of already-parsed data — no
// parser, no validation, no side effects. Consumers join the four tables
// (requirements, traces, results, findings) in their own warehouse.
package exporter

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"reqmd/internal/graph"
	"reqmd/internal/model"
)

// JSON exports the parsed tree as NDJSON tables. Verdicts set via
// SetVerdicts are emitted by ExportResults; the graph passed to the
// Export* methods is the single source of truth for trace edges,
// verdicts, and findings.
type JSON struct {
	verdicts map[string]VerdictInfo
}

// SetVerdicts stores verification verdicts for JSON export, mirroring CSV.
func (j *JSON) SetVerdicts(v map[string]VerdictInfo) {
	j.verdicts = v
}

// promotedAttrs maps a built-in requirement attribute to the top-level
// NDJSON key it is promoted to, so warehouse SQL can read it directly
// instead of digging into the attrs envelope.
var promotedAttrs = map[string]string{
	"trace":               "trace",
	"requires-trace-from": "requires_trace_from",
	"status":              "status",
	"disposition":         "disposition",
	"disposition-reason":  "disposition_reason",
	"version":             "version",
	"verify":              "verify",
}

// jsonRequirement is one row of requirements.ndjson. Built-in attributes
// are promoted to top-level keys; the remaining declared attributes stay
// in Attrs so nothing is lost.
//
// Every fixed key is present in every row (null when unset) so warehouse
// schema inference — DuckDB, Spark, Snowflake's infer_schema — sees the
// full column set regardless of which row it samples. Attrs is the one
// intentional exception: it is the per-document attribute envelope and is
// null on nodes that declare no extra attributes.
type jsonRequirement struct {
	Attrs             map[string]any `json:"attrs"`
	ID                string         `json:"id"`
	Title             string         `json:"title"`
	DocPath           string         `json:"doc_path"`
	DocTitle          string         `json:"doc_title"`
	Level             string         `json:"level"`
	Type              string         `json:"type"`
	Status            string         `json:"status"`
	Disposition       string         `json:"disposition"`
	DispositionReason string         `json:"disposition_reason"`
	Verify            string         `json:"verify"`
	ParentID          string         `json:"parent_id"`
	Source            string         `json:"source"`
	Trace             []any          `json:"trace"`
	RequiresTraceFrom []any          `json:"requires_trace_from"`
	Suppressions      []string       `json:"suppressions"`
	Version           any            `json:"version"`
}

// ExportRequirements writes one NDJSON row per content node (requirements,
// containers, and info items) across all authored documents. Synthetic
// documents (verification results and test cases) are skipped: the
// requirement table is authored truth.
func (j *JSON) ExportRequirements(w io.Writer, docs []model.Document) error {
	enc := json.NewEncoder(w)

	var rows []jsonRequirement
	for i := range docs {
		doc := &docs[i]
		if doc.Synthetic {
			continue
		}
		level := ""
		if doc.XReqmd != nil {
			level = doc.XReqmd.Level
		}
		walkJSONNodes(doc.Nodes, func(n *model.Node) {
			rows = append(rows, jsonRequirement{
				DocPath:           doc.Path,
				DocTitle:          doc.Title,
				Level:             level,
				ID:                n.ID,
				Title:             n.Title,
				Type:              n.Kind.String(),
				ParentID:          n.ParentID,
				Source:            n.Source,
				Suppressions:      n.Suppressions,
				Status:            attrString(n.Attrs, "status"),
				Disposition:       attrString(n.Attrs, "disposition"),
				DispositionReason: attrString(n.Attrs, "disposition-reason"),
				Verify:            attrString(n.Attrs, "verify"),
				Trace:             attrSlice(n.Attrs, "trace"),
				RequiresTraceFrom: attrSlice(n.Attrs, "requires-trace-from"),
				Version:           n.Attrs["version"],
				Attrs:             remainingAttrs(n.Attrs),
			})
		})
	}

	sort.SliceStable(rows, func(a, b int) bool {
		if rows[a].DocPath != rows[b].DocPath {
			return rows[a].DocPath < rows[b].DocPath
		}
		return rows[a].ID < rows[b].ID
	})

	for i := range rows {
		if err := enc.Encode(&rows[i]); err != nil {
			return fmt.Errorf("writing requirement %s: %w", rows[i].ID, err)
		}
	}
	return nil
}

// jsonTrace is one row of traces.ndjson: one declared trace reference.
// To is the resolved bare requirement ID when resolved, otherwise the raw
// declared target. Ref always holds the raw form as written, and DocID
// holds the `doc-id` qualifier when the ref uses one. Resolved is false
// for a dangling reference; such a ref also appears in findings.ndjson, so
// consumers can count or filter it themselves.
type jsonTrace struct {
	From     string `json:"from"`
	To       string `json:"to"`
	DocID    string `json:"doc_id"`
	Ref      string `json:"ref"`
	File     string `json:"file"`
	Pin      *int   `json:"pin"`
	Resolved bool   `json:"resolved"`
}

// ExportTraces writes one NDJSON row per declared trace reference, read
// from the raw `trace` attribute values and resolved against the graph.
// Both are needed: the attribute carries the declared intent (including
// dangling refs), the graph carries resolution.
func (j *JSON) ExportTraces(w io.Writer, docs []model.Document, g *graph.Graph) error {
	enc := json.NewEncoder(w)

	var rows []jsonTrace
	for i := range docs {
		doc := &docs[i]
		if doc.Synthetic {
			continue
		}
		for _, req := range doc.Requirements() {
			for _, ref := range attrSlice(req.Attrs, "trace") {
				raw, ok := ref.(string)
				if !ok {
					continue
				}
				bare, pin, pinned := model.StripPin(raw)
				resolvedID, resolved := g.ResolveTraceRef(raw)
				to := bare
				if resolved {
					to = resolvedID
				}
				row := jsonTrace{
					From:     req.ID,
					To:       to,
					Ref:      raw,
					File:     req.Source,
					Resolved: resolved,
				}
				if docID, _, found := strings.Cut(bare, "/"); found {
					row.DocID = docID
				}
				if pinned {
					p := pin
					row.Pin = &p
				}
				rows = append(rows, row)
			}
		}
	}

	sort.SliceStable(rows, func(a, b int) bool {
		if rows[a].From != rows[b].From {
			return rows[a].From < rows[b].From
		}
		return rows[a].Ref < rows[b].Ref
	})

	for i := range rows {
		if err := enc.Encode(&rows[i]); err != nil {
			return fmt.Errorf("writing trace %s -> %s: %w", rows[i].From, rows[i].Ref, err)
		}
	}
	return nil
}

// jsonCase is one contributing verification result in a measure's
// evidence set.
type jsonCase struct {
	Case        string `json:"case"`
	Outcome     string `json:"outcome"`
	File        string `json:"file"`
	Description string `json:"description"`
}

// jsonResult is one row of results.ndjson: one rolled-up measure verdict.
type jsonResult struct {
	MeasureID string     `json:"measure_id"`
	Outcome   string     `json:"outcome"`
	Source    string     `json:"source"`
	Verify    string     `json:"verify"`
	Cases     []jsonCase `json:"cases"`
}

// ExportResults writes one NDJSON row per measure with a non-empty
// evidence set. Verdicts come from the graph's strict roll-up over a
// measure's own results and its approved downstream test cases, so the
// value matches what HTML badges and CSV columns show.
func (j *JSON) ExportResults(w io.Writer, g *graph.Graph) error {
	if j.verdicts == nil {
		return nil
	}
	enc := json.NewEncoder(w)

	ids := make([]string, 0, len(j.verdicts))
	for id := range j.verdicts {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		v := j.verdicts[id]
		row := jsonResult{
			MeasureID: id,
			Outcome:   v.Outcome,
			Source:    v.Source,
			Verify:    measureVerify(g.Documents(), id),
		}
		for _, it := range v.Evidence {
			row.Cases = append(row.Cases, jsonCase{
				Case:        it.Case,
				Outcome:     it.Outcome,
				File:        it.File,
				Description: it.Description,
			})
		}
		if err := enc.Encode(&row); err != nil {
			return fmt.Errorf("writing result %s: %w", id, err)
		}
	}
	return nil
}

// jsonFinding is one row of findings.ndjson, mirroring the reporter's
// CheckResult fields. ERROR and WARNING both appear; consumers filter.
type jsonFinding struct {
	Level     string `json:"level"`
	ReqID     string `json:"req_id"`
	Code      string `json:"code"`
	Direction string `json:"direction"`
	Outcome   string `json:"outcome"`
	File      string `json:"file"`
	Message   string `json:"message"`
}

// ExportFindings writes one NDJSON row per graph check finding (the same
// results the text and JSON reporters print). Rows are ordered by
// requirement, then code, then message so repeated exports are identical.
func (j *JSON) ExportFindings(w io.Writer, g *graph.Graph) error {
	enc := json.NewEncoder(w)

	results := g.CheckResults()
	rows := make([]jsonFinding, 0, len(results))
	for _, r := range results {
		rows = append(rows, jsonFinding{
			Level:     r.Level,
			ReqID:     r.ReqID,
			Code:      r.Code,
			Direction: r.Direction,
			Outcome:   r.Outcome,
			File:      r.File,
			Message:   r.Message,
		})
	}

	sort.SliceStable(rows, func(a, b int) bool {
		if rows[a].ReqID != rows[b].ReqID {
			return rows[a].ReqID < rows[b].ReqID
		}
		if rows[a].Code != rows[b].Code {
			return rows[a].Code < rows[b].Code
		}
		return rows[a].Message < rows[b].Message
	})

	for i := range rows {
		if err := enc.Encode(&rows[i]); err != nil {
			return fmt.Errorf("writing finding for %s: %w", rows[i].ReqID, err)
		}
	}
	return nil
}

// walkJSONNodes visits every node in document order, depth-first.
func walkJSONNodes(nodes []*model.Node, fn func(*model.Node)) {
	for _, n := range nodes {
		fn(n)
		walkJSONNodes(n.Children, fn)
	}
}

// remainingAttrs returns a copy of attrs without the promoted built-in
// keys, so a value never appears both top-level and in the envelope.
func remainingAttrs(attrs map[string]any) map[string]any {
	if len(attrs) == 0 {
		return nil
	}
	out := make(map[string]any, len(attrs))
	for k, v := range attrs {
		if _, promoted := promotedAttrs[k]; promoted {
			continue
		}
		out[k] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// attrString returns a string attribute, or "" for absent/non-string.
func attrString(attrs map[string]any, key string) string {
	s, _ := attrs[key].(string)
	return s
}

// attrSlice returns an array attribute in its native form, or nil.
func attrSlice(attrs map[string]any, key string) []any {
	switch v := attrs[key].(type) {
	case []any:
		return v
	case []string:
		out := make([]any, len(v))
		for i, s := range v {
			out[i] = s
		}
		return out
	}
	return nil
}

// measureVerify returns the `verify` attribute of an authored measure, or
// "" when the ID is unknown or the attribute is absent.
func measureVerify(docs []model.Document, id string) string {
	for i := range docs {
		for _, req := range docs[i].Requirements() {
			if req.ID == id {
				return attrString(req.Attrs, "verify")
			}
		}
	}
	return ""
}
