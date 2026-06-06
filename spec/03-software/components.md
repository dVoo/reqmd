# reqmd Software Components

## SW-PAR-001: Goldmark-based markdown parser
```attr
status: approved
package: parser
priority: Critical
trace:
  - SYS-FMT-001
  - SYS-FMT-004
```
The parser package shall use goldmark to walk the Markdown AST, extracting requirement ID from headings and YAML attributes from fenced ` ```attr ``` ` blocks. The first heading+attr pair encountered determines the requirement level (`reqLevel`); headings at `reqLevel` with an adjacent attr block are top-level requirements, and headings at `reqLevel+1` with an adjacent attr block are sub-requirements (`ParentID` set to the active parent). Non-attr headings are body text. The parser SHALL also extract YAML frontmatter from `.md` files, with first-file-wins merge semantics across files in the same document directory. Parsing shall run in parallel across files via a worker pool sized to `runtime.NumCPU()`.

*Rationale:* Dynamic `reqLevel` discovery decouples requirement detection from hardcoded heading levels. The `isNextAttrBlock()` lookahead provides a universal gate. First-file-wins merge avoids metadata conflicts. Parallel worker pools maximize throughput on multi-document repositories.

## SW-SCH-001: JSON Schema compilation
```attr
status: approved
package: schema
priority: Critical
trace:
  - SYS-FMT-002
  - SYS-FMT-003
  - SYS-TRC-001
```
The schema package shall compile `schema.yaml` via `github.com/google/jsonschema-go` (JSON Schema 2020-12), inject six built-in attribute definitions (`trace`, `version`, `root`, `leaf`, `disposition`, `disposition-reason`) before validation, and provide property introspection via direct `map[string]any` access. The parent-child hierarchy is structural (from parser heading detection), not schema-driven — no built-in attribute is needed.

*Rationale:* The Google jsonschema-go library is the official JSON Schema 2020-12 implementation for Go with zero external dependencies. Direct map access avoids marshal/unmarshal overhead on the hot validation path. Keeping parent-child out of the schema keeps the validation layer focused on project-specific attributes.

## SW-GRA-001: In-memory directed graph
```attr
status: approved
package: graph
priority: Critical
trace:
  - SYS-VAL-001
```
The graph package shall build a pure Go in-memory directed graph from all parsed requirements via three sub-passes: Pass 1 creates `CachedNode` structs with typed fields (ReqID, Root, Leaf, IsChild, Disposition, External, IDPrefix, Inbound/Outbound slices); Pass 2 builds adjacency from manual `trace` attributes; Pass 3 builds parent-child edges from the parser-set `ParentID` field. Broken parent references in Pass 3 produce ERROR-level results. No external graph database is used at runtime.

*Rationale:* Three sub-passes guarantee deterministic node initialization before edge construction. The `IsChild` flag propagates into untraced-check suppression. Pure Go adjacency eliminates temp-directory I/O and serialization overhead.

## SW-CHK-001: Ten trace integrity checks
```attr
status: approved
package: graph
priority: Critical
trace:
  - SYS-FMT-002
  - SYS-TRC-001
```
The graph package shall implement ten Pass 2 trace checks: broken reference (WARNING), circular (ERROR), untraced (WARNING — suppressed for sub-requirements via `!node.IsChild`), no-downstream (WARNING), disposition without reason (WARNING), mandatory disposition (ERROR), ID prefix mismatch (ERROR), ID prefix collision (ERROR), duplicate ID (ERROR), and ambiguous reference (ERROR). ERROR-level results produce exit code 1; WARNING and INFO do not. Broken parent references from Pass 3 also produce ERROR-level results collected into the same report.

*Rationale:* ISO 26262 requires demonstrated absence of broken or circular traces. ERROR-level checks enforce structural integrity; WARNING-level checks flag maintainability concerns without blocking CI. Untraced suppression for sub-requirements avoids false positives — children are expected to be referenced by their parent, not by external traces.

## SW-EXP-001: CSV, HTML, and graph export
```attr
status: approved
package: exporter
priority: High
trace:
  - SYS-CLI-001
  - SYS-HTM-001
```
The exporter package shall produce three output formats: CSV (attribute columns with `bufio.Writer` streaming), HTML (standalone page with card-based layout, goldmark-rendered body and rationale, upstream/downstream trace columns, doc-level trace chain tab strip, search-and-filter toolbar, persistent 240px fixed left sidebar tree TOC with collapsible parent nodes and scroll-spy highlighting, nested sub-requirement cards inside their parent, mobile hamburger toggle, and theme toggle), and graph (persistent graph database in ladybugdb format for external Cypher query tools).

The HTML export SHALL render requirement body and rationale through goldmark to support inline Markdown syntax (code spans, emphasis, emoji, math, fenced code blocks) rather than plain text escaping.

*Rationale:* CSV covers bulk data exchange; HTML supports human traceability review with full Markdown rendering and navigation features. Nested child cards preserve V-model decomposition visually. Graph export via ladybugdb provides a persistent queryable graph database for the planned `reqmd-query` standalone analysis tool.

## SW-CLI-001: Cobra command dispatch
```attr
status: approved
package: cmd
priority: Critical
trace:
  - SYS-CLI-001
```
The cmd package shall use cobra to dispatch subcommands and coordinate the full pipeline: parse documents, compile schemas, validate attributes, build the graph, run trace checks, and format the report to stdout with the appropriate exit code.

*Rationale:* Cobra provides standard CLI conventions (help text, flags, subcommands). A single coordinator function ensures consistent pipeline execution across all commands.

