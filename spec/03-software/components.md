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


## SW-ERR-001: Parser discovery error handling
```attr
status: approved
package: parser
priority: High
trace:
  - SYS-FMT-001
```
The parser package shall surface discovery errors (unreadable directories, malformed schema.yaml, unparseable .md) as `ParseError` entries in the report rather than aborting the run, so that a single bad file does not mask the rest of the tree. `Discover` walks the tree, locates every directory containing a `schema.yaml`, and parses all `.md` files in that directory in parallel; per-file failures are collected and reported with file context.

*Rationale:* Partial-failure reporting lets CI surface every problem in one run instead of failing fast on the first error and hiding subsequent issues. Parallel parsing keeps large trees fast; error collection preserves the worker-pool's throughput benefit.

## SW-SAF-001: Version-pin demotion in check
```attr
status: approved
package: cmd
priority: Medium
trace:
  - SYS-VAL-001
```
The `check` command shall accept a `--relaxed-versions` flag that demotes `version-pin` findings with `direction: "outdated"` (upstream newer than the pin) from ERROR to WARNING. Predated findings (pin ahead of upstream) remain ERROR regardless of the flag, since they indicate a downstream referencing a version that does not exist on the upstream.

*Rationale:* Outdated pins are often intentional (a downstream validated against a known-good older version and has not yet revalidated); demoting them to WARNING lets teams adopt version pinning without blocking CI on every upstream bump. Predated pins are data-integrity errors and must always block.

## SW-SAF-002: Graph construction safety
```attr
status: approved
package: graph
priority: High
trace:
  - SYS-VAL-001
```
The graph package's `New` constructor shall build the in-memory directed graph from parsed requirements without performing I/O or side effects, so that graph construction is deterministic and cannot fail at runtime. Node creation, trace-edge building, and parent-child edge building run as three ordered sub-passes over the already-parsed in-memory document set.

*Rationale:* Keeping graph construction pure (no file I/O, no network, no allocations beyond the adjacency slices) means the graph can be rebuilt cheaply on every file change in `serve` mode and unit-tested without fixtures. The three-sub-pass ordering guarantees nodes exist before edges are wired.

## SW-SEC-001: Exporter trace-link resolution
```attr
status: approved
package: exporter
priority: High
trace:
  - SYS-HTM-001
```
The exporter's `ExportWithTraces` shall resolve cross-document trace links via a `TraceResolver` that maps requirement IDs to output-relative HTML paths, so that rendered HTML contains only safe relative links (no absolute filesystem paths, no unescaped user-controlled content in href attributes). Link targets that do not resolve to a known requirement are omitted rather than emitted as broken anchors.

*Rationale:* HTML export is the primary human-review artifact and may be served over HTTP or opened from the filesystem. Relative-only links keep the output portable; omitting unresolvable targets avoids dead anchors and prevents path-traversal-style hrefs from reaching the rendered page.

## SW-SRV-001: Live-reload serve subcommand
```attr
status: approved
package: cmd
priority: Medium
trace:
  - SYS-CLI-001
  - SYS-HTM-001
```
The `serve` subcommand shall watch a requirements directory tree and serve a live-reloading HTML preview over HTTP. On any `.md` or `schema.yaml` change (create, write, remove, or rename) detected via fsnotify — or via polling fallback when fsnotify is unavailable — it re-parses, rebuilds the trace graph, re-checks graph-level invariants, re-exports all documents, and pushes a reload event to connected browsers via Server-Sent Events. `serve` runs graph-level checks (trace refs, cycles, coverage) but does not re-run JSON Schema validation; the status line reports graph-check counts, not per-requirement validity.

*Rationale:* A live preview shortens the author→review feedback loop for the HTML output. fsnotify gives sub-second response on supported platforms; the polling fallback keeps `serve` usable on network filesystems and in containers. Separating graph checks from full validation keeps rebuilds fast enough for interactive use.

## SW-BAS-001: Baseline diff via git tags
```attr
status: approved
package: cmd
priority: High
trace:
  - SYS-BAS-001
```
The `baseline diff <tag1> <tag2>` subcommand shall compare requirement specifications between two git tags without a working-tree checkout, by extracting the repository at each tag via `git archive | tar`, parsing both versions with the standard pipeline, and producing a semantic diff of requirements (added, removed, modified with attribute-level detail) and schemas (new/removed properties, changed required fields). It shall also report added, removed, and updated git submodules via `git ls-tree`. Output is colored text by default with `--json` for structured output.

*Rationale:* Requirement-level diffs drive change-impact analysis and release notes. Using git tags as baseline anchors inherits the team's existing version-control workflow; `git archive` avoids checking out tags into separate worktrees, keeping the operation fast and side-effect-free. Submodule diff surfaces pinned-dependency drift that requirement diffs alone miss.

## SW-CHK-002: Scoped check pipeline
```attr
status: draft
package: cmd
priority: Medium
trace:
  - SYS-VAL-001
```
The `check` command shall accept a `--scope` flag that restricts validation to a bounded subgraph around seed requirement IDs: a lightweight discovery pass collects IDs and traces across the whole tree, seeds are resolved, the bounded subgraph (seeds plus one hop upstream and downstream) is computed, only in-scope documents are full-parsed and Pass-1-validated, and the cheap global checks (duplicate ID, ID-prefix) run against the lightweight index. Expensive global checks (circular, requires-trace-from coverage, version-pin) are skipped in scoped mode and the omission is documented in a stderr preamble.

*Rationale:* Scoped checks let developers validate only the requirements affected by a change, making incremental CI fast on large trees. Running the cheap global checks even in scoped mode preserves duplicate-ID and prefix-collision signal without the cost of the full graph build. This requirement is draft: the scoped pipeline is planned but not yet implemented.
