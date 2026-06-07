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
