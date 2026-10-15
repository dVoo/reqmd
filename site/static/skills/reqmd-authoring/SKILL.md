---
name: reqmd-authoring
description: Create, write, and maintain reqmd requirement documents. Use to scaffold a document directory with reqmd init, author schema.yaml (JSON Schema 2020-12 plus the x-reqmd block), compose requirement headings with fenced attr YAML blocks, apply id-prefix/document-id/level/upstream rules, use built-in attributes (trace, status, disposition, version, verify, requires-trace-from, reqmd-suppress), keep container and info headings, and evolve a spec tree without breaking the trace graph. Triggers include "add a requirement", "new spec directory", "write schema.yaml", "attr block", "id-prefix", "document-id", and "restructure requirements".
---

# Authoring reqmd documents

How to create, write, and maintain reqmd Markdown. For running commands and
fixing findings, use the `reqmd-usage` skill.

## When to use

- Add a requirement, a document directory, or a whole spec tree.
- Write or change a `schema.yaml`.
- Decide between a requirement, a container, and an info heading.
- Introduce a new attribute, status value, or level.
- Restructure a document without breaking existing trace links.

## Ground rules

1. **Read the directory's `schema.yaml` before adding a requirement.** With
   `additionalProperties: false`, an undeclared attribute fails validation.
2. **The heading is the entity.** The ID in the heading is the trace target;
   never rename it. To retire a requirement, keep the heading and set
   `disposition: rejected` with a reason.
3. **Preserve prose.** A change is a heading/`attr` edit, not a file rewrite.
   Keep the rationale and history.
4. **Nothing is dropped.** Headings without `attr` blocks are content
   (containers/info) — do not delete them to "tidy up".
5. **One document directory = one `schema.yaml`.** Different subsystems can
   have different schemas.
6. **Run `reqmd check <root>` after every edit.** See `reqmd-usage`.

## Directory layout

Any directory containing a `schema.yaml` is a document directory. reqmd
validates every `.md` directly inside it (non-recursive), then recurses into
subdirectories. Document directories nest naturally by V-model level:

```
requirements/
├── 01-stakeholder/
│   ├── schema.yaml            # level: stakeholder-needs
│   └── goals.md
├── 02-system/
│   ├── schema.yaml            # level: system-requirements
│   └── features.md
├── 03-software/
│   ├── schema.yaml
│   └── components.md
└── 04-tests/
    ├── schema.yaml            # mandatory-disposition: true
    └── boot-test.md
```

Scaffold with a preset instead of hand-writing the first files:

```sh
reqmd init my-feat/ --id-prefix SYS                    # generic (default)
reqmd init my-feat/ --preset aspice --id-prefix REQ    # ASPICE template
reqmd init reviews/ --preset results --id-prefix VR    # manual V&V results
reqmd init my-feat/ --preset ./my-preset/              # custom preset dir
```

## Requirement anatomy

A requirement is a level-2+ heading, a fenced `attr` block (YAML), then
free-form prose. An optional title follows the ID after a colon and space.

````markdown
## SYS-001: Boot sequence

```attr
status: approved
priority: High
verify: Test
version: 1
trace: [stakeholder/STK-GOAL-001]
requires-trace-from: [software]
```

The system shall complete the boot sequence within 2 seconds of ignition on.

*Rationale:* Fast boot is the top perceived-quality driver.
````

- The heading text is the ID plus optional title.
- The `attr` fence language must be exactly `attr` — the parser looks for it.
- `*Rationale:*` is extracted and rendered separately; body prose is everything
  else.
- Sub-requirements nest under the nearest shallower requirement heading.

### ID rules

- ID pattern: `^[A-Z][A-Z0-9]*(-[A-Z0-9]+)*$` with a trailing number segment —
  e.g. `SYS-001`, `STK-GOAL-002`, `SYS-FMT-003`.
- No colons or spaces inside the ID; they delimit the optional title
  (`## SYS-001: Boot sequence`).
- When `x-reqmd.id-prefix` is set, every ID in that directory must start with
  it (otherwise ERROR).
- Two directories may not declare the same `id-prefix` (ERROR); two documents
  may not define the same ID (ERROR).
- Requirement headings must be **h2 or higher**. An `h1` is the document title
  and never a requirement; an `h1` with an `attr` block is a parse error.

## Content model: requirement, container, info

Every heading becomes a node. Choose deliberately:

| Kind | When | Behaviour |
|---|---|---|
| `req` | heading followed by an `attr` block | validated, traced, filtered, exported |
| `container` | heading without `attr` block that has child headings | folder-like grouping; collapsible in HTML, `container` rows in `ls`/CSV |
| `info` | heading without `attr` block and no children, or leading prose | plain content block; `info` rows in `ls`/CSV |

Containers and info items carry no attributes and skip all validation, trace,
and coverage checks. `--filter` prunes requirements but never removes them.
Use them for document structure, overviews, and explanatory prose — anything
that is not itself a requirement.

```markdown
# System Overview            ← h1: document title (not a node)

Intro prose.                 ← info

## Functional Requirements   ← container (no attr, has children)

### SYS-001: Boot            ← requirement

### SYS-002: Shutdown        ← requirement

## Glossary                  ← info (no attr, no children)
```

## `schema.yaml`

Standard JSON Schema 2020-12 (YAML-serialized) plus an optional `x-reqmd`
extension block for reqmd-specific document metadata.

````yaml
$schema: "https://json-schema.org/draft/2020-12/schema"
$id: "system-requirements"          # relative URI; resolves against the file path
title: "System Requirements"
type: object                        # always object (a requirement's attributes)
required: [priority, verify]
properties:
  priority:
    type: string
    enum: [Critical, High, Medium, Low]
  verify:
    type: string
    enum: [Test, Review, Inspection, Analysis, Demonstration]
  variant:
    type: array
    items:
      type: string
      enum: [Base, Premium, Sport]
additionalProperties: false         # reject undeclared attributes

x-reqmd:
  level: system-requirements
  document-id: system
  id-prefix: SYS-
  requires-trace-from: [software-requirements]
  upstream:
    level: stakeholder-needs
    sources:
      - ../01-stakeholder/
  mandatory-disposition: true
  external: false
````

### `x-reqmd` fields

| Field | Purpose |
|---|---|
| `level` | Labels the V-model layer. Functional use: a `requires-trace-from` token may name a level to match every document at that layer. |
| `document-id` | Stable short identifier; used in qualified `trace: [doc-id/ID]` refs and in `requires-trace-from`. |
| `id-prefix` | Required ID prefix for this directory. |
| `upstream` | `level` is informational; `sources` lists relative paths of upstream dirs and drives HTML doc-chain navigation and path-based boundary inference. |
| `requires-trace-from` | Document-wide default coverage expectation, inherited by requirements without their own attribute. Empty list opts the document out. |
| `mandatory-disposition` | `true` makes a missing `disposition` an ERROR. |
| `external` | `true` marks imported/generated reference requirements: trace targets, not authored here; untraced warnings suppressed; pins out of scope. |
| `additional-status-values` | Lowercase extensions to the built-in status enum. |
| `ignore-status` | `true` makes all requirements coverage providers regardless of status; hides the status filter in HTML. |
| `disjoint-check` | Array-typed attribute(s) that trace-linked requirements must overlap on. |
| `url` | Human link for the HTML export (meaningful when `external: true`). |
| `source.path` / `source.format` | Input glob and format for scanned/proxy documents. |

Use stable `document-id`s for anything assembled from submodules or independent
repos, and qualify trace refs with them.

### Standard JSON Schema fields to use

`$schema`, `$id` (relative), `title`, `type: object`, `required`,
`properties` (`type`/`enum`/`description`/`items`), and
`additionalProperties: false`. Declare in `required` every attribute that must
be present on each requirement.

### Built-in attributes — do NOT redeclare

reqmd injects these automatically. Listing them in `properties` is a schema
compile error.

| Attribute | Meaning |
|---|---|
| `trace` | Upstream refs, e.g. `[SYS-001, stakeholder/STK-002, SYS-001~3]`. |
| `requires-trace-from` | Expected downstream coverage (document-ids or levels); `[]` opts out. |
| `status` | `draft` or `approved` (default `approved`). |
| `disposition` | `implemented`, `deferred`, `rejected`. |
| `disposition-reason` | Required for `deferred`/`rejected`. |
| `version` | Integer version used by pin staleness checks. |
| `verify` | `Test`, `Review`, `Inspection`, `Analysis`, `Demonstration`. |
| `reqmd-suppress` | Checks to suppress for this requirement. |
| `external` | Per-requirement external marker. |

## Status, disposition, coverage

- `draft` means visible but **not a coverage provider**. `approved` (or omitted)
  counts.
- `deferred`/`rejected` require `disposition-reason`.
- Coverage expectations: per requirement, or once per document:
  `x-reqmd.requires-trace-from: [software, tests]`. A requirement attribute
  replaces the document default entirely; `[]` opts that requirement out.
- Only add `requires-trace-from: []` when no downstream coverage is genuinely
  expected — it is a contract, not a way to silence a warning.
- Extend status only when the built-ins are genuinely too narrow:
  `x-reqmd.additional-status-values: [review]` (lowercase, no collisions,
  no duplicates).

## Version pins

Add `version: N` to an upstream that changes over time. Downstreams that
verified against a specific revision record it on the trace ref:

````markdown
## DN-001
```attr
status: approved
trace: [UP-001~3]
```
````

Keep pins current with `reqmd repin` (see `reqmd-usage`). Never author a pin
higher than the upstream's current version — predated pins are data-integrity
errors.

## Frontmatter and body features

Each `.md` file may start with YAML frontmatter. `description` is rendered
below the document header in HTML; keys from multiple files merge
first-file-wins.

```markdown
---
description: "This document defines the **system requirements** for boot."
---
```

Body Markdown supports GFM tables/task lists, mermaid blocks, `::: {.class}`
fenced divs, syntax highlighting, emoji shortcodes, and KaTeX (`$x^2$`).

## Maintenance workflow

1. **Read** the target document's `schema.yaml` and a neighboring requirement
   for conventions (`reqmd ls <root>` shows current attributes).
2. **Edit** the heading/`attr`/prose, or create the new directory with
   `reqmd init`.
3. **Check** `reqmd check <root>`; fix ERROR-level findings, then confirm no
   new WARNING-level findings.
4. **Migrate deliberately.** When renaming/merging requirements, keep IDs,
   update every inbound `trace:`, and re-run coverage checks. For large waves,
   use `reqmd repin --yes` for pins and `reqmd baseline diff` to review the
   semantic change before/after.
5. **Keep `x-reqmd.document-id` stable** across directory moves; that is what
   makes qualified refs survive restructuring and submodule assembly.

## Anti-patterns

- Adding an attribute not declared in `properties` (fails under
  `additionalProperties: false`).
- Dropping the `attr` fence language, or writing `yaml`/no language.
- Using an `h1` for a requirement.
- Putting a colon/space in an ID, or renaming an ID.
- Deleting container/info headings to simplify the file.
- Writing paths in `trace:` instead of `document-id`-qualified IDs.
- Using `external: true` as a path workaround — it changes validation and
  boundary semantics.
- Editing generated/imported or `RESULT:`/`TC:` nodes by hand.
- Leaving `disposition: deferred|rejected` without a reason.

## Verification

- `reqmd check <root>` exits 0; `--json` has no ERROR-level checks and no new
  warnings.
- `reqmd ls <root>` shows the new requirement with the attributes you intended,
  and containers/info where expected.
- Every new `trace:` resolves to an existing ID (`reqmd ls <root> | grep <ID>`).
- `reqmd stats <root>` reflects the new attribute values.
