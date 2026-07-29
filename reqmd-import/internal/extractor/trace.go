// Package extractor holds language-agnostic logic that operates on the
// output of language plugins. M2's only responsibility is trace extraction:
// pulling requirement IDs out of source-code comments and docstrings.
package extractor

import (
	"regexp"
	"strings"
)

// validIDRe is a permissive pre-filter for candidate requirement IDs
// extracted from source code. It is intentionally looser than the
// canonical reqmd ID pattern (which requires a digit segment at the end).
//
// This filter is the coarse-grained gate that runs at extraction time;
// the strict, canonical shape check happens later in `reqmd check`,
// where a malformed ID becomes a WARNING-level broken-ref rather than a
// hard error. By accepting any plausible ID shape here, we honor what
// the user wrote in their `reqmd:trace` marker (even if it would later
// fail strict validation) instead of silently dropping it.
//
// The pattern accepts:
//   - Unqualified: REQ-XYZ, REQ-ARITH-001, SYS-AUTH-001, REQ-MATH-ADD
//   - Qualified:   doc-id/REQ-X (e.g. system/REQ-FEAT-007)
//   - Version pin: REQ-XYZ~3
//
// It rejects obviously-wrong shapes like `REQ-foo-bar` (lowercase
// letters) or bare `foo` (no leading uppercase run).
//
// See reqmd/internal/schema/builtin.go for the canonical definition that
// `reqmd check` enforces.
var validIDRe = regexp.MustCompile(`^([a-z0-9_-]+/)?[A-Z][A-Z0-9-]+(~[0-9]+)?$`)

// traceRe matches an explicit `reqmd:trace` directive line. One ID per
// line is canonical; a comma-separated list on one line is also accepted
// as a tolerated fallback.
var traceRe = regexp.MustCompile(`(?m)^reqmd:trace[ \t]+(.+?)$`)

// heuristicRe matches bare requirement IDs in free-form text. Used only
// when the CLI flag --heuristic-traces is set. Requires at least 3 digits
// in the trailing number to avoid false positives like UTF-8, HTTP-2.
//
// Pattern: one uppercase-letter run, then 1+ "-<uppercase/digits>" segments,
// then a "-<digits>" segment with 3+ digits, optionally followed by a
// "~<digits>" version pin. This matches the multi-segment shape of real
// requirement IDs (e.g. REQ-AUTH-001, SYS-LOG-002) without swallowing
// short-numbered prose tokens like UTF-8 or HTTP-2.
var heuristicRe = regexp.MustCompile(`\b([A-Z][A-Z0-9]*(?:-[A-Z0-9]+)*-\d{3,}(?:~\d+)?)\b`)

// ExtractTraces pulls explicit and (optionally) heuristic trace IDs out
// of the given source text. The text is expected to be the BindDoc-cleaned
// comment/docstring of a symbol, but the function is language-agnostic.
//
// Returned slices are stable, deduplicated, and preserve first-seen order.
// Explicit traces take precedence: an ID found both as explicit and as a
// heuristic match appears in explicit only, never in heuristic.
//
// useHeuristic enables the broader bare-ID scan. It is intended to be
// driven by a CLI flag (e.g. `reqmd-import --heuristic-traces`).
func ExtractTraces(text string, useHeuristic bool) (explicit, heuristic []string) {
	seen := make(map[string]bool)

	// 1. Explicit `reqmd:trace` lines.
	for _, m := range traceRe.FindAllStringSubmatch(text, -1) {
		// m[1] may be a single ID or a comma-separated list.
		for _, raw := range strings.Split(m[1], ",") {
			id := strings.TrimSpace(raw)
			if id == "" {
				continue
			}
			if !validIDRe.MatchString(id) {
				// Silently skip invalid IDs (e.g., typos like "REQ-foo"
				// or bare "foo"). The regex is the contract; we don't
				// error on a malformed marker because prose often
				// contains things that look like IDs but aren't.
				continue
			}
			if !seen[id] {
				seen[id] = true
				explicit = append(explicit, id)
			}
		}
	}

	// 2. Heuristic bare-ID scan (opt-in only).
	if useHeuristic {
		for _, m := range heuristicRe.FindAllStringSubmatch(text, -1) {
			if !seen[m[1]] {
				seen[m[1]] = true
				heuristic = append(heuristic, m[1])
			}
		}
	}

	return explicit, heuristic
}
