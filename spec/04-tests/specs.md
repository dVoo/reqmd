# reqmd Test Specifications

## TST-UNT-001: Core package unit tests
```attr
status: approved
verify: Test
test-type: unit
disposition: implemented
trace:
  - SW-PAR-001
  - SW-SCH-001
  - SW-GRA-001
  - SW-CHK-001
```
Unit tests shall cover the parser, schema, and graph packages with inline fixtures (no external files), covering all documented check conditions, parsing edge cases, schema compilation paths, graph trace checks, and Pass 3 broken-parent validation.

*Rationale:* Inline fixtures keep tests self-contained and fast. These packages constitute the core validation pipeline and require the highest confidence. Each validation check has distinct logic paths exercised independently.

## TST-INT-001: End-to-end integration tests
```attr
status: draft
verify: Test
test-type: integration
disposition: implemented
trace:
  - SW-EXP-001
```
Integration tests shall run the full pipeline end-to-end: parse a multi-level requirement tree (including sub-requirements), validate all attributes, build the graph (including Pass 3 parent-child edges), run all ten Pass 2 checks plus Pass 3 broken-parent validation, and produce correct CSV/HTML/graph output with hierarchical display.

*Rationale:* Integration tests catch cross-package contract violations that unit tests miss. The reqmd's own `spec/` directory serves as the primary integration test fixture, and should exercise the sub-requirement hierarchy.

## TST-FIX-001: Build and test gate
```attr
status: approved
verify: Test
test-type: build
disposition: implemented
trace:
  - SW-CLI-001
```
The project shall build cleanly with `go build ./...` and all tests shall pass with `go test ./...` after every change.

*Rationale:* A build-and-test gate in CI prevents regressions. Existing tests must be maintained and extended as features are added.

## TST-UNT-002: Frontmatter and rendering tests
```attr
status: draft
verify: Test
test-type: unit
disposition: implemented
trace:
  - SW-PAR-001
  - SW-EXP-001
```
Unit tests SHALL verify: frontmatter extraction via goldmark-meta with correct `Document.Meta` population; goldmark body rendering produces correct HTML for inline code, emoji, math, and fenced blocks; search-and-filter JS logic (debounce, AND-combined filters, card hide/show, count update, child card visibility); sub-requirement parsing (dynamic reqLevel, `ParentID` population, `isNextAttrBlock` backward compatibility); and the tree TOC sidebar renders with correct parent/child nesting, toggle expand/collapse, and scroll-spy highlights.

*Rationale:* These features have distinct logic paths requiring dedicated test coverage. Frontmatter parsing touches the parser AST pipeline; sub-requirement parsing adds a new state machine with reqLevel discovery and parentID tracking; tree TOC and search/filter are client-side JS features tested via rendered HTML output.

## TST-VER-001: Verify package tests
```attr
status: approved
verify: Test
test-type: unit
disposition: implemented
trace:
  - SW-VER-001
  - SW-VER-002
  - SW-VER-003
```
Unit tests shall cover the `internal/verify` package: CTRF JSON parsing (status→outcome mapping, flaky→inconclusive, unmapped-test warnings, `~N` pin preservation), multi-file latest-wins merge by timestamp, manual-results loading via `parser.Discover`, non-CTRF JSON skip, and pseudo-requirement synthesis (ID, trace with pin, outcome attr).

*Rationale:* The verify package is the entry point for the V&V results feature. Each parsing path (CTRF, manual, auto-detection) has distinct logic that must be exercised independently. The merge semantics (latest-wins, pin stripping) are load-bearing for correctness.

## TST-VER-002: Outcome-gated graph checks tests
```attr
status: approved
verify: Test
test-type: unit
disposition: implemented
trace:
  - SW-VER-004
```
Unit tests shall cover the `missing-verdict` and `failing-verdict` graph checks: no-op when no result nodes present, missing-verdict fires for approved measures without results, missing-verdict skips draft measures, failing-verdict fires for fail outcome at ERROR, failing-verdict does not fire for pass, and version-pin fires on stale result pins.

*Rationale:* The outcome-gated checks are the core of the V-model right-side closure. Each check has status-gated and outcome-gated branches that must be independently verified to prevent false positives and false negatives.

## TST-SRV-001: Serve with results integration tests
```attr
status: approved
verify: Test
test-type: integration
disposition: implemented
trace:
  - SW-SRV-001
  - SW-VER-003
  - SW-VER-004
```
Integration tests shall verify that `serve --results <path>` loads verification results, renders verdict badges in the HTML output, watches result file paths for changes (both fsnotify and polling fallback), and triggers a rebuild when result files change. Tests shall confirm that CTRF JSON changes update verdict badges from pass to fail and that the `failing-verdict` check fires on rebuild.

*Rationale:* The `serve --results` feature combines the live-reload watcher with the verify pipeline. Both the fsnotify and polling watcher paths must be exercised to ensure result-file changes trigger rebuilds in both modes. The verdict badge rendering must be verified in the HTML output, not just the graph checks.
