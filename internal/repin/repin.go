// Package repin implements the `reqmd repin` command: it computes
// version-pin changes (proposed or applied) for a reqmd spec tree.
//
// The package separates computation from application: Build parses
// the tree and produces a sorted list of RepinDelta values; Apply
// takes those deltas and rewrites the source files in place, scoped
// to ```attr blocks only so surrounding Markdown prose is left
// untouched.
package repin

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"reqmd/internal/graph"
	"reqmd/internal/model"
	"reqmd/internal/parser"
)

// Result is the value Build returns: a sorted list of deltas plus a
// per-file summary derived from the same list. The caller passes
// Deltas to Apply or to one of the Format* functions.
//
// The JSON field names form the stable contract for the
// `reqmd repin --json` interface.
type Result struct {
	Deltas   []graph.RepinDelta `json:"deltas"`
	ByFile   map[string]int     `json:"by_file"`
	Outdated int                `json:"outdated"`
	Unpinned int                `json:"unpinned"`
	Predated int                `json:"predated"`
}

// ApplyReport describes the outcome of a repin apply step. Applied
// is the number of ref replacements actually written to disk (a
// delta whose SourceRef wasn't found in any attr block is a no-op
// and not counted). Skipped holds the predated deltas, which repin
// never auto-fixes. Files is the number of distinct spec files
// actually rewritten.
type ApplyReport struct {
	Applied int                `json:"applied"`
	Skipped []graph.RepinDelta `json:"skipped"`
	Files   int                `json:"files"`
}

// Build parses the spec tree at root and returns the list of
// version-pin changes that repin would propose (or apply, when
// followed by Apply). When promoteUnpinned is true, refs without
// a pin against a versioned upstream are also proposed for pinning
// — converting "no claim" into "claimed at current version".
//
// The returned Result is sorted (delegated to graph.RepinDeltas).
func Build(root string, promoteUnpinned bool) (Result, error) {
	docs, err := parser.DiscoverMeta(root)
	if err != nil {
		return Result{}, fmt.Errorf("discovering docs: %w", err)
	}
	g, err := graph.New(docs)
	if err != nil {
		return Result{}, fmt.Errorf("building graph: %w", err)
	}
	deltas := g.RepinDeltas(promoteUnpinned)
	if deltas == nil {
		deltas = []graph.RepinDelta{}
	}
	res := Result{Deltas: deltas, ByFile: make(map[string]int)}
	for _, d := range deltas {
		res.ByFile[d.File]++
		switch d.Kind {
		case "outdated":
			res.Outdated++
		case "unpinned":
			res.Unpinned++
		case "predated":
			res.Predated++
		}
	}
	return res, nil
}

// Apply rewrites each file containing at least one outdated or
// unpinned delta. Predated deltas are reported back in the Skipped
// slice and never applied — they are data-integrity errors that
// require manual review.
//
// The rewrite is scoped to ```attr blocks: each block's contents
// are scanned for the SourceRef substring, which is replaced with
// the new ref form (REF~newPin, or REF~newPin if the ref was
// previously unpinned). All other bytes of the file are preserved
// verbatim.
//
// On success each touched file is rewritten via os.WriteFile. If a
// write fails, the function returns the error; the apply report
// lists every delta that was successfully applied.
func Apply(deltas []graph.RepinDelta) (ApplyReport, error) {
	report := ApplyReport{}

	// Predated deltas are never auto-fixed. skip is pre-allocated so
	// the JSON marshal emits [] (not null) when there are none.
	apply := make([]graph.RepinDelta, 0, len(deltas))
	skip := make([]graph.RepinDelta, 0)
	for _, d := range deltas {
		if d.Kind == "predated" {
			skip = append(skip, d)
			continue
		}
		apply = append(apply, d)
	}

	// Group by file to minimise reads.
	byFile := make(map[string][]graph.RepinDelta)
	for _, d := range apply {
		byFile[d.File] = append(byFile[d.File], d)
	}

	for path, fileDeltas := range byFile {
		src, err := os.ReadFile(path)
		if err != nil {
			return report, fmt.Errorf("reading %s: %w", path, err)
		}
		updated, matches, err := applyToFile(src, fileDeltas)
		if err != nil {
			return report, fmt.Errorf("rewriting %s: %w", path, err)
		}
		if matches == 0 {
			// No ref was actually replaced; leave the file untouched
			// and don't count it toward Files.
			continue
		}
		if err := os.WriteFile(path, updated, 0o644); err != nil {
			return report, fmt.Errorf("writing %s: %w", path, err)
		}
		report.Applied += matches
		report.Files++
	}
	report.Skipped = skip
	return report, nil
}

// applyToFile rewrites the attr-block portions of src so that each
// delta's SourceRef is replaced with the new ref form. A delta is
// considered a no-op if its SourceRef isn't found in any attr block
// (or every match is at a non-boundary position); in that case the
// file is returned unchanged.
//
// Returns the (possibly rewritten) bytes and the total number of
// ref replacements actually made across all attr blocks.
//
// Matching is whole-ref: SourceRef "UP-001" only matches when the
// ref is at a list-element boundary (followed by `,`, `]`, or
// whitespace) AND is not the prefix of a longer ref (not followed
// by `~` for the unpinned-promote case). This prevents a naive
// substring match from corrupting an already-pinned ref like
// `UP-001~3` when the same ref is being promoted from unpinned.
func applyToFile(src []byte, deltas []graph.RepinDelta) ([]byte, int, error) {
	ranges := locateAttrBlocks(src)
	if len(ranges) == 0 {
		return src, 0, nil
	}
	// Build the (oldRef, newRef) pairs up front so the per-block
	// pass can scan each pair without re-parsing the deltas.
	pairs := buildPairs(deltas)
	if len(pairs) == 0 {
		return src, 0, nil
	}

	// Process ranges in reverse so byte offsets remain valid as we
	// mutate the buffer.
	out := src
	totalMatches := 0
	for i := len(ranges) - 1; i >= 0; i-- {
		r := ranges[i]
		block := out[r.start:r.end]
		newBlock, matches := applyPairsToBlock(block, pairs)
		totalMatches += matches
		if matches > 0 {
			out = bytesSliceReplace(out, r.start, r.end, newBlock)
		}
	}
	return out, totalMatches, nil
}

// refPair is a pre-computed (oldRef, newRef) pair ready for byte
// matching. oldRef and newRef are []byte (not strings) to avoid
// repeated []byte() conversions inside the hot scan loop.
type refPair struct {
	old, new []byte
}

// buildPairs projects a slice of RepinDelta into refPairs, dropping
// any delta whose old ref already equals the new ref (no-op).
func buildPairs(deltas []graph.RepinDelta) []refPair {
	pairs := make([]refPair, 0, len(deltas))
	for _, d := range deltas {
		oldRef := d.SourceRef
		newRef := newRefFrom(d)
		if oldRef == newRef {
			continue
		}
		pairs = append(pairs, refPair{old: []byte(oldRef), new: []byte(newRef)})
	}
	return pairs
}

// applyPairsToBlock scans block and emits a new []byte with every
// matching ref replaced. For each position it finds the earliest
// next match across all pairs, copies the unmatched run verbatim,
// then emits the replacement. Returns the new bytes and the number
// of replacements made. When no pair matches, returns block
// unchanged (zero allocation) and a count of 0.
//
// Cost note: this is N sub-scans into one output buffer, where N is
// len(pairs) — not a true single pass. For the typical repin
// workload (a handful of deltas, small attr blocks) this is fine;
// an Aho-Corasick automaton would be the single-pass alternative but
// is overkill here.
func applyPairsToBlock(block []byte, pairs []refPair) ([]byte, int) {
	// Fast path: if no pair matches at all, return block unchanged
	// so the caller skips the splice and we avoid a copy.
	hasMatch := false
	for _, p := range pairs {
		if bytes.Contains(block, p.old) {
			hasMatch = true
			break
		}
	}
	if !hasMatch {
		return block, 0
	}

	out := make([]byte, 0, len(block))
	pos := 0
	matches := 0
	for pos < len(block) {
		bestIdx, bestEnd, bestNew := -1, 0, []byte(nil)
		for _, p := range pairs {
			rel := bytes.Index(block[pos:], p.old)
			if rel < 0 {
				continue
			}
			abs := pos + rel
			end := abs + len(p.old)
			if !isRefBoundary(block, abs, end) {
				continue
			}
			if bestIdx == -1 || abs < bestIdx {
				bestIdx, bestEnd, bestNew = abs, end, p.new
			}
		}
		if bestIdx < 0 {
			out = append(out, block[pos:]...)
			break
		}
		out = append(out, block[pos:bestIdx]...)
		out = append(out, bestNew...)
		pos = bestEnd
		matches++
	}
	return out, matches
}

// isRefBoundary reports whether the match at [start, end) inside buf
// is a whole ref (not a substring of a longer token). Both edges are
// checked:
//   - The byte before the match must be a list-element starter (`[`,
//     `,`, whitespace, or start-of-block) so we don't match "UP-001"
//     inside "XUP-001".
//   - The byte after the match must be a list-element terminator (` `,
//     `\t`, `\n`, `\r`, `,`, `]`, or end-of-block) so we don't match
//     "UP-001" inside "UP-001X".
//   - The byte after must NOT be `~` — that would mean we matched the
//     prefix of a longer ref (e.g. "UP-001" inside "UP-001~3"), which
//     must never be rewritten.
func isRefBoundary(buf []byte, start, end int) bool {
	// Start boundary: the byte before must be a list starter or BOS.
	if start > 0 {
		before := buf[start-1]
		switch before {
		case ' ', '\t', '\n', '\r', ',', '[':
			// ok
		default:
			return false
		}
	}
	// End boundary.
	if end >= len(buf) {
		return true
	}
	after := buf[end]
	if after == '~' {
		return false
	}
	switch after {
	case ' ', '\t', '\n', '\r', ',', ']':
		return true
	}
	return false
}

// newRefFrom computes the new source-form ref string for a delta.
// The transform is:
//   - "doc-id/ID~1" + newPin=3 → "doc-id/ID~3"
//   - "ID~1"       + newPin=3 → "ID~3"
//   - "doc-id/ID"  + newPin=3 → "doc-id/ID~3"  (unpinned promote)
//   - "ID"         + newPin=3 → "ID~3"          (unpinned promote)
func newRefFrom(d graph.RepinDelta) string {
	bare, _, _ := model.StripPin(d.SourceRef)
	return bare + "~" + strconv.Itoa(d.NewPin)
}

// attrRange is an inclusive [start, end) byte range covering the
// inner YAML of an attr block (NOT including the ``` fences).
type attrRange struct{ start, end int }

// locateAttrBlocks scans src for ```attr\n ... \n``` fences and
// returns the inner-YAML byte ranges. Blank lines between the
// opening fence and the first YAML line are skipped over by the
// parser; the same applies here: we treat the byte immediately
// after the opening fence's newline as the range start.
//
// The match is anchored to the literal string "```attr" followed
// by an optional space/info segment. This matches the parser's
// attrFenceInfo = "attr" convention exactly.
func locateAttrBlocks(src []byte) []attrRange {
	const open = "```attr"
	const close = "```"
	var ranges []attrRange
	for i := 0; i+len(open) <= len(src); {
		idx := bytes.Index(src[i:], []byte(open))
		if idx < 0 {
			break
		}
		openStart := i + idx
		// Require the fence to be on its own logical line: the byte
		// immediately before is either a newline or the start of file.
		if openStart > 0 && src[openStart-1] != '\n' {
			i = openStart + 1
			continue
		}
		// Find the end of the opening fence line.
		eol := bytes.IndexByte(src[openStart:], '\n')
		if eol < 0 {
			break
		}
		contentStart := openStart + eol + 1
		// Find the closing fence. The closing fence is a line that
		// begins with ``` (optionally followed by a language tag we
		// ignore) and is followed by a newline or EOF.
		rest := src[contentStart:]
		closeIdx := bytes.Index(rest, []byte("\n"+close))
		if closeIdx < 0 {
			// EOF or no closing fence — leave the rest as-is.
			break
		}
		contentEnd := contentStart + closeIdx
		ranges = append(ranges, attrRange{start: contentStart, end: contentEnd})
		i = contentEnd + len(close) + 1
	}
	return ranges
}

// bytesSliceReplace returns a new []byte with src[start:end]
// replaced by replacement. Single allocation via nested appends.
func bytesSliceReplace(src []byte, start, end int, replacement []byte) []byte {
	return append(append(append([]byte{}, src[:start]...), replacement...), src[end:]...)
}

// stripPinForDisplay removes a trailing ~N from a ref so the dry-run
// diff reads naturally. Delegates to model.StripPin.
func stripPinForDisplay(ref string) string {
	bare, _, _ := model.StripPin(ref)
	return bare
}

// FormatText renders the Result as the human-readable change list
// shown by `reqmd repin` in dry-run mode. The format is:
//
//	spec/path/to/file.md  REQ-ID  REF → REF~newPin
//	...
//	N outdated + M unpinned + P predated refs across K files.
func FormatText(res Result) string {
	if len(res.Deltas) == 0 {
		return "no version-pin changes needed\n"
	}
	var b strings.Builder
	for _, d := range res.Deltas {
		oldRef := stripPinForDisplay(d.SourceRef)
		fmt.Fprintf(&b, "%s  %s  %s → %s\n",
			filepath.ToSlash(d.File),
			d.ReqID,
			oldRef,
			newRefFrom(d),
		)
	}
	fmt.Fprintf(&b, "%d outdated, %d unpinned, %d predated across %d files.\n",
		res.Outdated, res.Unpinned, res.Predated, len(res.ByFile))
	return b.String()
}

// FormatTextApplied renders an ApplyReport as a one-line summary.
// Used after a successful --yes run.
func FormatTextApplied(rep ApplyReport) string {
	if rep.Applied == 0 {
		return "no changes applied\n"
	}
	return fmt.Sprintf("applied %d changes across %d files\n", rep.Applied, rep.Files)
}

