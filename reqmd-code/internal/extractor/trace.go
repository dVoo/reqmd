// Package extractor holds language-agnostic logic that operates on the
// output of language plugins. Its only responsibility is trace extraction:
// pulling explicit `reqmd:trace` targets and bare requirement IDs out of
// source-code comments and docstrings.
package extractor

import (
	"regexp"
	"strings"
)

// idShapeRe separates requirement IDs from symbol references. It is
// deliberately looser than the schema's trace pattern, because an
// ID-shaped token must reach `reqmd check` as a ref (where the canonical
// pattern is enforced) instead of being mistaken for a missing symbol:
//
//   - authored IDs: an uppercase first segment, at least one hyphen, and
//     uppercase/digit segments (SW-PAR-001, REQ-MATH-ADD, REQ-1,
//     system/REQ-1~2). The numeric-tail check lives in the schema, not here.
//   - generated source-proxy IDs: prefix, package, symbol, and a lowercase
//     7-hex hash (IMP-clang-Parse-479c6b0). At least two middle segments are
//     required so a typo such as SW-imp-0000001 does not pass as generated.
//
// A token matching either branch is passed through as an Item.Ref and
// validated by `reqmd check`; everything else is a symbol reference resolved
// by reqmd-code.
var idShapeRe = regexp.MustCompile(`^([a-z0-9_-]+/)?([A-Z][A-Z0-9]*(-[A-Z0-9]+)+|[A-Z][A-Z0-9]*(-[A-Za-z0-9_]+){2,}-[0-9a-f]{7})(~[0-9]+)?$`)

// traceRe matches an explicit `reqmd:trace` directive line. One target per
// line is canonical; a comma-separated list on one line is also accepted.
//
// At most one leading space is tolerated: doc-comment code examples are
// indented by a tab or four spaces, and must not become real traces.
var traceRe = regexp.MustCompile(`(?m)^ ?reqmd:trace[ \t]+(.+?)\r?$`)

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

// TraceRefs is the classified result of scanning one item's doc text.
type TraceRefs struct {
	// ReqIDs are tokens shaped like requirement IDs. They pass through as
	// exchange Item.Refs and are validated by `reqmd check`.
	ReqIDs []string
	// SymbolRefs are explicit `reqmd:trace <symbol>` targets. reqmd-code
	// resolves them exactly (no prefix matching) and reports an error when
	// one is unknown or ambiguous.
	SymbolRefs []string
	// Malformed are directive targets that cannot be a symbol or an ID
	// (for example a raw sentence with spaces). They are always an error.
	Malformed []string
	// Heuristic are bare requirement IDs found by the opt-in scan. They are
	// untrusted and never satisfy coverage.
	Heuristic []string
	// OptOut records an explicit `reqmd:trace none`, which disables
	// convention-based test linking for this item.
	OptOut bool
}

// ExtractTraces classifies the `reqmd:trace` directives in the given source
// text. The text is expected to be the BindDoc-cleaned comment/docstring of
// a symbol, but the function is language-agnostic.
//
// Returned slices are stable, deduplicated, and preserve first-seen order.
// A token found as a requirement ID never appears in SymbolRefs, and a
// heuristic match never repeats an explicit requirement ID.
//
// useHeuristic enables the broader bare-ID scan (driven by the CLI flag
// `--heuristic-traces`).
func ExtractTraces(text string, useHeuristic bool) TraceRefs {
	var out TraceRefs
	seenID := make(map[string]bool)
	seenSym := make(map[string]bool)
	seenBad := make(map[string]bool)

	for _, m := range traceRe.FindAllStringSubmatch(text, -1) {
		for raw := range strings.SplitSeq(m[1], ",") {
			tok := strings.TrimSpace(raw)
			if tok == "" {
				continue
			}
			switch {
			case strings.ContainsAny(tok, " \t"):
				// A target with whitespace cannot be a symbol or an ID; it
				// is almost always a prose sentence left on the directive.
				if !seenBad[tok] {
					seenBad[tok] = true
					out.Malformed = append(out.Malformed, tok)
				}
			case tok == "none":
				out.OptOut = true
			case idShapeRe.MatchString(tok):
				if !seenID[tok] {
					seenID[tok] = true
					out.ReqIDs = append(out.ReqIDs, tok)
				}
			default:
				if !seenSym[tok] {
					seenSym[tok] = true
					out.SymbolRefs = append(out.SymbolRefs, tok)
				}
			}
		}
	}

	if useHeuristic {
		for _, m := range heuristicRe.FindAllStringSubmatch(text, -1) {
			id := m[1]
			if seenID[id] {
				continue
			}
			seenID[id] = true
			out.Heuristic = append(out.Heuristic, id)
		}
	}

	return out
}
