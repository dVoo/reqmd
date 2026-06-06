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
The `attr` block shall support four built-in attributes — `trace`, `disposition`, `disposition-reason`, and `version` — that carry tool-level semantics in Pass 2 validation. Built-in attributes are injected automatically and must not be declared in `schema.yaml`. The parent-child relationship between requirements is structural (derived from heading level in the parser), not an attribute — no `parent` built-in is needed.

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
Each document directory shall contain exactly one `schema.yaml` that defines project-specific attributes using JSON Schema 2020-12 (YAML-serialized). An optional `x-reqmd` extension block provides directory-level metadata (level, upstream, id-prefix, external, mandatory-disposition, url, source).

The `external: true` flag marks a directory as a proxy for artefacts originating outside authored reqmd specs (standards, architecture tools, test frameworks). Proxy directories suppress the untraced WARNING globally. The `url` field provides a human-readable link; the `source` block (path + format) drives planned reqmd-scan generation.

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
The `reqmd` binary shall provide CLI subcommands: `check` (validate), `ls` (list), `stats` (attribute-value breakdown), and `export` (csv, html, graph). All commands accept a root directory for recursive document discovery. Exit codes distinguish success (0), validation errors (1), and parse errors (2).

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
The validation pipeline shall execute four sequential passes: Parse (goldmark AST), Pass 1 (JSON Schema validation with built-in injection), Graph build (pure Go in-memory adjacency with three sub-passes: node creation, trace-edge building, parent-child edge building), and Pass 2 (ten trace checks: broken reference, circular, untraced, no-downstream, disposition without reason, mandatory disposition, ID prefix mismatch, ID prefix collision [ERROR], duplicate ID [ERROR], ambiguous reference [ERROR]).

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


