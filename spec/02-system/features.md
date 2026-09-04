# reqmd System Features

## SYS-FMT-001: Markdown requirement format
```attr
status: approved
priority: Critical
trace:
  - STK-GOAL-001
  - ASP-SR-001
  - ASP-SR-003
  - ASP-SR-020
```
Each requirement shall be authored as a Markdown heading of any level (e.g., `#`, `##`, `###`) followed immediately by a fenced ` ```attr ``` ` YAML block, then free-form prose. The heading text is the requirement ID; an optional human-readable title may follow the ID after a colon and space (e.g., `## REQ-001: Login requirement`). The first heading+attr pair encountered in a document determines the requirement level (`reqLevel`). Headings at `reqLevel` with adjacent attr blocks are top-level requirements; headings at `reqLevel+1` with adjacent attr blocks are sub-requirements automatically linked to the preceding `reqLevel` parent. An optional `*Rationale:*` paragraph may follow the prose statement.

Headings without an adjacent attr block are treated as body text, enabling section dividers and annotations between requirements.

*Rationale:* Dynamic level discovery decouples requirement hierarchy from hardcoded heading depths, making documents easier to generate programmatically. The `isNextAttrBlock` lookahead provides a universal detection mechanism — a heading is a requirement iff its next AST sibling is an attr fence. Sub-requirements express V-model decomposition naturally via Markdown's native heading hierarchy.

## SYS-FMT-002: Built-in attribute system
```attr
status: approved
priority: Critical
trace:
  - STK-GOAL-002
  - ASP-SR-007
  - ASP-SR-013
```
The `attr` block shall support six built-in attributes — `trace`, `status`, `disposition`, `disposition-reason`, `version`, and `requires-trace-from` — that carry tool-level semantics in Pass 2 validation. Built-in attributes are injected automatically and must not be declared in `schema.yaml`. The parent-child relationship between requirements is structural (derived from heading level in the parser), not an attribute — no `parent` built-in is needed.

*Rationale:* Separation of concerns: project-specific attributes go in `schema.yaml`; tool-level attributes are always available without schema boilerplate. Blocking redefinition prevents ambiguity. Keeping the parent relationship structural avoids redundant data — the heading hierarchy is the source of truth.

## SYS-FMT-003: Schema.yaml document configuration
```attr
status: approved
priority: Critical
trace:
  - STK-GOAL-003
  - ASP-SR-002
  - ASP-SR-009
  - ASP-SR-015
```
Each document directory shall contain exactly one `schema.yaml` that defines project-specific attributes using JSON Schema 2020-12 (YAML-serialized). An optional `x-reqmd` extension block provides directory-level metadata (level, document-id, upstream, id-prefix, external, mandatory-disposition, url, source, requires-trace-from, ignore-status, additional-status-values, disjoint-check).

The `external: true` flag marks a directory as a proxy for artefacts originating outside authored reqmd specs (standards, architecture tools, test frameworks). Proxy directories suppress the untraced WARNING globally. The `url` field provides a human-readable link; the `source` block (path + format) drives planned reqmd-scan generation.

`x-reqmd.requires-trace-from` sets a document-wide default coverage expectation: an array of `document-id` or `level` tokens using the same vocabulary as the `requires-trace-from` attribute. Every requirement in the document that does not declare its own `requires-trace-from` attribute inherits this default; a requirement that declares the attribute overrides it entirely (no merging), and `requires-trace-from: []` opts that single requirement out of downstream coverage. A document without a default leaves its requirements on the generic boundary-inference fallback. The default is validated at schema compile time (an array of strings matching the attribute token pattern `^[a-z0-9_-]+$`, no duplicates); an empty default list means every requirement in the document expects no downstream coverage. Default tokens resolve exactly like per-requirement declarations (SYS-FMT-005).

Export output filenames follow the convention `<dirname>-requirements.<ext>`, written alongside each document directory or into `-o <dir>` if specified.

*Rationale:* JSON Schema 2020-12 is a widely-supported standard with validators in every language. The `x-reqmd` block keeps reqmd-specific configuration in the same file without breaking standard JSON Schema validators.

## SYS-CLI-001: CLI subcommands
```attr
status: approved
priority: Critical
trace:
  - STK-GOAL-001
  - STK-GOAL-004
  - ASP-SR-008
  - ASP-SR-016
  - ASP-SR-018
```
The `reqmd` binary shall provide CLI subcommands: `check` (validate), `ls` (list), `stats` (attribute-value breakdown), `export` (csv, html, graph), `init` (scaffold), `serve` (live-reloading HTML preview), `baseline diff` (compare git tags), and `repin` (bulk-update version pins). All commands accept a root directory for recursive document discovery. Exit codes distinguish success (0), validation errors (1), and parse errors (2).

*Rationale:* A single binary that walks the directory tree and handles all document directories in one pass minimizes CI complexity. Distinct exit codes enable pipeline branching (fail build on errors, warn on warnings).

## SYS-VAL-001: Multi-pass validation pipeline
```attr
status: approved
priority: Critical
trace:
  - STK-GOAL-003
  - STK-GOAL-005
  - ASP-SR-004
  - ASP-SR-005
  - ASP-SR-006
  - ASP-SR-008
  - ASP-SR-011
```
The validation pipeline shall execute four sequential passes: Parse (goldmark AST), Pass 1 (JSON Schema validation with built-in injection), Graph build (pure Go in-memory adjacency with three sub-passes: node creation, trace-edge building, parent-child edge building), and Pass 2 (thirteen trace checks: broken reference, circular, untraced, no-downstream, disposition without reason, mandatory disposition, ID prefix mismatch, ID prefix collision [ERROR], duplicate ID [ERROR], ambiguous reference [ERROR], version-pin [ERROR], missing-verdict [WARNING], failing-verdict [ERROR]).

Document discovery follows a recursive tree walk: every directory containing a `schema.yaml` is a document directory; all `*.md` files in that directory (non-recursive) are parsed against that directory's schema.

Properties are ordered in `ls` and CSV export as: required attributes first (in schema's `required` order), then optional attributes (in schema definition order), then built-in attributes.

The untraced check is suppressed for sub-requirements (`IsChild` flag set during graph construction), as their parent relationship provides the necessary trace context. Broken parent references discovered during Pass 3 produce ERROR-level results.

The disposition workflow is defined as:
- No `disposition` set → downstream traces expected (no-downstream is WARNING)
- `disposition: implemented` → downstream traces expected; no-downstream is WARNING
- `disposition: deferred` → no downstream traces expected; `disposition-reason` is REQUIRED
- `disposition: rejected` → no downstream traces expected; `disposition-reason` is REQUIRED

Defining a reserved built-in attribute in `schema.yaml` produces a compile-time error.

*Rationale:* Sequential passes ensure each stage has complete information. Parse must finish before validation; graph must be built before trace checks; all results are collected before reporting. Three sub-passes within graph construction guarantee deterministic node initialization before edges are wired.

## SYS-TRC-001: Upstream-only trace model
```attr
status: approved
priority: Critical
trace:
  - STK-GOAL-005
  - ASP-SR-004
  - ASP-SR-007
  - ASP-SR-009
  - ASP-SR-011
  - ASP-SR-013
  - ASP-SR-017
  - ASP-SR-020
```
The trace model shall be upstream-only: each requirement declares its parent(s) via the `trace` attribute. Downstream links are derived automatically when the graph is built across all documents. The tool shall also support trace pinning via a `version` integer attribute and an optional `~N` suffix on trace values (e.g., `SYS-042~3` means "fulfills SYS-042 version 3"). Additionally, sub-requirements receive an automatic parent-child edge from heading hierarchy during graph Pass 3 — no manual `trace` attribute is needed for the parent link. V-model boundaries are inferred from document topology: directories with no upstream traces suppress untraced warnings; directories not referenced as upstream suppress no-downstream warnings.

When multiple documents share the same ID prefix (e.g., two teams both use `STK-`), path-qualified trace references disambiguate: `stakeholder-a/STK-001` resolves to STK-001 in the `stakeholder-a/` directory. Unqualified references (`STK-001`) work when the ID is globally unique. An unqualified reference that matches IDs in multiple documents produces an ERROR suggesting the qualified form.

*Rationale:* Upstream-only declarations avoid redundancy — if every child names its parent, the full chain is determined without any requirement also listing its children. Version pinning lets PR reviewers detect upstream changes since the trace was last validated. Auto-edges for sub-requirements ensure decomposition hierarchy is always present without manual attribute maintenance. V-model boundaries are inferred from directory topology (upstream.sources), eliminating per-requirement boilerplate.

## SYS-FMT-004: YAML frontmatter support
```attr
status: approved
priority: High
trace:
  - STK-GOAL-002
```
Documents SHOULD support YAML frontmatter (`---`-delimited) parsed via goldmark-meta. If present, the `description` key provides document-level prose rendered in the HTML export below the document header. Frontmatter parsing must not interfere with requirement `attr`-block parsing within the same file.

*Rationale:* YAML frontmatter is a standard convention across Markdown toolchains (Jekyll, Hugo, GitHub Pages). Supporting it lets authors add document-level context — like an abstract or navigation hint — without requiring a separate file.

## SYS-HTM-001: Standalone HTML export
```attr
status: approved
priority: High
trace:
  - STK-GOAL-004
  - ASP-SR-010
  - ASP-SR-016
```
The HTML exporter SHALL produce a standalone `.html` page per document directory suitable for human traceability review, with all requirement attributes, rendered prose, trace navigation, hierarchical document structure, and theme customization — without requiring network access after generation.

*Rationale:* The HTML export is the primary human-reviewable output for non-technical stakeholders. A self-contained page eliminates server dependencies for distribution.

## SYS-BAS-001: Baseline comparison via git tags
```attr
status: approved
priority: High
trace:
  - STK-GOAL-002
```
The tool shall provide a `reqmd baseline diff <tag1> <tag2>` subcommand that compares requirement specifications between two git tags without requiring a working-tree checkout. The command shall extract the repository at each tag using `git archive | tar`, parse both versions with the standard pipeline, and produce a semantic diff showing added, removed, and modified requirements with attribute-level detail, plus schema changes (new/removed properties, changed required fields). Output shall be colored text by default with a `--json` flag for structured output.

*Rationale:* Requirement-level diffs enable change-impact analysis and release notes generation. Using git tags as baseline anchors means the diff inherits the team's existing version-control workflow — no separate baseline store to maintain. `git archive` avoids the need to check out tags in separate worktrees, keeping the operation fast and side-effect-free.

## SYS-REPIN-001: Bulk version-pin update via `reqmd repin`
```attr
status: approved
priority: High
requires-trace-from: [system-requirements]
trace:
  - STK-GOAL-002
```
The tool shall provide a `reqmd repin <root>` subcommand that scans the spec tree for trace refs whose `~N` version pin is below the upstream's current `version` and either lists the proposed changes (dry-run, default) or rewrites the affected ` ```attr ` blocks in place (`--yes`). Predated findings (pin ahead of upstream) shall be surfaced in the change list but never auto-fixed, because they are data-integrity errors. The subcommand shall accept a `--promote-unpinned` flag that additionally proposes pins for trace refs without any `~N` against a versioned upstream, and a `--json` flag that emits a stable JSON contract (`deltas[]` with `kind` ∈ {`outdated`, `unpinned`, `predated`}, `req_id`, `file`, `target_id`, `source_ref`, `old_pin`, `new_pin`, `new_version`; plus `outdated`/`unpinned`/`predated` counts and a `by_file` map). Without `--yes`, the command shall run by default in dry-run mode; in an interactive TTY it shall prompt for confirmation before applying. The rewrite shall be scoped to ` ```attr ` blocks only, leaving surrounding Markdown prose verbatim.

*Rationale:* The version-pin check (SYS-VAL-001) reports every outdated pin individually; without an apply path users have to hand-edit each `trace:` line. `repin` is the bulk-update counterpart to the check, with a mandatory dry-run/confirm step so the change set is reviewable before files are written. Whole-ref matching (not substring) prevents an unpinned promote from corrupting an already-pinned sibling, e.g. rewriting `UP-001` inside `UP-001~3` to `UP-001~3~3`. Predated findings are excluded from the apply set because they indicate that the upstream itself needs attention (a newer version was rolled back, or the pin is wrong), not that the pin should be lowered.


## SYS-VAL-002: Verification result ingestion and outcome-gated checks
```attr
status: approved
priority: High
trace:
  - STK-GOAL-005
  - ASP-SR-006
```
The `check` and `serve` commands shall accept a repeatable `--results <path>` flag that loads ephemeral verification results from CTRF JSON reports (automated tests) and manual-results markdown directories (review, inspection, analysis). Each path is auto-detected: directories are walked, `.ctrf.json` and CTRF-shaped `.json` files are parsed as CTRF, subdirectories with a `schema.yaml` are loaded as manual results via the standard document pipeline, and non-CTRF files are skipped silently. Single files are parsed as CTRF.

Results are loaded per invocation and are not persisted in the spec repo. CTRF test entries map to verification measures via the `x-reqmd.id` extra field (with optional `~N` version pin). CTRF `status` maps to a reqmd `outcome` (`passed→pass`, `failed→fail`, `skipped→skipped`, `pending|other|flaky→inconclusive`). Manual results carry `outcome`, `verifier`, `evidence`, `verified-at`, and `trace: [MEASURE-ID]`. Across all inputs, the latest verdict per measure wins by timestamp.

Each merged result is synthesized as a pseudo-requirement (`RESULT:<id>`, `trace: [MEASURE-ID~N]`, `outcome: <verdict>`) appended to the document slice before graph build. Two outcome-gated checks run when result nodes are present: `missing-verdict` (approved measure with no result, WARNING, suppressible) and `failing-verdict` (latest result is fail, ERROR, suppressible). The existing `version-pin` check applies to result→measure traces, reusing stale-verdict detection with no new logic. Draft measures are skipped by `missing-verdict`.

CTRF entries may additionally carry `x-reqmd.case`, `x-reqmd.verifies`, and `x-reqmd.description` so a run binds through an explicit test case: `id` alone attaches the result directly to a measure; `case` + `verifies` (or `verifies` alone, keyed by the normalized `(suite, name)`) synthesizes a test case whose `trace` edges target the verified requirements and whose body is the `description`. Results merge latest-wins per (measure, case) by CTRF `stop`. Outcome-gated checks roll up a measure's evidence over its own results and its approved downstream test cases (authored or synthesized) with strict aggregation (any fail → fail, else inconclusive → inconclusive, else skipped → skipped, else pass); draft downstream cases are reported as ignored, and deferred/rejected measures are exempt. `check` accepts `--ignore-unbound-results` to suppress `unbound-result` warnings for entries that declare `x-reqmd` but bind to nothing; instrumented entries binding nowhere surface as broken references.

*Rationale:* reqmd traces the left side of the V-model (stakeholder → system → software → test specs) but could not represent the right side: verification *results*. Every ASPICE BP that says "record the verification results including pass/fail status and evaluate" was unrepresentable. The `status: draft|approved` axis is a lifecycle gate, not a verdict — an approved, implemented test can still fail. Ephemeral results keep CI run-to-run churn out of git and `baseline diff`, while the outcome-gated checks close the V-model right side by enforcing that every approved measure has a passing latest verdict.

## SYS-FIL-001: Generic attribute filtering
```attr
status: approved
priority: Medium
trace:
  - STK-GOAL-003
  - STK-GOAL-005
  - ASP-SR-005
  - ASP-SR-008
  - ASP-SR-018
```
The `check`, `ls`, `stats`, `export csv`, `export html`, and `serve` commands shall accept a `--filter "<expr>"` flag that scopes the operation to requirements matching an expression evaluated against the requirement's attributes plus the built-in `id` and `title` variables. Expressions shall be compiled once at startup into bytecode, and compile-time validation shall reject expressions referencing attributes not declared in any `schema.yaml` (global typo check). Per-document attribute absence shall evaluate to nil rather than erroring.

Coverage checking shall be filter-aware: a filtered-out requirement cannot satisfy the `requires-trace-from` expectation of a filtered-in requirement, so filtered-out requirements do not cause false coverage failures. The graph is always built from the full tree so trace refs to filtered-out requirements still resolve; only per-requirement checks and coverage are scoped. `baseline diff` shall additionally accept `--filter-a` and `--filter-b` to compare two filtered views of a single commit (e.g. "what does Premium add over Base").

*Rationale:* Large specs make full-tree review and incremental CI expensive. Attribute filtering lets teams scope `check`, exports, and diffs to the slice of the tree that a change touches, without losing the ability to resolve traces globally. Rejecting undeclared attribute names at compile time catches typos before they silently match nothing, and filter-aware coverage (RFC §3.2) keeps filtered-out requirements from causing false failures.

## SYS-CHK-003: Disjoint-attribute consistency check
```attr
status: approved
priority: Medium
trace:
  - STK-GOAL-003
  - ASP-SR-005
  - ASP-SR-011
```
The `check` command shall accept a repeatable `--disjoint-check <attr>` flag (and the equivalent `x-reqmd.disjoint-check` key in `schema.yaml`, string or array) that, for every trace link, verifies that the source and target requirements have at least one overlapping value for the named array-typed attribute. A zero intersection shall produce an ERROR (`disjoint-attribute`); an empty or absent value is exempt ("applies to all"). The check shall be suppressible per requirement via `reqmd-suppress: [disjoint-attribute]`.

*Rationale:* Attribute-consistency gates capture cross-cutting invariants that pure trace-structure checks miss — e.g. "a trace link may only connect two requirements that share a product variant or a safety level". They turn a manual review convention into an enforceable CI check with a single flag, without adding a new built-in attribute.

## SYS-STS-001: Requirement status lifecycle
```attr
status: approved
priority: Medium
trace:
  - STK-GOAL-003
  - STK-GOAL-005
  - ASP-SR-006
  - ASP-SR-008
```
`status` shall be a built-in attribute with enum `[draft, approved]`; a missing `status` shall default to `approved` (silently). Only `approved` requirements shall count as upstream coverage providers — a `draft` requirement referenced by a `trace` shall NOT satisfy a `requires-trace-from` expectation. Schemas may extend the enum via `x-reqmd.additional-status-values` (lowercase-only values, no built-in collision, no duplicates) or opt out of the lifecycle per document via `x-reqmd.ignore-status: true` (all requirements become coverage providers; the status filter is hidden in HTML export). When a missing-coverage message is produced and at least one inbound was filtered by the status gate, the message shall report how many draft downstreams were ignored.

*Rationale:* A lifecycle axis distinct from `disposition` lets teams keep in-progress requirements in the spec without letting them (falsely) satisfy coverage. Draft requirements remain traceable and validated, but cannot close a coverage expectation until promoted to `approved`.

## SYS-SUP-001: Check suppression
```attr
status: approved
priority: Medium
trace:
  - STK-GOAL-003
  - ASP-SR-008
```
Any requirement may declare a `reqmd-suppress: [<check-code>]` attribute listing graph checks to suppress for that requirement (e.g. `version-pin`, `missing-verdict`, `failing-verdict`, `disjoint-attribute`). Suppressed checks shall emit no result for the declaring requirement, in all commands that run graph checks.

*Rationale:* Not every check applies to every requirement; a deliberate, documented exception should not fail CI. Suppression is per-requirement and opt-in, so the default remains strict — an exception is visible in the source and auditable in review.

## SYS-FMT-005: Level-typed trace coverage
```attr
status: approved
priority: High
trace:
  - STK-GOAL-005
  - ASP-SR-004
  - ASP-SR-020
```
A `requires-trace-from` token shall name either a `document-id` or an `x-reqmd.level`. When a token names a level, reqmd shall resolve it to every document directory that declares that level and check that at least one approved requirement from one of those directories traces back — covering all documents at a V-model layer without listing each `document-id`. A token that matches neither a `document-id` nor a declared level shall produce a WARNING. Coverage tokens inherited from the `x-reqmd.requires-trace-from` document default (SYS-FMT-003) shall resolve identically: a requirement inheriting the default is subject to the same per-token checks, status gate, and filter-awareness as one declaring the tokens in its own `requires-trace-from` attribute.

*Rationale:* Multi-document layers (e.g. several `system-requirements` directories owned by different teams) otherwise force every coverage expectation to enumerate every document-id. Level-typed coverage makes the V-model layer itself a first-class coverage target and stays robust as documents are added or removed within a layer.

## SYS-IMP-001: Source-code requirement extraction
```attr
status: approved
priority: High
trace:
  - STK-GOAL-002
  - STK-GOAL-005
  - ASP-SR-001
  - ASP-SR-002
  - ASP-SR-022
```
The toolchain shall include `reqmd-import`, a companion extraction tool that walks a source tree, dispatches each file to the tree-sitter language plugin claiming its extension, extracts top-level symbols (functions, methods, types, constants, variables) from C, C++, Go, Python, and Rust source, binds the adjacent doc comment as the requirement body, extracts explicit `reqmd:trace` markers (and, opt-in via `--heuristic-traces`, bare requirement IDs from prose), and writes per-package `.md` requirement files into a target spec directory.

Generated requirements shall be first-class spec documents: `status: approved`, `x-reqmd.imported: true`, provenance attributes (`x-reqmd.source-file`, `x-reqmd.source-line`, `x-reqmd.symbol-kind`), and IDs of the form `<id-prefix><package>-<symbol>-<hash>` (hash over package+symbol so IDs survive line shifts). The generated files shall validate, export, and diff like any other spec document, closing the V-model gap from spec to code.

*Rationale:* Software requirements trace down to tests (ASP-SR-006) but not to the implementation that satisfies them. Automatic extraction turns source symbols into requirements in the same ID namespace, so a broken spec→code trace is caught by the same `reqmd check` that catches broken spec refs — and re-running the extractor keeps the code-side of the V-model current without hand-maintained mapping tables.


## SYS-CON-001: Content tree — nothing dropped
```attr
status: approved
priority: High
trace:
  - STK-GOAL-003
  - STK-GOAL-005
  - ASP-SR-008
```
The parser shall capture the full document content as a content tree in which every heading is a node except the level-1 document title: a heading followed by an `attr` block is a requirement; a heading without an `attr` block that contains nested nodes is a container; a heading without an `attr` block and no children, or prose preceding the first heading, is an info item. A level-1 heading shall be treated as the document title, not content: it is captured as the document title, is never a content node, and must not carry an `attr` block — requirement headings shall be level 2 or higher, and a level-1 heading with an `attr` block is a parse error. Prose, code blocks, and thematic breaks shall attach to the nearest open node's body. No content shall be dropped: leading prose, requirement-less files, and headings without attributes are all preserved as nodes.

*Rationale:* Requirements documents carry structure and narrative beyond requirement records. Dropping non-requirement content (as early prototypes did) made exported pages silently incomplete. The tree keeps document order and hierarchy faithful to the source so `ls`, `stats`, CSV, and HTML can render the document as written, not just its requirement subset.

## SYS-CON-002: Items in output commands
```attr
status: approved
priority: Medium
trace:
  - SYS-CON-001
```
Containers and info items shall be visible in `ls` (a leading `Type` column — `req`/`container`/`info` — with item rows showing heading text), `stats` (per-document item counts plus a type breakdown), CSV export (a `Type` column, rows in document order), and HTML export (collapsible folder sections for containers, plain blocks for info items, and sidebar TOC entries). Items shall have no attributes and shall not participate in validation, trace checks, coverage, or `--filter` scoping — filtering prunes requirements only. `check` shall remain requirement-scoped.

*Rationale:* Items are presentation, not specification: they carry no schema-validated data, so no graph check can attach to them. Keeping them out of validation prevents noise while still rendering the authored document faithfully.
