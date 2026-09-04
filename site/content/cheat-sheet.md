---
title: "Cheat sheet"
description: "Full reqmd command reference with example outputs and schema.yaml options — every command, every flag, every x-reqmd option."
weight: 4
---

A quick reference for every reqmd command and every `schema.yaml` option. Each command shows real output from the bundled `spec/` tree. Bookmark this page.

## Command overview

### `reqmd check` — validate requirements and trace links

The main command. Recursively finds all `schema.yaml` files under the given directory and validates every `.md` file with requirement `attr` blocks.

```sh
reqmd check spec/                          # validate everything under spec/
reqmd check spec/ --json                   # JSON output for CI parsing
reqmd check spec/ --results ci-out/        # load verification results
reqmd check spec/ --results ci-out/ --results reviews/
reqmd check file.md -s schema.yaml         # single-file mode
reqmd check spec/ --relaxed-versions       # demote version-pin errors to warnings
reqmd check spec/ --filter '"Premium" in variant'   # only matching requirements
reqmd check spec/ --disjoint-check variant # error on cross-configuration trace links
```

**Exit codes:** 0 = all valid, 1 = validation errors, 2 = parse error.

**Checks run:** schema validation, broken references, circular dependencies, missing coverage (`requires-trace-from`), version-pin staleness, missing verdict, failing verdict, disjoint attributes (`--disjoint-check`).

**Verification results:** CTRF `extra.x-reqmd` supports `id` (bind directly to a measure), `case` (stable case identity), `verifies` (requirements a synthesized test case exercises) and `description` (case body). Verdicts roll up strictly over a measure's own results plus its approved downstream test cases (any `fail` → fail, else inconclusive → inconclusive, else skipped → skipped, else pass); deferred/rejected measures are skipped. Uninstrumented tests are skipped silently; a test that declares `x-reqmd` but binds nothing warns as `unbound-result` (suppress with `--ignore-unbound-results`). Result-attributed findings appear in the `Verification results` section / JSON `results` array.

**Filters:** `--filter "<expr>"` scopes the check to requirements matching an expr-lang expression (see [Filter expressions](#filter-expressions)). Coverage is computed within the filtered subset, so a filtered-out requirement can neither require nor provide coverage. `--disjoint-check <attr>` verifies that every trace link has at least one overlapping value for the named array attribute (zero intersection → ERROR). With `--json`, the summary gains a `"filter"` field recording the active expression.

{{< details summary="Example output" >}}
```text
Schema : reqmd-system-requirements — reqmd System Requirements
File   : spec/02-system  (11 requirements)

  ✅  SYS-FMT-001  all attributes valid
  ✅  SYS-FMT-002  all attributes valid
  ✅  SYS-FMT-003  all attributes valid
  ✅  SYS-CLI-001  all attributes valid
  ✅  SYS-VAL-001  all attributes valid

  ⚠  SYS-FMT-001  broken reference: target "STK-GOAL-001" not found
  ⚠  SYS-FMT-001  broken reference: target "ASP-SR-001" not found
  ⚠  SYS-CLI-001  broken reference: target "STK-GOAL-004" not found

Summary: 11 total, 11 valid, 0 invalid, 0 parse errors, 50 warnings
```
{{< /details >}}

{{< details summary="Example JSON output (`--json`)" >}}
```json
{
  "version": 1,
  "exit_code": 0,
  "summary": {
    "total": 11,
    "valid": 11,
    "invalid": 0,
    "parse_errors": 0,
    "warnings": 50
  },
  "documents": [
    {
      "path": "spec/02-system",
      "schema_title": "reqmd-system-requirements — reqmd System Requirements",
      "req_count": 11,
      "requirements": [
        {
          "id": "SYS-FMT-001",
          "valid": true,
          "checks": [
            {
              "level": "WARNING",
              "source": "spec/02-system/features.md",
              "message": "broken reference: target \"STK-GOAL-001\" not found"
            }
          ]
        }
      ]
    }
  ]
}
```
{{< /details >}}

### `reqmd ls` — list all content

Prints a table of every node in document order, with a leading `Type`
column (`req` / `container` / `info`). Requirements show all schema
attributes; containers and info items show their heading text.

```sh
reqmd ls spec/                # table output
reqmd ls spec/ --json         # JSON output (rows array with a type field)
reqmd ls spec/ --filter '"Premium" in variant'   # only matching requirements (items always shown)
```

{{< details summary="Example output" >}}
```text
=== spec/02-system ===
Type       | ID                       | priority     | status       |
---------------------------------------------------------------------
container  | reqmd System Features    |              |              |
req        | SYS-FMT-001: Markdown requirement format | Critical     | approved     |
req        | SYS-FMT-002: Built-in attribute system | Critical     | approved     |
req        | SYS-FMT-003: Schema.yaml document configuration | Critical     | approved     |
req        | SYS-CLI-001: CLI subcommands | Critical     | approved     |
req        | SYS-VAL-001: Multi-pass validation pipeline | Critical     | approved     |
```
{{< /details >}}

### `reqmd stats` — attribute-value breakdown per document

Shows how many requirements have each attribute value, grouped by document directory, plus item counts and a type breakdown.

```sh
reqmd stats spec/             # table output
reqmd stats spec/ --json      # JSON output
reqmd stats spec/ --filter 'priority == "Critical"'   # scope to matching requirements
```

{{< details summary="Example output" >}}
```text
Requirements: 19
Documents:    1

=== spec/02-system (19 reqs, 1 item) ===
  type:
    container            1
    req                  19
  priority:
    Critical             6
    High                 8
    Medium               5
  status:
    approved             19
```
{{< /details >}}

### `reqmd init` — scaffold a new project

Creates a directory with `schema.yaml` and an example `.md` file.

```sh
reqmd init my-project/                              # generic template (default)
reqmd init my-project/ --preset aspice --id-prefix REQ  # ASPICE-oriented template with safety attributes
reqmd init my-project/ --preset results --id-prefix VR  # manual verification results template
reqmd init my-project/ --preset ./my-preset/        # custom preset directory
reqmd init my-project/ --id-prefix REQ --title "My Requirements"
reqmd init my-project/ --force                      # overwrite existing
```

**Built-in presets:** `generic` (default), `aspice` (ASPICE-oriented with safety attributes), `results` (manual verification results).

{{< details summary="Example output" >}}
```text
Initialized generic requirements in /tmp/my-project/
  /tmp/my-project/schema.yaml
  /tmp/my-project/requirements.md
Run: reqmd check /tmp/my-project/
```

Generated `schema.yaml`:

```yaml
x-reqmd:
  level: "requirements"
  id-prefix: REQ
$schema: "https://json-schema.org/draft/2020-12/schema"
$id: "my-project"
title: "my-project Requirements"
type: object
properties:
  owner:
    type: string
    description: "Responsible person or team"
additionalProperties: false
```

Generated `requirements.md` (excerpt):

````text
## REQ-001
```attr
status: draft
owner: Team Alpha
```
The system shall provide a login mechanism.

*Rationale: Authentication is required for secure access.*
````
{{< /details >}}

**Flags:**

| Flag | Description |
|------|-------------|
| `--preset <name\|dir>` | Built-in preset name or path to a custom preset directory |
| `--id-prefix <prefix>` | ID prefix for requirements (e.g. `SYS-`, `REQ-`) |
| `--id <id>` | Schema `$id` (default: directory basename) |
| `--title <title>` | Schema title (default: `<dir> Requirements`) |
| `--level <level>` | `x-reqmd.level` value (default: `requirements`) |
| `--force` | Overwrite existing files |

### `reqmd serve` — live-reloading HTML preview

Watches the spec tree and serves a live HTML preview via HTTP. Re-parses, re-checks, and re-exports on every file change. The browser auto-reloads via Server-Sent Events.

```sh
reqmd serve spec/                           # default: localhost:8080
reqmd serve spec/ --addr :9090              # custom port
reqmd serve spec/ --no-open                 # don't auto-open browser
reqmd serve spec/ --headless                # terminal-only, no HTTP server
reqmd serve spec/ --results tests/ --results reviews/
reqmd serve spec/ --filter '"Sport" in variant'   # preview one configuration
```

{{< details summary="Example output" >}}
```text
reqmd serve spec/
  serving http://localhost:8080
  watching spec/ ...
  initial build: 249 reqs, 6 docs
  file changed: spec/02-system/features.md
  rebuilt: 249 reqs (48ms)
  file changed: spec/02-system/schema.yaml
  rebuilt: 249 reqs (52ms)
```
{{< /details >}}

| Flag | Description |
|------|-------------|
| `--addr <addr>` | HTTP server address (default: `localhost:8080`) |
| `--results <path>` | Load verification results (repeatable). Also watched for changes. |
| `--no-open` | Don't open a browser on start |
| `--headless` | Terminal-only mode (no HTTP server) |
| `--debounce <dur>` | Debounce window for file events (default: 500ms) |
| `--filter <expr>` | Serve only requirements matching an expr-lang expression |

### `reqmd export csv` — export to CSV

One CSV file per document directory (`<dirname>-requirements.csv`).

```sh
reqmd export csv spec/ -o csv-out/
reqmd export csv spec/ -o csv-out/ --results ci-out/
reqmd export csv spec/ -o csv-out/ --filter '"Premium" in variant'
```

With `--results`, adds `Verdict` and `Verdict Source` columns.

{{< details summary="Example output" >}}
```text
exporting spec/ → csv-out/
  spec/02-system  → csv-out/system-requirements.csv   (11 rows)
  spec/03-software → csv-out/software-requirements.csv (17 rows)
  spec/04-tests   → csv-out/tests-requirements.csv     (7 rows)
  ...
  6 documents, 249 requirements exported
```

Each CSV has columns: `Type, ID, Title, <schema attributes>, Body, Rationale`.
Rows are emitted in document order for every node — requirements and
containers/info items alike.
{{< /details >}}

### `reqmd export html` — export to HTML

Standalone HTML files with trace links, document chain, theme toggle, search.

```sh
reqmd export html spec/ -o html-out/
reqmd export html spec/ -o html-out/ --results ci-out/ --results reviews/
reqmd export html spec/ -o html-out/ --filter '"Sport" in variant'
```

With `--results`, renders color-coded verdict badges (pass/fail/skipped/inconclusive) on measure cards.

{{< details summary="Example output" >}}
```text
exporting spec/ → html-out/
  spec/02-system  → html-out/system-requirements.html
  spec/03-software → html-out/software-requirements.html
  spec/04-tests   → html-out/tests-requirements.html
  ...
  6 documents exported
```

Each HTML file is standalone (no external JS), with requirement cards,
collapsible container sections and info blocks for headings without
`attr` blocks, a sidebar TOC with folder nesting, trace links, document
chain, search, and theme toggle.

[**See a live HTML export →**](/export-sample/02-system-requirements.html) (from the reqmd project's own spec)
{{< /details >}}

### `reqmd export graph` — export to LadybugDB graph

Creates a graph database with one node per requirement and `TracesTo` edges. Requires the `ladybug` build tag.

```sh
go build -tags ladybug -o reqmd ./cmd/reqmd
reqmd export graph spec/ -o graph-out/
```

{{< details summary="Example output" >}}
```text
exporting spec/ → graph-out/reqmd-graph.lbug
  249 nodes, 1240 edges
  done
```

Query with the `lbug` CLI: `lbug graph-out/reqmd-graph.lbug`
{{< /details >}}

### `reqmd baseline diff` — compare two git tags

Extracts the spec tree at each tag (no checkout), parses both, and produces a semantic diff.

```sh
reqmd baseline diff v1.0 v2.0
reqmd baseline diff v1.0 v2.0 --json
reqmd baseline diff HEAD~10 HEAD
reqmd baseline diff v1.0 v2.0 --filter '"Base" in variant'   # one configuration across two tags
reqmd baseline diff --filter-a 'variant == nil or "Base" in variant' \
                    --filter-b '"Premium" in variant' HEAD    # two views of one commit
```

With `--filter`, both snapshots are scoped to matching requirements before diffing. With `--filter-a`/`--filter-b` (used together, mutually exclusive with `--filter`), a single snapshot at the given ref (default `HEAD`) is compared two ways — the "what does Premium add over Base" report from one commit. `--filter-a` and `--filter-b` are required together.

Reports added, removed, and modified requirements (with attribute-level detail), schema changes, and submodule pin changes. Always exits 0.

{{< details summary="Example output" >}}
```text
Requirements
  + IVI-FUN-004  added in v2.0
  - IVI-FUN-002  removed in v2.0
  ~ IVI-FUN-001
      priority: medium -> high
      trace: ["SYS-001"] -> ["SYS-001","SAFE-003"]

Schemas
  ~ ivi-requirements
      + property: owner
      - property: priority
      required: ["priority","maturity","status","verify"] -> ["maturity","status","verify","owner"]

--- Submodule Changes ---
  vendor/spec-a: a1b2c3d -> e4f5g6h (updated)
```
{{< /details >}}

### `reqmd repin` — update version pins

Updates `~N` version pins in trace references to match the upstream's current version. Dry-run by default; pass `--yes` to apply.

```sh
reqmd repin spec/                                    # dry-run: list proposed changes
reqmd repin spec/ --yes                              # apply changes
reqmd repin spec/ --yes --promote-unpinned           # also pin refs with no ~N
reqmd repin spec/ --json                             # machine-readable output
```

Rewrites only the `attr` blocks; surrounding Markdown is preserved. Idempotent — a second run after the first reports no changes. Predated pins (pin > upstream version) are surfaced but never auto-fixed.

{{< details summary="Example output" >}}
Dry-run (default):

```text
spec/03-software/dn.md  DN-001  UP-001 → UP-001~3
spec/03-software/dn.md  DN-002  UP-002 → UP-002~7
2 outdated, 0 unpinned, 0 predated across 1 files.
```

Apply:

```text
spec/03-software/dn.md  DN-001  UP-001 → UP-001~3
spec/03-software/dn.md  DN-002  UP-002 → UP-002~7
2 outdated, 0 unpinned, 0 predated across 1 files.
Apply 2 changes across 1 files? [y/N] y
applied 2 changes across 1 files
```

No changes needed:

```text
no version-pin changes needed
```
{{< /details >}}

| Flag | Description |
|------|-------------|
| `--yes`, `-y` | Apply changes without prompting |
| `--json` | Output as JSON |
| `--promote-unpinned` | Also pin trace refs that have no `~N` against a versioned upstream |
| `--dry-run` | Print the change list without applying (default; implicit when `--yes` is absent) |

## Filter expressions

`--filter "<expr>"` is available on `check`, `ls`, `stats`, `export csv`, `export html`, `serve`, and `baseline diff`. The expression is evaluated against each requirement's attributes (plus the built-ins `id`, `title`, `status`, `disposition`, `trace`, `version`), and the command runs only on matching requirements. It's the generic primitive behind variant management — the same flag slices by variant, platform, region, priority, or any other attribute.

| Expression | Matches |
|---|---|
| `"Premium" in variant` | Requirements tagged Premium (array membership) |
| `variant == nil` | Requirements with no `variant` attribute — "common to all" |
| `"Premium" in variant or variant == nil` | The whole Premium configuration (Premium-specific plus common) |
| `status == "approved"` | Approved requirements |
| `version > 3` | Requirements with `version` greater than 3 |
| `id startsWith "SYS-"` | Requirements whose ID starts with `SYS-` |
| `title contains "boot"` | Requirements whose title contains "boot" |
| `"Base" in variant and "Premium" in variant` | Multi-configuration requirements |
| `not "Sport" in variant` | Everything except Sport-tagged |

**Rules:**

- The engine is [`expr-lang`](https://github.com/expr-lang/expr); supported operators include `==`, `!=`, `<`, `>`, `<=`, `>=`, `in`, `and`, `or`, `not`, `contains`, `startsWith`, `endsWith`.
- A filter referencing an attribute not declared in any `schema.yaml` is a compile-time ERROR (fails fast on typos, before any requirement is evaluated).
- A filter with invalid syntax is a compile-time ERROR, reported once.
- A requirement lacking a referenced attribute evaluates that variable to `nil` and is excluded unless the expression handles absence (e.g. `variant == nil`).
- The `--json` summary includes the active expression in a `"filter"` field.

**Filter-aware coverage:** with `--filter`, coverage (`requires-trace-from`) is computed within the filtered subset — a filtered-out requirement can neither require nor provide coverage, so per-configuration checks don't produce false "missing coverage" failures.

**Disjoint-attribute check:** `--disjoint-check <attr>` verifies that every trace link has at least one overlapping value for the named array attribute. Zero intersection → ERROR; empty/absent value = "applies to all" (exempt). Enable via CLI flag (repeatable) or `x-reqmd.disjoint-check` in `schema.yaml`.

## All commands at a glance

| Command | What it does | Key flags |
|---------|-------------|-----------|
| `check <dir>` | Validate + trace checks | `--json`, `--results`, `--relaxed-versions`, `--filter`, `--disjoint-check`, `--ignore-unbound-results`, `-s` |
| `ls <dir>` | List all requirements | `--json`, `--filter` |
| `stats <dir>` | Stats per document | `--json`, `--filter` |
| `init <dir>` | Scaffold a new project | `--preset`, `--id-prefix`, `--force` |
| `serve <dir>` | Live HTML preview | `--addr`, `--results`, `--headless`, `--filter` |
| `export csv <dir>` | CSV export | `-o`, `--results`, `--filter` |
| `export html <dir>` | HTML export | `-o`, `--results`, `--filter` |
| `export graph <dir>` | Graph export (ladybug tag) | `-o` |
| `baseline diff <t1> <t2>` | Compare two git tags | `--json`, `--filter`, `--filter-a`, `--filter-b` |
| `repin <dir>` | Update version pins to upstream's current version | `--yes`, `--json`, `--promote-unpinned` |

## The `schema.yaml` file

Every document directory has a `schema.yaml` that defines what attributes a requirement must have. It's a standard JSON Schema 2020-12 file in YAML, plus a `x-reqmd` extension block for reqmd-specific options.

### Full example

```yaml
$schema: "https://json-schema.org/draft/2020-12/schema"
$id: "my-system-requirements"
title: "System Requirements"
type: object
required:
  - priority
  - verify
properties:
  priority:
    type: string
    enum: [Critical, High, Medium, Low]
  verify:
    type: string
    enum: [Test, Review, Inspection, Analysis, Demonstration]
  status:
    type: string
    enum: [draft, approved]
  trace:
    type: array
    items:
      type: string
  version:
    type: integer
additionalProperties: false

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
```

### `x-reqmd` options

| Option | Type | Description |
|--------|------|-------------|
| `level` | string | A label naming this document's V-model layer (e.g. `stakeholder-needs`, `system-requirements`, `software-requirements`, `test-specs`). Free-form string. **Only functional use:** a `requires-trace-from` token may name a `level` instead of a `document-id`; reqmd resolves it to every directory declaring that level. Not used for HTML chain navigation or boundary inference (those use `upstream.sources`). See the [FAQ](/faq/#what-is-the-level-directive-in-x-reqmd-for). |
| `document-id` | string | A short identifier for this document directory (e.g. `system`, `stakeholder`, `tests`). Used in qualified references (`document-id/ID`) and as a `requires-trace-from` token. |
| `id-prefix` | string | The ID prefix for requirements in this directory (e.g. `SYS-`, `STK-GOAL-`, `TST-`). Requirements get IDs like `SYS-001`, `SYS-002`. |
| `upstream` | object | Declares which document directories are valid upstream sources for trace links. |
| `upstream.level` | string | The `x-reqmd.level` of the expected upstream. **Informational only** — never read by the tool. The wiring that matters is `upstream.sources`. |
| `upstream.sources` | array | List of relative paths to upstream document directories. reqmd reads their `schema.yaml` to resolve trace targets. **This** drives HTML document-chain navigation and path-based boundary inference. |
| `mandatory-disposition` | boolean | If `true`, every requirement in this directory must carry a `disposition` attribute (`implemented`, `deferred`, or `rejected`). Used for test specs and verification measures. |
| `external` | boolean | If `true`, requirements in this directory are external reference requirements (imported, not authored). They can be traced *to* but don't need upstream traces. |
| `additional-status-values` | array | Extend the built-in `status` enum (`[draft, approved]`) with custom values (e.g. `[review, withdrawn]`). Lowercase only, no built-in collision, no duplicates. |
| `ignore-status` | boolean | If `true`, all requirements in this directory are treated as coverage providers regardless of status. The status filter is hidden in HTML export. |
| `requires-trace-from` | array | Document-wide default coverage expectation (same `document-id`/`level` tokens as the built-in attribute). Requirements that don't declare their own `requires-trace-from` inherit it; an explicit `requires-trace-from: []` on a requirement overrides it; an empty default opts every requirement in the document out. |
| `disjoint-check` | string \| array | Enables the disjoint-attribute check for the named array-typed attribute(s) (e.g. `variant`) on every `reqmd check`. Same effect as the `--disjoint-check <attr>` CLI flag. |

### Standard JSON Schema fields

These are the JSON Schema 2020-12 fields you'll use most:

| Field | Description |
|-------|-------------|
| `$schema` | Always `"https://json-schema.org/draft/2020-12/schema"`. |
| `$id` | A relative URI identifying this schema. Resolves against the file's base URI. |
| `title` | Human-readable title. Shown in HTML export and `stats` output. |
| `type` | Always `object` (each requirement is an object of attributes). |
| `required` | Array of attribute names that must be present on every requirement. |
| `properties` | Defines each attribute: its `type`, `enum`, `description`, etc. |
| `additionalProperties` | Set to `false` to reject attributes not listed in `properties`. Recommended. |

A custom classification attribute (the basis for variant management) is a plain JSON Schema property — an array with an `enum`:

```yaml
properties:
  variant:
    type: array
    items:
      type: string
      enum: [Base, Premium, Sport]
additionalProperties: false
```

The `enum` gives free validation (a typo like `variant: [Premuim]` fails `reqmd check`), and `"Premium" in variant` works as a `--filter` expression without any reqmd-specific declaration. To also reject trace links between incompatible variants, add `x-reqmd.disjoint-check: variant`. See [Step 14 — Variant management](/quickstart/14-variant-management/).

### Built-in attributes

reqmd recognizes these attribute names automatically (they don't need to be in your `properties`, but they can be):

| Attribute | Description |
|-----------|-------------|
| `trace` | Array of upstream IDs this requirement traces to. Supports `~N` version pins (e.g. `SYS-001~3`). |
| `requires-trace-from` | Declares that this requirement expects downstream coverage. Used for coverage checks. Omitted on a requirement → inherits the document's `x-reqmd.requires-trace-from` default when one is set; `[]` opts out. |
| `status` | `draft` or `approved`. Only `approved` counts as coverage. |
| `disposition` | `implemented`, `deferred`, or `rejected`. Required when `mandatory-disposition: true`. |
| `version` | Integer version of the requirement. Used by version-pin staleness checks. |
| `verify` | Verification method: `Test`, `Review`, `Inspection`, `Analysis`, or `Demonstration`. |
| `reqmd-suppress` | Array of check names to suppress (e.g. `[version-pin, missing-verdict]`). |
| `external` | If `true`, this is an external reference requirement (imported, not authored in this tree). |

### Content model — nothing is dropped

Every heading in a document is a node in a content tree:

| Type | Definition | Shown in |
|------|-----------|----------|
| `req` | heading followed by an `attr` block | cards, `ls`/CSV rows, all checks |
| `container` | heading without an `attr` block that has children | collapsible folder sections in HTML, `container` rows in `ls`/CSV, TOC folders |
| `info` | heading without an `attr` block and no children, or prose before the first heading | plain blocks in HTML, `info` rows in `ls`/CSV |

Containers and info items carry no attributes and skip validation, trace
checks, coverage, and `--filter` pruning — they exist so exported
documents stay faithful to the source. `check` is requirement-scoped.

## The requirement format

Each requirement is a heading, a YAML `attr` block in a fenced code block, and free-form prose. Any heading level works — the first requirement heading anchors the hierarchy, and deeper requirement headings nest under the nearest shallower requirement as sub-requirements:

````text
## SYS-001  Temperature warning

```attr
priority: High
status: approved
trace:
  - STK-003~2
verify: Test
version: 1
```

The system shall display a warning when cabin temperature
exceeds 80°C for more than 5 seconds of continuous operation.

*Rationale:* Driver distraction from sudden thermal events
is a safety concern.
````

The heading text is the requirement ID plus a title. The `attr` block holds the structured attributes. The prose is the requirement body. The `Rationale:` paragraph is optional but shown in the HTML export.

## See also

- [Quickstart](/quickstart/) — a hands-on tour of every command.
- [Step 14 — Variant management](/quickstart/14-variant-management/) — the `--filter` workflow end to end.
- [FAQ](/faq/) — answers to common questions.
- [Use cases](/use-cases/) — what reqmd is for, and what it's not.