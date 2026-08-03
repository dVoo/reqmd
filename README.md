# ReqMD

**ReqMD** (pronounced "req-em-dee") is a lightweight CLI tool for writing, validating, and exporting
requirement specifications in plain Markdown. Designed for teams that want Git-native, text-first
requirements management without a heavy toolchain.

```
$ reqmd check example/
=
Schema : ivi-requirements — IVI Requirements
File   : example  (3 requirements)
=
  ⚠  IVI-FUN-001  attributes valid (see trace checks below)
  ⚠  IVI-FUN-001  broken reference: target "SYS-001" not found
  ⚠  IVI-FUN-001  broken reference: target "SAFE-003" not found
  ⚠  IVI-FUN-001  no upstream reference
  ⚠  IVI-FUN-002  attributes valid (see trace checks below)
  ⚠  IVI-FUN-002  broken reference: target "SYS-007" not found
  ❌  IVI-FUN-003  validating ivi-requirements: required: missing properties: ["asil"]
  ⚠  IVI-FUN-003  untraced: no downstream reference
  ⚠  IVI-FUN-003  no upstream reference

Summary: 3 total, 2 valid, 1 invalid, 0 parse errors, 6 warnings
```

---

## Installation

```sh
git clone <your-repo>
cd mdreq
go build -o reqmd ./cmd/reqmd
go build -o reqmd-import ./reqmd-import/cmd/reqmd-import   # extraction tool
```

The repo is a Go workspace (`go.work`) linking both modules, so `go build ./...`
and `go test ./...` from the root cover reqmd and reqmd-import together:

```sh
go test ./...        # tests for both modules
```

Each module keeps its own `go.mod` and dependencies; install either
independently with `go install ./cmd/reqmd` or
`go install ./reqmd-import/cmd/reqmd-import`.

Requires Go 1.25+. The graph export subcommand requires Cgo support;
build with `go build -tags ladybug -o reqmd ./cmd/reqmd` to include it.

### Pre-built binaries

Each tagged release publishes pre-built `reqmd` and `reqmd-import` binaries
for Linux (amd64), Windows (amd64), and macOS (arm64) plus a `SHA256SUMS.txt`
file. Download them from the [Releases page](../../releases) and verify with:

```sh
sha256sum -c SHA256SUMS.txt --ignore-missing
```

> Note: the released `reqmd` binaries are built with `-tags ladybug`, so
> `reqmd export graph` is available. This links against the native `liblbug`
> shared library; at runtime the library must be reachable via the system's
> dynamic linker search path (e.g. `LD_LIBRARY_PATH` on Linux,
> `DYLD_LIBRARY_PATH` on macOS, or `PATH` on Windows). Download the matching
> `liblbug-*` archive from [LadybugDB/ladybug releases](https://github.com/LadybugDB/ladybug/releases)
> and place the shared library on the appropriate path. On Linux, OpenSSL 3
> (`libssl`/`libcrypto`) must also be installed system-wide.
>
> If you do not need `export graph`, build from source without the `ladybug`
> tag for a pure-Go, dependency-free binary (see Installation above).

---

## Quick Start

The repo ships with an example you can try immediately:

```sh
go run ./cmd/reqmd check example/
go run ./cmd/reqmd ls example/
go run ./cmd/reqmd stats example/
go run ./cmd/reqmd check --json example/
go run ./cmd/reqmd export csv example/ -o /tmp/out/
go run ./cmd/reqmd export html example/ -o /tmp/out/
```

Scaffold a new requirements project — choose a built-in preset or use your own template directory:

```sh
reqmd init my-requirements/                              # generic preset (default)
reqmd init my-requirements/ --preset aspice               # automotive SPICE template
reqmd init my-results/ --preset results --id-prefix VR    # manual verification results
reqmd init my-project/ --preset ./my-preset/              # custom preset directory
reqmd check my-requirements/
```

Watch for changes and serve a live-reloading HTML preview:

```sh
go run ./cmd/reqmd serve spec/          # opens browser
go run ./cmd/reqmd serve spec/ --headless   # terminal-only
```

---

## Core Concepts

### 1. Requirement files

Each requirement is a level-2 heading followed by a fenced `attr` block with
YAML attributes, then free-form prose. The heading text **is** the requirement ID.
An optional title can follow the ID after a colon and space:

````markdown
## IVI-FUN-001: Fast startup
```attr
status: approved
asil: QM
maturity: Production
verify: Test
owner: TierOneSupplierA
trace: [SYS-001, SAFE-003]
version: 1
```
The system shall display the home screen within 5 seconds after ignition on,
provided the head unit is operational.

*Rationale:* Fast startup improves perceived quality.
````

Demonstrates: `trace` (upstream refs), `version` (version pinning), and the `ID: Title` syntax (optional human-readable title after the ID).

```markdown
## IVI-FUN-002
```attr
status: draft
asil: B
maturity: Prototype
verify: Test
owner: TierOneSupplierA
trace: [SYS-007]
disposition: deferred
disposition-reason: "Moved to Phase 2 — requires next PCB revision"
```
The system shall disable manual text entry while vehicle speed is greater
than 0 km/h, except for approved passenger-only functions.
```

Demonstrates: `disposition` + `disposition-reason` (defers traces, suppresses warnings).

```markdown
## IVI-FUN-003
```attr
status: draft
maturity: Concept
verify: Test
```
The system shall resume the last active audio source after an ignition cycle.
```

No built-in attrs — triggers untraced, no upstream reference, and any missing
required fields from the schema.

**Rules:**
- Heading must start with a unique requirement ID (e.g. `IVI-FUN-001`, `SYS-001`)
- An optional title can follow the ID after a colon and space: `## IVI-FUN-001: Fast startup`
- The `attr` block must be valid YAML — keys must match the directory's schema
- Requirement IDs use uppercase letters, digits, and hyphens: `^[A-Z][A-Z0-9]*(-[A-Z0-9]+)*$` with a trailing number segment (e.g. `STK-GOAL-001`, `SYS-FMT-002`). Optional version pinning suffix `~N` is allowed for trace references (e.g. `SYS-001~3`). Colons and spaces are not allowed in IDs — they delimit the optional title.
- If `x-reqmd.id-prefix` is set, every ID in that directory must start with the declared prefix

### 2. Schema files (`schema.yaml`)

Every directory with requirements needs a `schema.yaml` — standard JSON Schema 2020-12
in YAML, defining which attributes are allowed, required, and their types.

```yaml
x-reqmd:
  level: software-requirements    # V-model layer label
  id-prefix: IVI-FUN-             # enforce ID prefix

$schema: "https://json-schema.org/draft/2020-12/schema"
$id: "ivi-requirements"
title: "IVI Requirements"
type: object
required: [asil, maturity, status, verify]
properties:
  status:
    type: string
    enum: [draft, approved]
  asil:
    type: string
    enum: [QM, A, B, C, D]
  maturity:
    type: string
    enum: [Concept, Prototype, Production, Serial]
  verify:
    type: string
    enum: [Test, Analysis, Inspection, Review]
  owner:
    type: string
  priority:
    type: string
    enum: [Low, Medium, High, Critical]
  safety_relevant:
    type: boolean
additionalProperties: false
```

### 3. Built-in attributes

These attributes are reserved by reqmd and injected automatically — you must
**not** list them in `schema.yaml`:

| Attribute | Type | Description |
|-----------|------|-------------|
| `trace` | `ref[]` | Cross-document upstream references, e.g. `[SYS-001, SAFE-003]` |
| `disposition` | `enum` | How intent is addressed: `implemented`, `deferred`, or `rejected` |
| `disposition-reason` | `string` | Required when disposition ≠ `implemented` |
| `requires-trace-from` | `string[]` | Coverage expectations — which downstream levels or document-ids are expected to trace to this requirement |
| `version` | `int` | Version number for trace pinning (e.g. `SYS-001~3`) |
| `status` | `enum` | Approval lifecycle. Default: `approved`. Only `approved` satisfies traceability coverage. |

All built-in attributes are **optional**. Redefining them in `schema.yaml` is a compile error.

### 4. Status lifecycle

Every requirement has an approval status. The built-in enum is `[draft, approved]`.
Omitting `status` is equivalent to `status: approved` (silent default).

**Coverage rule:** only `approved` requirements count as upstream coverage
providers. A `draft` requirement can be referenced by a `trace`, but it does
**not** satisfy a `requires-trace-from:` expectation. This is the only behavioral difference
between `draft` and `approved` — drafts remain visible, exported, and clickable,
they simply do not complete a coverage chain.

```yaml
## REQ-001: System boot
```attr
status: approved
trace: [STK-001]
```

```yaml
## REQ-002: Power management (in progress)
```attr
status: draft
trace: [STK-002]
requires-trace-from: [system]
```
The `requires-trace-from: [system]` declaration will emit a warning while REQ-002 is still
`draft`, because no system-level requirement traces here yet. Bump it to
`status: approved` once downstream work is complete.

The coverage-gate message keeps authors informed:

```
no upstream trace: no approved requirement from "system" traces to this item (3 draft downstreams ignored)
```

The `(N draft downstreams ignored)` parenthetical appears only when at least
one inbound candidate was filtered by the status gate. When no inbound exists
at all, the message is silent on draft counts.

### 4a. Extending the status enum

When the built-in enum is too narrow, declare extensions in `x-reqmd`:

```yaml
x-reqmd:
  additional-status-values: [review, in-progress]
```

Rules (violations are **config errors** at compile time):
- Values must be lowercase, matching `^[a-z][a-z0-9_-]*$`.
- `draft` and `approved` cannot be redeclared.
- Duplicates are rejected.
- An empty array (`[]`) is a no-op.

Extensions appear in the status filter and header breakdown alongside the
built-ins. Their dot color is hash-derived (DJB2a, 4 collision-free hue
regions) so authors don't need to configure colors.

### 4b. Opting out of the lifecycle

Documents that don't follow the lifecycle (legacy imports, generated specs,
one-off documents) can opt out:

```yaml
x-reqmd:
  ignore-status: true
```

In an `ignore-status` document:
- All requirements count as coverage providers, regardless of their `status`.
- The status filter is hidden in the HTML export.
- The status breakdown row is not rendered.

### 5. Directory metadata (`x-reqmd`)

The `x-reqmd` key in `schema.yaml` carries directory-level metadata. It's optional —
standard JSON Schema validators ignore it because of the `x-` prefix.

```yaml
x-reqmd:
  level: software-requirements                 # V-model layer label
  upstream:                          # parent layer config
    level: system-requirements
    sources:
      - ../sys/
  mandatory-disposition: false                 # promote missing disposition to ERROR
  external: false                              # proxy for non-reqmd artefacts
  url: "https://example.com/model-export"      # human link (HTML export)
  source:
    path: ".reqmd/arch/sys/*.md"               # reqmd-scan input glob
    format: archi                               # input format
  id-prefix: IVI-FUN-                          # enforce ID prefix
```

| Field | What it does |
|-------|-------------|
| `level` | Labels the V-model layer for reporting and diagram generation |
| `document-id` | Stable identifier for the document. Used to disambiguate `trace: [doc-id/ID]` references and as the `requires-trace-from:` coverage target |
| `upstream.level` | Names the expected parent layer (informational) |
| `upstream.sources` | Relative paths to upstream document directories |
| `mandatory-disposition` | When `true`, missing `disposition` becomes an **ERROR** |
| `external` | When `true`, marks as proxy directory; untraced warnings suppressed |
| `url` | Human-readable link for HTML export (meaningful only when `external: true`) |
| `source.path` | Path/glob/URL for the source artefact (future `reqmd-scan`) |
| `source.format` | Source format: `archi`, `doxygen`, `gtest`, `pytest`, `junit`, `reqif` |
| `id-prefix` | Enforces that all requirement IDs start with this prefix; collision is an ERROR |
| `additional-status-values` | Lowercase extensions to the built-in `status` enum (e.g. `[review]`). See [Status lifecycle](#4-status-lifecycle). |
| `ignore-status` | When `true`, the document opts out of the status lifecycle (see [Opting out](#4b-opting-out-of-the-lifecycle)) |

**Simple rule:**
- No `external` field or `external: false` → normal authored requirement directory
- `external: true` → proxy/generated directory for external or scanned artefacts

### 5. Directory layout

Any folder with a `schema.yaml` is a document directory. ReqMD validates all `.md`
files inside it (non-recursive), then recurses into subdirectories.

```
requirements/
├── stakeholder/
│   ├── schema.yaml                # level: stakeholder-needs
│   └── needs.md
├── sys/requirements/
│   ├── schema.yaml                # level: system-requirements
│   └── system-functions.md
├── swe/requirements/
│   ├── schema.yaml                # level: software-requirements
│   ├── startup.md
│   └── media.md
├── tests/swe/unit/
│   ├── schema.yaml                # level: test-spec-unit
│   └── ecum-tests.md
├── external/
│   └── autosar-ecum/
│       └── schema.yaml            # external: true — committed config only
└── .github/workflows/
    └── validate.yml               # CI: runs reqmd check on every PR
```

### 6. Document frontmatter (YAML)

Each `.md` file MAY include YAML frontmatter delimited by `---`, parsed via
[goldmark-meta](https://github.com/yuin/goldmark-meta). The `description` key
provides document-level prose rendered in the HTML export below the document
header:

```yaml
---
description: "This document defines the **stakeholder goals** that drive all downstream requirements."
---
```

Frontmatter keys from multiple `.md` files in the same document directory are
merged with first-file-wins semantics. The `description` field is processed
through goldmark, supporting inline Markdown (bold, code, links, etc.).

### 7. Goldmark extensions

ReqMD uses [goldmark](https://github.com/yuin/goldmark) for Markdown rendering
in both parsing and HTML export, with these extensions enabled:

| Extension | Package | What it does |
|-----------|---------|-------------|
| **GFM** | `github.com/yuin/goldmark/extension` | Tables, strikethrough, autolinks, task lists |
| **AutoHeadingID** | `github.com/yuin/goldmark/parser` | Auto-generates `id` attributes for headings |
| **Mermaid** | `go.abhg.dev/goldmark/mermaid` | Diagram rendering from `` ```mermaid `` blocks |
| **Fenced divs** | `github.com/stefanfritsch/goldmark-fences` | Pandoc-style `::: {.class}` containers |
| **Highlighting** | `github.com/yuin/goldmark-highlighting/v2` | Syntax-highlighted code blocks via chroma |
| **Emoji** | `github.com/yuin/goldmark-emoji` | GitHub-style `:joy:` → 😊 emoji |
| **KaTeX** | `github.com/FurqanSoftware/goldmark-katex` | Math rendering: inline `$x^2$` and display `$$...$$` |

---

## Commands

| Command | What it does |
|---------|--------------|
| `reqmd check <root>` | Validate all `.md` against their `schema.yaml` recursively |
| `reqmd check <file> -s <schema>` | Validate a single file against an explicit schema |
| `reqmd check --json <root>` | JSON validation report |
| `reqmd check --relaxed-versions <root>` | Demote outdated version-pin findings from ERROR to WARNING (predated stays ERROR) |
| `reqmd check --results <path> <root>` | Load ephemeral verification results (CTRF `.ctrf.json` or manual-results dirs) and run outcome-gated checks (missing-verdict, failing-verdict). `--results` is repeatable; auto-detects CTRF vs manual by extension + shape. |
| `reqmd check --filter "<expr>" <root>` | Scope validation to requirements matching an [expr-lang](https://expr-lang.org) expression (e.g. `"Premium" in variant`). Coverage checking becomes filter-aware: filtered-out requirements cannot cause false coverage failures. `--json` adds a `"filter"` field to the summary. |
| `reqmd check --disjoint-check <attr> <root>` | Check that trace-linked requirements have overlapping values for the named array-typed attribute (e.g. `variant`). Zero intersection → ERROR; empty/absent = "applies to all" (exempt). Repeatable. Also settable via `x-reqmd.disjoint-check` in `schema.yaml`. |
| `reqmd init <dir>` | Scaffold a new requirements directory with schema.yaml and example file. Presets: `generic` (default), `aspice`, `results` (manual verification results), or a custom preset directory path. Flags: `--preset`, `--id-prefix`, `--id`, `--title`, `--level`, `--force` |
| `reqmd ls <root>` | Table of all requirements with attribute values. `--filter "<expr>"` scopes to matching requirements. |
| `reqmd ls --json <root>` | JSON list. `--filter` supported. |
| `reqmd stats <root>` | Attribute-value breakdown per document. `--filter "<expr>"` scopes to matching requirements. |
| `reqmd stats --json <root>` | JSON stats. `--filter` supported. |
| `reqmd export csv <root> -o <dir>` | CSV export with Body and Rationale columns. `--results <path>` (repeatable) adds Verdict and Verdict Source columns. `--filter "<expr>"` exports only matching requirements. |
| `reqmd export html <root> -o <dir>` | Standalone HTML: card layout, goldmark-rendered body, trace columns, doc chain tab strip, search/filter, theme toggle. `--results <path>` (repeatable) renders color-coded verdict badges (pass/fail/skipped/inconclusive) on measure cards. `--filter "<expr>"` exports only matching requirements. |
| `reqmd export graph <root> -o <dir>` | Exports trace graph to ladybugdb for Cypher querying (requires `-tags ladybug` build). `--results <path>` (repeatable) includes `RESULT:` nodes with `outcome` and `source` properties, enabling graph traversal from requirements through measures to verification results. |
| `reqmd serve <root>` | Watch for changes and serve live-reloading HTML preview with SSE auto-reload (flags: `--addr`, `--headless`, `--no-open`, `--debounce`, `--results`, `--filter`) |
| `reqmd baseline diff <tag1> <tag2>` | Compare requirements and submodule pins between two git tags (flags: `--json`, `--filter "<expr>"` to scope both snapshots) |
| `reqmd baseline diff --filter-a "<expr>" --filter-b "<expr>" [<ref>]` | Compare two filtered views of the same commit (default `HEAD`). Useful for "what does Premium add over Base" reports from a single commit. |
| `reqmd repin <root> [-y/--yes] [--json] [--promote-unpinned]` | Propose or apply `~N` version-pin updates so trace refs match the upstream's current `version`. Dry-run by default; `--yes` applies. `--promote-unpinned` also pins refs that have no `~N` against a versioned upstream. Predated findings (pin > upstream) are surfaced but never auto-fixed. |

Aliases: `check` = `validate` or `v`; `ls` = `list` or `l`.

### Example: `check`

```
=
Schema : ivi-requirements — IVI Requirements
File   : example  (3 requirements)
=
  ⚠  IVI-FUN-001  attributes valid (see trace checks below)
  ⚠  IVI-FUN-001  broken reference: target "SYS-001" not found
  ⚠  IVI-FUN-001  broken reference: target "SAFE-003" not found
  ⚠  IVI-FUN-001  no upstream reference
  ⚠  IVI-FUN-002  attributes valid (see trace checks below)
  ⚠  IVI-FUN-002  broken reference: target "SYS-007" not found
  ❌  IVI-FUN-003  validating ivi-requirements: required: missing properties: ["asil"]
  ⚠  IVI-FUN-003  untraced: no downstream reference
  ⚠  IVI-FUN-003  no upstream reference

Summary: 3 total, 2 valid, 1 invalid, 0 parse errors, 6 warnings
```

- `⚠` — trace validation warning (broken reference, untraced, no upstream reference, disposition without reason)
- `❌` — schema validation error (missing required, wrong type, invalid enum)
- IVI-FUN-001 traces to external targets (broken references; untraced suppressed by `requires-trace-from: []`)
- IVI-FUN-002 uses `disposition: deferred` (requires `disposition-reason`)

### Example: `ls`

```
=== example ===
ID                       | asil         | maturity     | status       | verify       | owner        | priority     | safety_relevant | trace        | disposition  | disposition-reason | version  |
----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
IVI-FUN-001              | QM           | Production   | approved     | Test         | TierOneSupplierA |              |              | ["SYS-001","SAFE-003"] |              |              | 1            |
IVI-FUN-002              | B            | Prototype    | draft        | Test         | TierOneSupplierA |              |              | ["SYS-007"]  | deferred     | Moved to Phase 2 — serial peripheral interface requires next PCB revision |              |
IVI-FUN-003              |              | Concept      | draft        | Test         |              |              |              |                   |              |              |              |
```

User-defined attributes first, then built-in attributes. Empty cells mean the
requirement didn't set that attribute. Non-string values (arrays, integers,
booleans) are formatted automatically.

### Example: `stats`

```
Requirements: 3
Documents:    1

=== example (3 reqs) ===
  asil:
    B                    1
    QM                   1
  maturity:
    Concept              1
    Production           1
    Prototype            1
  status:
    approved             1
    draft                2
  verify:
    Test                 3
  owner:
    TierOneSupplierA     2
  disposition:
    deferred             1
  disposition-reason:
    Moved to Phase 2...  1
```

### Example: `check --json`

```json
{
  "version": 1,
  "exit_code": 1,
  "summary": {
    "total": 3,
    "valid": 2,
    "invalid": 1,
    "parse_errors": 0,
    "warnings": 6
  },
  "documents": [
    {
      "path": "example",
      "schema_title": "ivi-requirements — IVI Requirements",
      "req_count": 3,
      "requirements": [
        {
          "id": "IVI-FUN-001",
          "valid": true,
          "checks": [
            {"level": "WARNING", "message": "broken reference: target \"SYS-001\" not found"},
            {"level": "WARNING", "message": "broken reference: target \"SAFE-003\" not found"},
            {"level": "WARNING", "message": "no upstream reference"}
          ]
        },
        {
          "id": "IVI-FUN-002",
          "valid": true,
          "checks": [
            {"level": "WARNING", "message": "broken reference: target \"SYS-007\" not found"}
          ]
        },
        {
          "id": "IVI-FUN-003",
          "valid": false,
          "checks": [
            {"level": "ERROR", "message": "validating ivi-requirements: required: missing properties: [\"asil\"]"},
            {"level": "WARNING", "message": "untraced: no downstream reference"},
            {"level": "WARNING", "message": "no upstream reference"}
          ]
        }
      ]
    }
  ],
  "parse_errors": []
}
```

---

## 8. Baseline diff

Compare two requirement baselines anchored by git tags without checking either
one out. `reqmd baseline diff <tag1> <tag2>` extracts the repository at each tag
via `git archive | tar`, parses both versions through the standard pipeline, and
produces a semantic diff:

- **Requirements**: added (green `+`), removed (red `-`), and modified
  (yellow `~`) with attribute-level detail.
- **Schemas**: new properties, removed properties, and changes to `required`
  fields.

Use `--json` for a structured report suitable for CI or release-note generation.
The command always exits `0` — the diff is informational, not validation.

```sh
reqmd baseline diff v1.0.0 v1.1.0
```

```text
Requirements
  + IVI-FUN-004  added in v1.1.0
  - IVI-FUN-002  removed in v1.1.0
  ~ IVI-FUN-001
      asil: QM -> B
      trace: ["SYS-001"] -> ["SYS-001","SAFE-003"]

Schemas
  ~ ivi-requirements
      + property: owner
      - property: safety_relevant
      required: ["asil","maturity","status","verify"] -> ["asil","maturity","status","verify","owner"]
```

The extraction uses `git archive` so no working-tree checkout or extra worktree
is needed; both baselines are processed in temporary directories.

### Submodule changes

When comparing two git tags, `reqmd baseline diff` also reports submodule
changes (pinned commit SHA differences) in a `--- Submodule Changes ---`
section. This is always-on and uses `git ls-tree -t <tag>` to read the
submodule pins. The section is hidden when the repository has no submodules.

Example output:

```
--- Submodule Changes ---
  vendor/spec-a:  a1b2c3d  →  e4f5g6h  (updated)
  vendor/spec-b:  f7g8h9i  →  (removed)
  vendor/spec-c:  (added)   →  j0k1l2m
```

---

## Trace validation in detail

The `trace` attribute links requirements across documents. Each value is a
requirement ID (uppercase letters, digits, hyphens) with optional version pinning
(e.g. `SYS-001~3` means "fulfills SYS-001 version 3").

### Version pins

A downstream requirement can pin the upstream version it was last verified
against. When the upstream `version` is bumped, the pin becomes stale and
reqmd flags the downstream as needing re-verification (change-impact analysis,
ISO 29148).

```yaml
# Upstream declares its current version
```attr
id: UP-001
version: 3
```

```yaml
# Downstream pins the version it was verified against
```attr
id: DN-001
trace: [UP-001~2]   # re-verify against UP-001 v3
```

Findings:

| Condition | Meaning | Level | Demote with |
|-----------|---------|-------|-------------|
| `pin < upstream.version` | Outdated — upstream bumped, downstream needs re-verify | **ERROR** | `--relaxed-versions` |
| `pin > upstream.version` | Predated — downstream claims a non-existent version | **ERROR** | (never — data integrity) |
| `pin == upstream.version` | Current | — | — |
| No pin | No version check applies | — | — |
| Upstream `version: 0` / unset | No ground truth to compare | — | — |
| Upstream `external: true` | Out of scope | — | — |

Suppression per-requirement: `reqmd-suppress: [version-pin]`.

Bulk-fix all outdated pins in a tree with `reqmd repin <root> --yes`
(prints a dry-run change list by default; safe to re-run after
applying).

When multiple documents share the same ID prefix (e.g., two teams both use `STK-`), use a stable `document-id` to qualify references:

```yaml
trace:
  - stakeholder-a/STK-001   # `stakeholder-a` is x-reqmd.document-id, not a directory path
  - STK-002                 # unqualified: works when ID is globally unique
```

If an unqualified reference matches IDs in multiple documents, reqmd reports an
ERROR suggesting the qualified form. Use document-ID-qualified references when
integrating independently-authored documents, including git submodules.


ReqMD builds an **ephemeral graph** of all requirements and their traces, then
runs these checks:

| Check | Level | Condition | Exit code? |
|-------|-------|-----------|------------|
| **Broken reference** | WARNING | `trace` references an ID not found in any document | No |
| **Circular dependency** | ERROR | Cycle detected via DFS along `TRACES` edges | **Yes** |
| **requires-trace-from coverage** | WARNING | When a requirement declares `requires-trace-from: [..]`, at least one inbound edge must originate from each named `document-id` or `level` | No |
| **Untraced (generic)** | WARNING | Fallback: no incoming traces, not a top-boundary dir, not a sub-req, `requires-trace-from` not set | No |
| **No upstream reference (generic)** | WARNING | Fallback: no outgoing traces, not a bottom-boundary dir, `requires-trace-from` not set | No |
| **Disposition without reason** | WARNING | `disposition` is `deferred`/`rejected` but reason is missing | No |
| **Mandatory disposition** | ERROR | `mandatory-disposition: true` and `disposition` is missing | **Yes** |
| **ID prefix mismatch** | ERROR | ID doesn't start with directory's `id-prefix` | **Yes** |
| **ID prefix collision** | ERROR | Two directories declare the same `id-prefix` | **Yes** |
| **Duplicate ID** | ERROR | Same requirement ID defined in two different documents | **Yes** |
| **Ambiguous reference** | ERROR | Unqualified trace reference matches IDs in multiple documents | **Yes** |
| **Version pin predated** | ERROR | `trace: [UP-001~5]` but `UP-001` is at `version: 2` — downstream claims a version that doesn't exist | **Yes** |
| **Missing verdict** | WARNING | An approved verification measure (a requirement with a `verify` attribute) has no verification result tracing to it. Only fires when `--results` is supplied. | No |
| **Failing verdict** | ERROR | A verification measure's latest result has outcome `fail`. Only fires when `--results` is supplied. | **Yes** |

Boundary inference: Directories with no `upstream.sources` are
**top-boundary** (generic untraced suppressed). Directories not referenced by
any other directory's `upstream.sources` are **bottom-boundary**
(generic no-upstream-reference suppressed). For fine-grained control, declare
`requires-trace-from: [..]` on a requirement to specify exactly which document-ids or
levels must trace to it; use `requires-trace-from: []` to explicitly opt out of generic
boundary coverage.

Per-requirement check suppression is available via the `reqmd-suppress` attr:
```yaml
reqmd-suppress:
  - untraced
  - id-prefix
```
Supported suppression names: `broken-ref`, `circular`, `untraced`, `no-downstream`,
`disposition-reason`, `mandatory-disposition`, `id-prefix`, `version-pin`,
`requires-trace-from-coverage`.

Only **ERROR** level checks affect the exit code. WARNING and INFO are
informational.

Supported suppression names: `broken-ref`, `circular`, `untraced`, `no-downstream`,
`disposition-reason`, `mandatory-disposition`, `id-prefix`, `version-pin`,
`requires-trace-from-coverage`, `missing-verdict`, `failing-verdict`.

## FAQ

### How do I trace to a parent document when its path is unknown?

Use a stable `x-reqmd.document-id` and omit `x-reqmd.upstream.sources`. Paths are
not needed to resolve requirement traces. Give each independently maintained
document a unique ID:

```yaml
# system/schema.yaml
x-reqmd:
  document-id: system
  level: system-requirements
  upstream:
    level: stakeholder-needs
    # sources intentionally omitted; the integration repository chooses the path
```

Reference the parent by document ID in Markdown attributes:

```yaml
trace:
  - stakeholder/STK-001
```

Use the same stable IDs in `requires-trace-from`, for example
`requires-trace-from: [software]`. When the documents are assembled, run
`reqmd check` from a common root containing all submodules. ReqMD discovers all
schemas below that root and resolves `document-id/requirement-id` references.

`upstream.sources` remains optional but is still used for HTML document-chain
navigation and path-based boundary inference. Add it only where the assembled
repository has a stable layout. `external: true` is not a path workaround; it
changes the document's validation and boundary semantics.

See [`quickstart/10-submodule-configuration/`](quickstart/10-submodule-configuration/)
for a complete three-level example.

## Disposition workflow

When a stakeholder requirement can't be immediately implemented, use the built-in
`disposition` + `disposition-reason` attributes instead of leaving silent gaps:

```yaml
disposition: deferred
disposition-reason: "Deferred to v2.0 — HW SPI interface not available until next PCB revision"
```

| Disposition | What it signals | Rationale needed? |
|-------------|-----------------|-------------------|
| `implemented` | Actively developed | No |
| `deferred` | Accepted, postponed | **Yes** |
| `rejected` | Not accepted | **Yes** |

Disposition is orthogonal to trace coverage. Use `requires-trace-from: []` to explicitly state that no downstream coverage is expected.

### Worked example — deferred with rationale (passes)

```markdown
## STAKE-007
```attr
status: approved
disposition: deferred
disposition-reason: "Deferred to Phase 2 per steering committee 2025-03-14"
```
The system shall support over-the-air firmware updates for all ECUs.
```
**Result:** ✅ `disposition: deferred` requires reason. Reason present → no
"disposition without reason" warning. `requires-trace-from: []` → no "no upstream reference" warning.
No `trace` → correct for a postponed need.

### Worked example — rejected without reason (warning)

```markdown
## STAKE-008
```attr
status: approved
disposition: rejected
```
The system shall support wireless charging.
```
**Result:** ⚠ WARNING — `disposition: rejected` but no `disposition-reason` provided.

---
## Verification & validation results

reqmd traces the left side of the V-model (stakeholder → system → software →
test specs). The `--results` flag closes the right side by loading ephemeral
verification results and running outcome-gated checks against the
verification measures (requirements with a `verify` attribute).

Results are **ephemeral**: they are loaded per `check` invocation, never
written to the spec repo. This keeps CI run-to-run churn out of git and
out of `baseline diff`.

### Two ingestion paths, one flag

```
reqmd check <root> --results ./ci-out/ --results ./reviews/
```

Each `--results` path is auto-detected:

- **Directory** (walked):
  - `.ctrf.json` → parsed as a [CTRF](https://ctrf.io/) report.
  - Plain `.json` with a CTRF top-level `results` object → parsed as CTRF.
  - Other `.json` (coverage.json, junit exports) → skipped silently.
  - Subdirectories with a `schema.yaml` → loaded as **manual results**
    (markdown `attr` blocks with `outcome`, `verifier`, `evidence`,
    `verified-at`, `trace: [MEASURE-ID]`).
- **File**: parsed as CTRF (`.ctrf.json` or CTRF-shaped `.json`).

### CTRF mapping (automated tests)

Each CTRF `results.tests[]` entry maps to a measure via the `x-reqmd.id`
extra field. The test framework's reporter emits this field per test; it
carries the reqmd measure requirement ID (with optional `~N` version pin).

CTRF `status` → reqmd `outcome`:

| CTRF status | reqmd outcome |
|---|---|
| `passed` | `pass` |
| `failed` | `fail` |
| `skipped` | `skipped` |
| `pending` / `other` / unknown | `inconclusive` |
| any status + `flaky: true` | `inconclusive` |

The CTRF file itself (duration, logs, extra fields) is the "corresponding
verification measure data" every ASPICE record BP requires; it is linked
from the result's `evidence` field.

Unmapped tests (no `x-reqmd.id` in `extra`) are skipped with a WARNING.

### Manual results (review / inspection / analysis)

For non-test methods, author results as markdown with a user-supplied
`schema.yaml` in a directory outside the spec root:

```yaml
# example manual-results schema.yaml
$schema: "https://json-schema.org/draft/2020-12/schema"
$id: "verify-results"
title: "Verification Results"
type: object
required: [outcome, verifier, verified-at]
properties:
  outcome: { enum: [pass, fail, skipped, inconclusive] }
  verifier: { type: string }
  evidence: { type: string, description: "URI/path to minutes, record, or CTRF file" }
  verified-at: { type: string, format: date }
additionalProperties: false
x-reqmd:
  level: verify-results
```

```markdown
## VR-001: Parser review result
```attr
outcome: pass
verifier: "D. Author"
verified-at: 2026-07-15
trace: [TST-FIX-001]
```
Parser review passed.
```

### Outcome-gated checks

When `--results` is supplied, two new checks run alongside the existing
trace checks:

| Check | Level | Condition | Suppression |
|---|---|---|---|
| **missing-verdict** | WARNING | An approved verification measure has no result tracing to it. Draft measures are skipped. | `reqmd-suppress: [missing-verdict]` |
| **failing-verdict** | ERROR | A measure's latest result has outcome `fail`. | `reqmd-suppress: [failing-verdict]` |

The existing **version-pin** check also applies to result→measure traces:
pin a result with `MEASURE-ID~3` against a measure now at `version: 4` and
the `outdated` finding fires (demotable via `--relaxed-versions`). This is
how stale-verdict is detected — no new check, just the existing one on a
new edge type.

### History

Result history is not stored in-file. Each run loads the latest CTRF /
manual results; the previous run's results are discarded. Across all
`--results` inputs, the latest verdict per measure wins by CTRF
`tests[].stop` (ms-epoch) or manual `verified-at`. Run-to-run history lives
in CI artifacts, not in reqmd.

### Exporting results

All export formats support `--results`:

| Format | What `--results` adds |
|--------|----------------------|
| **CSV** | Two extra columns: `Verdict` (pass/fail/skipped/inconclusive) and `Verdict Source` (file path) |
| **HTML** | Color-coded verdict badges on measure cards — green (pass), red (fail), orange (inconclusive), gray (skipped). Tooltip shows outcome + source file. |
| **Graph** | `RESULT:` pseudo-nodes with `outcome` and `source` properties, connected via `TracesTo` edges to their measures. Enables Cypher traversal from requirements through measures to verification results. |

```sh
reqmd export csv requirements/ --results ci-out/ -o docs/
reqmd export html requirements/ --results ci-out/ -o docs/
reqmd export graph requirements/ --results ci-out/ -o graph/
```

Without `--results`, all export formats are unchanged — no verdict columns, badges, or result nodes.

### Scaffolding a results directory

Use `reqmd init` with the `results` preset to scaffold a manual verification
results directory (review, inspection, analysis, demonstration):

```sh
reqmd init my-results/ --preset results
reqmd init my-results/ --preset results --id-prefix REV
```

This creates `schema.yaml` (with `level: verify-results`, required `outcome`/`verifier`/`verified-at` attributes) and `results.md` (one example result). See the [Commands](#commands) table for all `init` flags.

### Custom presets

`reqmd init` accepts a directory path for `--preset` in addition to the
built-in names. A custom preset is a directory containing:

- `schema.yaml.tmpl` + `example.md.tmpl` (Go template files, preferred), or
- `schema.yaml` + any `.md` file (plain files, used verbatim)

Templates support Go `text/template` syntax with four variables:
`{{ .ID }}`, `{{ .Title }}`, `{{ .Level }}`, `{{ .IDPrefix }}`.

```sh
reqmd init my-project/ --preset ./my-preset/ --id-prefix MY
```

---

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | All valid |
| 1 | Validation errors or ERROR-level trace check triggered |
| 2 | Schema or file parse error |

---

## Real-world workflow

```sh
# Scaffold a new requirements project
reqmd init my-project/ --preset generic

# Validate your entire requirements tree
reqmd check requirements/

# List all requirements with all attributes
reqmd ls requirements/

# Check attribute coverage
reqmd stats requirements/

# Export for stakeholder review
reqmd export html requirements/ -o docs/

# Generate CSV for spreadsheet import
reqmd export csv requirements/ -o docs/

# JSON for CI scripting
reqmd check --json requirements/

# Load ephemeral V&V results and run outcome-gated checks
reqmd check requirements/ --results ci-out/ --results reviews/

# Export with verification verdicts (badges in HTML, columns in CSV)
reqmd export html requirements/ --results ci-out/ -o docs/
reqmd export csv requirements/ --results ci-out/ -o docs/

# Scaffold a manual verification results directory
reqmd init reviews/ --preset results --id-prefix VR

# Live preview while editing
reqmd serve requirements/          # browser auto-opens
reqmd serve requirements/ --headless  # terminal-only
reqmd serve requirements/ --results tests/  # with verdict badges
```

Wire `reqmd check requirements/` into CI (GitHub Actions, GitLab CI, etc.)
to catch missing attributes and broken trace links on every PR. Use `--json`
output for programmatic parsing.

---

## Design principles

- **No lock-in.** Your data is plain Markdown and YAML — openable in any editor,
  renderable on GitHub/GitLab, diffable with standard Git tools.
- **No database.** The trace graph is ephemeral, built in a temp directory on
  each invocation and discarded when done. The `*.md` files are the single
  source of truth — nothing is committed except text files.
- **Per-directory schemas.** Different subsystems (IVI, safety, system) can have
  different required fields and validation rules, all validated in one pass.
- **Scale.** Parses thousands of files in parallel using a `runtime.NumCPU()`
  worker pool.
- **Dogfooding.** The reqmd tool's own requirements are defined in `spec/`
  using the reqmd format and validate with `reqmd check spec/`. This ensures
  the format is always production-ready for the team's own use.

---

## Self-hosted requirements

ReqMD dogfoods its own format. The `spec/` directory contains a complete
6-level V-model tree defining the tool itself:

| Level | Directory | Reqs | Description |
|-------|-----------|:----:|-------------|
| External | `spec/00-aspice/` | 191 | Automotive SPICE v4.0 base practices (`external: true`) |
| Stakeholder | `spec/01-stakeholder/` | 5 | Stakeholder goals (top boundary, no upstream) |
| ASPICE SR | `spec/01a-aspice-stakeholder/` | 18 | ASPICE stakeholder requirements mapped to base practices |
| System | `spec/02-system/` | 9 | Feature specifications |
| Software | `spec/03-software/` | 6 | Component-level design |
| Tests | `spec/04-tests/` | 4 | Test specifications (mandatory-disposition) |

All 233 requirements validate cleanly:

```sh
$ reqmd check spec/
# ... 233 total, 233 valid, 0 invalid, 0 parse errors, 197 warnings

$ reqmd serve spec/          # live-reloading HTML preview
$ reqmd export html spec/ -o /tmp/out/
Wrote /tmp/out/00-aspice-requirements.html
Wrote /tmp/out/01-stakeholder-requirements.html
Wrote /tmp/out/01a-aspice-stakeholder-requirements.html
Wrote /tmp/out/02-system-requirements.html
Wrote /tmp/out/03-software-requirements.html
Wrote /tmp/out/04-tests-requirements.html
```

The 196 warnings are expected — external base practices (`00-aspice`) and
stakeholder goals (`01-stakeholder`) have no upstream traces (untraced warnings
suppressed by boundary inference and `external: true`). The ASPICE proxy layer
(`01a-aspice-stakeholder`) is the first fully-traced tier.

### Documentation website

The `site/` directory is a self-contained [Hugo](https://gohugo.io) project
(no external theme) hosting the public docs at **<https://reqmd.dev>**. On
every push to `main` that touches `site/**`, the
[`.github/workflows/hugo.yml`](.github/workflows/hugo.yml) workflow installs
Hugo extended, builds with `hugo --minify -s site`, and deploys `site/public/`
to GitHub Pages via `actions/deploy-pages`.

Build locally to preview:

```sh
hugo server -s site --buildDrafts --disableFastRender
```

One-time repo setup (web UI): **Settings → Pages → Build and deployment →
Source = "GitHub Actions"**, and set the custom domain to `reqmd.dev` (a
`CNAME` file is committed at `site/static/CNAME`). Point the apex domain's DNS
A records at the [GitHub Pages IPs](https://docs.github.com/en/pages/configuring-a-custom-domain-for-your-github-pages-site/managing-a-custom-domain-for-your-github-pages-site).
