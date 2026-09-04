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

## TST-UNT-002: Frontmatter, parsing, and rendering tests
```attr
status: draft
verify: Test
test-type: unit
disposition: implemented
trace:
  - SW-PAR-001
  - SW-EXP-001
```
Unit tests SHALL verify: frontmatter extraction via goldmark-meta with correct `Document.Meta` population; goldmark body rendering produces correct HTML for inline code, emoji, math, and fenced blocks; search-and-filter JS logic (debounce, AND-combined filters, card hide/show, count update, child card visibility); content-tree parsing (every heading a node, `ParentID` population, container promotion, nothing dropped); and the tree TOC sidebar renders with correct parent/child nesting, toggle expand/collapse, and scroll-spy highlights.

*Rationale:* These features have distinct logic paths requiring dedicated test coverage. Frontmatter parsing touches the parser AST pipeline; content-tree parsing drives the tree assembly and parent-chain semantics; tree TOC and search/filter are client-side JS features tested via rendered HTML output.

## TST-UNT-003: Content-tree and item tests
```attr
status: approved
verify: Test
test-type: unit
disposition: implemented
trace:
  - SW-PAR-002
  - SW-EXP-002
```
Unit tests shall cover the content tree: preamble prose captured as a headingless info node, requirement-less files fully preserved, container headings promoted to `KindContainer` only when they have children, mid-file headings becoming nodes instead of body text, `ParentID` pointing at the nearest ancestor requirement (never a container), thematic breaks kept in bodies, deterministic multi-file merge order, and the chunked parallel path retaining the file preamble. Tests shall also cover item-aware output: `ls`/`stats` Type columns and counts, CSV rows in document order for all node kinds, HTML container sections and info blocks with stable anchors, index entries carrying `type`/`parentId`/`children`, and `--filter` pruning requirements while keeping items.

*Rationale:* The content tree is the parser's core contract — content must never be dropped and nesting must stay faithful to the source. Item-aware output crosses the parser, reporter, and exporter packages, so it needs dedicated end-to-end coverage.

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

Unit tests shall also cover case-keyed binding and evidence roll-up: `extra.x-reqmd.case`/`verifies`/`description` parsing and silent skip of uninstrumented tests, latest-wins per (measure, case) by `stop`, synthesized `TC:<case>` nodes with `trace` edges and description bodies, cased `RESULT:` identities, roll-up over authored and synthesized downstream test cases (mixed evidence with one failing case fails the requirement), draft downstream cases reported as ignored by missing-verdict, filter-aware evidence, deferred/rejected measure exemption, `unbound-result` warnings and their `--ignore-unbound-results` suppression, and the JSON/text `results` section.

*Rationale:* The outcome-gated checks are the core of the V-model right-side closure. Each check has status-gated and outcome-gated branches that must be independently verified to prevent false positives and false negatives. Case-keyed binding and roll-up change the merge, synthesis, and check semantics together, so they need dedicated coverage.

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

## TST-FIL-001: Filter package tests
```attr
status: approved
verify: Test
test-type: unit
disposition: implemented
trace:
  - SW-FIL-001
```
Unit tests shall cover the `internal/filter` package: expression compilation (including rejection of undeclared attributes), matching against requirement attribute maps plus `id`/`title`, per-document nil handling, `FilterDocs`/`MatchingIDs`, and filter-aware coverage (a filtered-out inbound cannot satisfy a filtered-in requirement).

*Rationale:* The filter engine is the gate for scoping `check`, exports, and diffs. Compile-time typo rejection and filter-aware coverage semantics are load-bearing and must be pinned by tests.

## TST-DIS-001: Disjoint-attribute check tests
```attr
status: approved
verify: Test
test-type: unit
disposition: implemented
trace:
  - SW-DIS-001
  - SW-SUP-001
```
Unit tests shall cover the `disjoint-attribute` graph check: zero intersection produces an ERROR, overlapping values pass, empty/absent attribute is exempt, enablement via `--disjoint-check` and `x-reqmd.disjoint-check`, and suppression via `reqmd-suppress`.

*Rationale:* The check gates cross-cutting attribute consistency (variants, safety levels) that other trace checks do not; both the intersection logic and the exemption/suppression paths need coverage.

## TST-STS-001: Status lifecycle tests
```attr
status: approved
verify: Test
test-type: unit
disposition: implemented
trace:
  - SW-STS-001
```
Unit tests shall cover the status lifecycle: missing status defaults to approved, draft requirements do not satisfy coverage, `x-reqmd.additional-status-values` validation (lowercase-only, reserved collision, duplicates), `x-reqmd.ignore-status` making all requirements coverage providers, and the draft-downstreams-ignored coverage message.

*Rationale:* The lifecycle is enforced across schema, graph, and exporter; the coverage-provider gate is the behavior that changes validation outcomes and must be regression-tested.

## TST-LEV-001: Level-typed coverage tests
```attr
status: approved
verify: Test
test-type: unit
disposition: implemented
trace:
  - SW-LEV-001
```
Unit tests shall cover level-typed `requires-trace-from`: a level token expands to every document declaring it, coverage is satisfied by any approved inbound from those documents, document-id tokens still resolve, and unknown tokens produce a WARNING.

Unit tests shall also cover the `x-reqmd.requires-trace-from` document default (SYS-FMT-003): a requirement that does not declare the attribute inherits the document default and is checked against it; a requirement that declares its own attribute overrides the default; `requires-trace-from: []` opts a requirement out even under a non-empty default; an empty document default opts every requirement in the document out; and a document without a default leaves its requirements on the generic boundary-inference fallback.

*Rationale:* Level resolution changes coverage semantics for multi-document layers; the expansion and fallback paths must be verified independently.

## TST-IMP-001: reqmd-import extraction tests
```attr
status: approved
verify: Test
test-type: unit
disposition: implemented
trace:
  - SW-IMP-001
  - SW-IMP-002
  - SW-IMP-003
```
Unit tests shall cover the `reqmd-import` pipeline: language plugin parse/classification for C, C++, Go, Python, and Rust (kinds, method receivers, doc binding, top-level-only scope rules), trace extraction (explicit `reqmd:trace` markers and opt-in heuristic IDs), and the writer's deterministic, idempotent per-package output with provenance attributes and generated-marker guard.

*Rationale:* reqmd-import closes the code side of the V-model; each language plugin has a distinct tree-sitter query and mapping that must be pinned, and the writer's determinism is what makes re-running the extractor safe in CI.
