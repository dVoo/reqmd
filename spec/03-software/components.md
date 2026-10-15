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
The parser package shall use goldmark to walk the Markdown AST, extracting requirement IDs from headings and YAML attributes from fenced ` ```attr ``` ` blocks. The package shall also extract YAML frontmatter from `.md` files, with first-file-wins merge semantics across files in the same document directory, and parse files in parallel via a worker pool sized to `runtime.NumCPU()` with deterministic (file-order) result merging. Heading structure, parent chains, and content capture are defined in SW-PAR-002.

*Rationale:* goldmark provides a battle-tested CommonMark AST. Deterministic merge ordering keeps output stable regardless of goroutine scheduling. Content-tree semantics live in SW-PAR-002.

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
The graph package shall build a pure Go in-memory directed graph from all parsed requirements via three sub-passes: Pass 1 creates `CachedNode` structs with typed fields (ReqID, Root, Leaf, IsChild, Disposition, External, IDPrefix, RequiresTraceFrom, Inbound/Outbound slices); Pass 2 builds adjacency from manual `trace` attributes; Pass 3 builds parent-child edges from the parser-set `ParentID` field. Broken parent references in Pass 3 produce ERROR-level results. No external graph database is used at runtime.

During Pass 1 each node's `requires-trace-from` expectation shall be resolved to its effective value: the requirement's own attribute when declared, otherwise the enclosing document's `x-reqmd.requires-trace-from` default when set. A requirement-level `requires-trace-from: []` and an empty document default both mean "no downstream coverage expected"; a node whose requirement declares nothing in a document without a default keeps `nil` so the generic boundary-inference fallback applies.

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

The `serve` subcommand shall accept a repeatable `--results <path>` flag that loads ephemeral verification results (CTRF JSON or manual-results markdown) via the same pipeline as `check --results`. When results are loaded, verdict badges (pass/fail/skipped/inconclusive) are rendered on measure cards in the HTML output, and the `failing-verdict` and `missing-verdict` graph checks run on every rebuild. Result file paths (directories and individual `.ctrf.json`/`.json` files) shall also be watched for changes — alongside spec files — so that editing a CTRF JSON or manual-results markdown triggers an immediate rebuild and browser refresh.

*Rationale:* A live preview shortens the author→review feedback loop for the HTML output. fsnotify gives sub-second response on supported platforms; the polling fallback keeps `serve` usable on network filesystems and in containers. Separating graph checks from full validation keeps rebuilds fast enough for interactive use. Extending `--results` to `serve` closes the V-model right side in the interactive authoring loop: a failing verdict turns the status line red immediately, without leaving the editor to run `check`.

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

## SW-VER-001: CTRF report parser
```attr
status: approved
package: verify
priority: High
trace:
  - SYS-VAL-002
```
The `internal/verify` package shall parse CTRF JSON reports (top-level `results.tests[]`) and map each test to a measure ID via the `tests[].extra.x-reqmd.id` field. CTRF `status` shall map to reqmd `outcome` as: `passed→pass`, `failed→fail`, `skipped→skipped`, `pending|other→inconclusive`, and any status with `flaky: true→inconclusive`. Tests with no `x-reqmd.id` shall be skipped with a WARNING. The `~N` version pin in the measure ID shall be preserved.

*Rationale:* CTRF is the open standard for JSON test reports. Using its `extra` extension point for the measure ID survives test renaming and avoids brittle name-parsing. Flaky tests are inconclusive because a pass-after-fail is not a clean pass.

## SW-VER-002: Manual-results markdown loader
```attr
status: approved
package: verify
priority: High
trace:
  - SYS-VAL-002
```
The `internal/verify` package shall load manual-results markdown directories via `parser.Discover`, extracting `x-reqmd.outcome`, `x-reqmd.verifier`, `x-reqmd.evidence`, `x-reqmd.verified-at`, and `trace` from each requirement's attr block. Manual-results dirs live outside the spec root and carry their own user-supplied `schema.yaml`.

*Rationale:* Reusing the existing document pipeline for manual results avoids a parallel parser and keeps schema validation uniform. Keeping results outside the spec root ensures a normal `check` never sees them.

## SW-VER-003: Result merge and synthesis
```attr
status: approved
package: verify
priority: High
trace:
  - SYS-VAL-002
```
The `internal/verify` package shall merge all loaded results by (measure, case) key (stripping `~N` pins from the key), keeping the latest result per key by CTRF `tests[].stop` (ms-epoch) or manual `x-reqmd.verified-at`. Each merged result shall be synthesized as a pseudo-requirement with ID `RESULT:<target>`, `trace: [MEASURE-ID~N]` (pin preserved for version-pin checks), `x-reqmd.outcome: <verdict>`, and `status: approved`, then appended to the document slice before `graph.New`. Result attributes (`x-reqmd.outcome`, `x-reqmd.verifier`, `x-reqmd.evidence`, `x-reqmd.verified-at`) are tool-owned and namespaced under `x-reqmd.*` so they never collide with authored schema attributes.

*Rationale:* Collapsing to one result per measure reflects "latest verdict wins" without storing history in-file. The synthetic `RESULT:` prefix lets graph checks distinguish result nodes from authored ones. Preserving the pin on the trace edge makes the existing `version-pin` check detect stale results with no new logic.

## SW-VER-004: Outcome-gated graph checks
```attr
status: approved
package: graph
priority: High
trace:
  - SYS-VAL-002
```
The graph shall run two new checks when result pseudo-requirements are present: `missing-verdict` (WARNING — an approved verification measure, identified by a non-empty `verify` attribute, has no result tracing to it; draft measures are skipped) and `failing-verdict` (ERROR — a measure's latest result has outcome `fail`). Both checks are no-ops when no result nodes exist. Both are suppressible via `reqmd-suppress`.

Each check shall operate on the measure's rolled-up evidence set: its own attached results plus the results attached to approved, in-filter nodes in its inbound trace closure (authored downstream test specs and synthesized test cases). Aggregation is strict — any fail → fail, else inconclusive → inconclusive, else skipped → skipped, else empty → none, else pass — computed bottom-up and memoized so a check pass stays linear. `missing-verdict` messages report the count of draft downstream cases ignored; `failing-verdict` names the failing case(s) and their source file(s). Deferred/rejected measures (expectsVerification) are exempt from both checks. The `MeasureVerdicts` accessor shall expose the rolled-up verdict per node for exporters.

*Rationale:* A measure is defined by the `verify` attribute, not by inbound edges — a measure with no result has no inbound result edge, so the check must key on `verify`. Draft measures are skipped because a draft measure is not yet expected to have results. Roll-up closes the V-model gap where requirements are verified through downstream test specs or CI cases rather than a direct result edge; the severity lattice is associative, so bottom-up memoization keeps large trees linear.

## SW-FIL-001: Attribute filter engine
```attr
status: approved
package: filter
priority: Medium
trace:
  - SYS-FIL-001
```
The `internal/filter` package shall compile a `--filter` expression once at startup into bytecode using expr-lang, reject at compile time any expression referencing an attribute not declared in some `schema.yaml` (global typo check), and evaluate the bytecode against each requirement's attribute map plus the built-in `id` and `title` variables. Per-document attribute absence shall evaluate to nil (no error) via `AllowUndefinedVariables`. The package shall provide a `MatchingIDs` helper and document filtering (`FilterDocs`).

*Rationale:* Compile-once/eval-many keeps per-requirement filtering cheap on large trees; the global schema check turns attribute typos into a startup error instead of silent empty matches.

## SW-DIS-001: Disjoint-attribute graph check
```attr
status: approved
package: graph
priority: Medium
trace:
  - SYS-CHK-003
```
The graph package shall implement `checkDisjointAttribute` (code `disjoint-attribute`), enabled via `SetDisjointAttrs` populated from the `--disjoint-check` flag or `x-reqmd.disjoint-check`. For every trace edge, the source and target requirements must share at least one value of each named array attribute; zero intersection shall produce an ERROR result, and an empty or absent value is exempt. The check shall respect `reqmd-suppress`.

*Rationale:* The check is a pure graph pass over already-built edges, so it adds no new data structures; naming the attribute per invocation keeps it project-configurable without a new built-in.

## SW-STS-001: Status lifecycle enforcement
```attr
status: approved
package: schema
priority: Medium
trace:
  - SYS-STS-001
```
The schema package shall validate `x-reqmd.additional-status-values` (lowercase-only matching `^[a-z][a-z0-9_-]*$`, no reserved built-in collision, no duplicates) and extend the built-in `status` enum per document accordingly. The graph shall gate coverage on status: only `approved` requirements satisfy a `requires-trace-from` expectation, and a missing-coverage message shall append the count of draft downstreams ignored. The exporter shall hide the status filter for documents with `x-reqmd.ignore-status: true`.

*Rationale:* The status axis is enforced at three points — schema (valid extension values), graph (coverage provider gate), and exporter (status filter visibility) — so the lifecycle is consistent across validation, coverage, and rendered output.

## SW-SUP-001: Check suppression
```attr
status: approved
package: graph
priority: Medium
trace:
  - SYS-SUP-001
```
The graph package shall parse a `reqmd-suppress: [<code>...]` attribute into each requirement's suppression set and skip the listed graph checks when collecting results for that requirement.

*Rationale:* Suppression is a per-requirement opt-out evaluated at result-collection time, so it composes with every existing and future graph check without touching each check's logic.

## SW-LEV-001: Level-typed coverage resolution
```attr
status: approved
package: graph
priority: High
trace:
  - SYS-FMT-005
```
The graph package shall index document directories by their `x-reqmd.level` and resolve each `requires-trace-from` token against that index: a token matching a declared level expands to every directory at that level (any approved inbound from any of them satisfies coverage), a token matching a `document-id` resolves as today, and a token matching neither shall produce a WARNING. Coverage tokens inherited from the `x-reqmd.requires-trace-from` document default (resolved onto the node during Pass 1, SYS-FMT-003) shall be resolved through the same index and are indistinguishable from per-requirement declarations downstream of node creation.

*Rationale:* Level indexing is built once at graph construction from data the parser already exposes, so coverage checks stay O(1) per token while remaining correct as directories within a level are added or removed.

## SW-IMP-001: Source-code extraction pipeline
```attr
status: approved
package: extractor
priority: High
trace:
  - SYS-IMP-001
```
The `reqmd-code` extractor shall walk a source tree, dispatch each file to the `lang` plugin claiming its extension, parse symbols in parallel via a worker pool, bind adjacent doc comments, and extract explicit `reqmd:trace` markers plus (opt-in) heuristic bare-ID references from the bound docs. It shall group symbols into per-package documents and hand them to the exchange bundle emitter as native-id-keyed requirements with stable, line-independent IDs.

*Rationale:* Decoupling the walk from parsing (and parsing from emitting) keeps the pipeline independently testable and lets the worker pool absorb per-file tree-sitter cost.

## SW-IMP-002: Language plugin registry
```attr
status: approved
package: lang
priority: High
trace:
  - SYS-IMP-001
```
The `internal/lang` registry shall define the `Language` interface (name, extensions, tree-sitter grammar, embedded query, doc binding, parse) and host self-registering plugins for C, C++, Go, Python, and Rust, blank-imported from the binary's entrypoint. Each plugin shall map its tree-sitter captures to the shared symbol model, populate method receivers from the enclosing type, and emit only top-level symbols (function-local declarations are excluded so same-named locals cannot collide on one ID).

*Rationale:* The registry lets a new language land as a single self-registering package with no changes to the extractor or exchange emitter; scope rules keep generated IDs unique within a package.

## SW-IMP-003: Exchange bundle emitter
```attr
status: approved
package: extractor
priority: High
trace:
  - SYS-IMP-001
```
The extractor shall emit a `reqmd-import/v1` JSON bundle containing one document per package plus a requirement-less root document that carries the umbrella schema. Each package document shall declare its own `id-prefix` (`<id-prefix><package>-`), its title and notice, and the standard attr properties read from `internal/schema/standard_schema.yaml` (the single source of truth for the provenance attribute set). Each requirement shall carry a line-independent `native-id`, `status: approved`, the bound doc comment as its body, explicit trace markers as `trace`, opt-in heuristic matches as `x-reqmd.heuristic-trace`, and provenance attrs (`x-reqmd.imported`, `x-reqmd.source-file`, `x-reqmd.source-line`, `x-reqmd.symbol-kind`); methods shall point at their parent type by native ID.

The emitter shall order symbols deterministically by (file, start line, name) so repeated runs produce byte-identical output, and shall skip a duplicate `(package, symbol, receiver)` — including a symbol declared twice under mutually exclusive build tags — with a warning rather than emitting two requirements with the same ID.

*Rationale:* The emitter is the boundary between code analysis and file syntax: it produces a data structure, not Markdown, so a determinism bug surfaces as a diff in the bundle instead of a corrupted requirement file.

## SW-IMP-004: Exchange format contract
```attr
status: approved
package: exchange
priority: High
trace:
  - SYS-IMP-002
```
The `reqmd-import/exchange` package shall define the `reqmd-import/v1` bundle types and be the shared contract between producers and the renderer, so the format cannot drift between them. Decoding shall reject unknown fields and trailing data (a producer typo fails loudly) and shall preserve numeric attribute values exactly. Building a bundle shall reject: an unsupported `format` value, an empty document list, a requirement without `native-id`, a requirement in a document without `id-prefix`, a `native-id` that is not unique across the bundle, a `parent` that is not in the same document, a `native-trace` that resolves to no requirement, an attr that redefines the built-in `status` or `trace`, a document directory that escapes the output root, a parent cycle, and two documents claiming the same output path.

*Rationale:* Every producer (a Go binary, a Python script, a future ReqIF reader) is untrusted input from the renderer's point of view. Failing at the contract boundary gives a precise message about the bundle instead of a malformed requirement file, and a shared types package makes producer/renderer drift a compile error rather than a runtime surprise.

## SW-IMP-005: Document renderer
```attr
status: approved
package: writer
priority: High
trace:
  - SYS-IMP-002
```
The renderer shall resolve each requirement's ID to `<id-prefix><native-id>` and rewrite `native-trace` refs to the referenced requirement's full ID. It shall order requirements so every child directly follows its parent and set each requirement's heading depth from its position in the parent chain (`##` for top-level, one more `#` per level), independent of the order the producer listed them in. It shall render the attr block with `status` first, the remaining attributes in sorted key order, and the trace list last, quoting any value containing YAML-significant characters and rendering flat lists as block sequences; it shall reject a nested map or list attribute. Output shall be byte-deterministic and idempotent: a file is rewritten only when its content changes, a generated file carries the `<!-- reqmd: generated -->` marker (plus the pre-rename `reqmd-code` marker for trees generated before the split), and content without a marker is never touched.

*Rationale:* One renderer is what makes the exchange format worth having: quoting, heading depth, and determinism live in a single implementation, so a new importer cannot invent a second, subtly different Markdown dialect, and re-running an import in CI is a safe no-op that leaves hand-authored files alone.

## SW-IMP-006: Document schema writer
```attr
status: approved
package: writer
priority: High
trace:
  - SYS-IMP-002
```
For every document in a bundle the renderer shall emit a `schema.yaml` built from the document's `properties` and `required` lists plus its `x-reqmd` metadata (`level`, `document-id`, `id-prefix`, `upstream`, `requires-trace-from`, `external`), and the JSON-Schema header fields (`$schema`, `$id`, `title`, `description`, `type: object`, `additionalProperties: false`). Empty optional sections shall be omitted rather than emitted as empty mappings, and the serialization shall be deterministic so an unchanged bundle produces a byte-identical schema.

*Rationale:* Producers describe the attr block as data; the renderer owns the JSON-Schema envelope so every generated document is validated against exactly the attributes its producer emits, without anyone hand-writing a schema.

## SW-IMP-007: Python exchange helper
```attr
status: approved
package: python
priority: Medium
trace:
  - SYS-IMP-002
```
The `reqmd_io` Python package shall provide dataclasses (`Bundle`, `Document`, `Requirement`) that serialize to the `reqmd-import/v1` format, with the format's hyphenated keys and `omitempty`-equivalent behaviour (empty values omitted), and shall contain no Markdown or YAML logic. It shall be dependency-free and importable as `reqmd_io`, and the repository shall ship a runnable example importer that reads a DOORS/codeBeamer CSV export into a bundle.

*Rationale:* Importer authors should not hand-build dictionaries and rediscover the format's edge cases, but a second renderer in another language would drift from the Go one; restricting the helper to serialization gives the ergonomics without the divergence risk.

## SW-PAR-002: Content-tree parsing
```attr
status: approved
package: parser
priority: High
requires-trace-from: [software-requirements]
trace:
  - SYS-CON-001
```
The parser package shall emit a flat node stream per `.md` file — one `model.Node` per heading (kind requirement, container, or info) except the level-1 document title, which is captured as the file title and excluded from the stream — with prose, rationales, code blocks, and thematic breaks attached to the open node's body — and assemble it into the document content tree with `assembleTree`: parent = nearest preceding heading with a shallower level; `ParentID` set for requirements to the nearest ancestor requirement; info nodes with children promoted to `KindContainer`. Requirement headings shall be level 2 or higher; a level-1 heading with an `attr` block shall be rejected as a parse error. Multi-file documents shall merge per-file trees in `mdFiles` (sorted) order so output ordering is deterministic regardless of parse parallelism, take the document title from the first file's h1 heading, and the parallel chunked path shall include the file preamble in chunk 0 so leading prose and headings are never dropped.

*Rationale:* The chunked parse path previously split files at root-requirement boundaries, silently discarding the preamble. Assembly after concatenating chunk node streams (instead of per-chunk) preserves nesting across chunk boundaries (e.g. a container in chunk 0 holding a requirement in chunk 1) while keeping parse parallelism.

## SW-EXP-002: Item-aware export
```attr
status: approved
package: exporter
priority: Medium
requires-trace-from: [software-requirements]
trace:
  - SYS-CON-002
```
The exporter package shall render the document tree depth-first: requirement cards with child cards nested, containers as collapsible folder sections with stable `info-N` anchors, and info items as plain blocks. The embedded JSON index shall include every node with a `type` discriminator and tree references (parent/children anchors) so the sidebar TOC renders folders and scroll-spy covers item blocks; the client-side search/filter shall count requirements only. CSV export shall emit a `Type` column and rows for all nodes in document order.

*Rationale:* The index previously carried only requirement IDs and parent/child lists derived from `ParentID`. Containers need their own anchors and tree edges; keeping the index shape flat (id/type/parentId/children) lets the existing Alpine and vanilla-JS code consume the extended data with minimal client changes.
