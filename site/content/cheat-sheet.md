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
```

**Exit codes:** 0 = all valid, 1 = validation errors, 2 = parse error.

**Checks run:** schema validation, broken references, circular dependencies, missing coverage (`requires-trace-from`), version-pin staleness, missing verdict, failing verdict.

<details>
<summary>Example output</summary>

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

</details>

<details>
<summary>Example JSON output (<code>--json</code>)</summary>

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

</details>

### `reqmd ls` — list all requirements

Prints a table of every requirement with all schema attributes.

```sh
reqmd ls spec/                # table output
reqmd ls spec/ --json         # JSON output
```

<details>
<summary>Example output</summary>

```text
=== spec/02-system ===
ID                       | priority     | trace                                    | status       |
-------------------------------------------------------------------------------------------------------------------------
SYS-FMT-001: Markdown requirement format | Critical     | ["STK-GOAL-001","ASP-SR-001"]            | approved     |
SYS-FMT-002: Built-in attribute system | Critical     | ["STK-GOAL-002","ASP-SR-007"]            | approved     |
SYS-FMT-003: Schema.yaml document config | Critical     | ["STK-GOAL-003","ASP-SR-002"]            | approved     |
SYS-CLI-001: CLI subcommands | Critical     | ["STK-GOAL-001","STK-GOAL-004"]          | approved     |
SYS-VAL-001: Multi-pass validation | Critical | ["STK-GOAL-003","STK-GOAL-005"]      | approved     |
```

</details>

### `reqmd stats` — attribute-value breakdown per document

Shows how many requirements have each attribute value, grouped by document directory.

```sh
reqmd stats spec/             # table output
reqmd stats spec/ --json      # JSON output
```

<details>
<summary>Example output</summary>

```text
Requirements: 11
Documents:    1

=== spec/02-system (11 reqs) ===
  priority:
    Critical             6
    High                 5
  status:
    approved             11
```

</details>

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

<details>
<summary>Example output</summary>

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

<pre><code>## REQ-001
&#96;&#96;&#96;attr
status: draft
owner: Team Alpha
&#96;&#96;&#96;
The system shall provide a login mechanism.

*Rationale: Authentication is required for secure access.*</code></pre>

</details>

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
```

<details>
<summary>Example output</summary>

```text
reqmd serve spec/
  serving http://localhost:8080
  watching spec/ ...
  initial build: 834 reqs, 17 docs, 542 warnings
  file changed: spec/02-system/features.md
  rebuilt: 834 reqs, 542 warnings (48ms)
  file changed: spec/02-system/schema.yaml
  rebuilt: 834 reqs, 542 warnings (52ms)
```

</details>

| Flag | Description |
|------|-------------|
| `--addr <addr>` | HTTP server address (default: `localhost:8080`) |
| `--results <path>` | Load verification results (repeatable). Also watched for changes. |
| `--no-open` | Don't open a browser on start |
| `--headless` | Terminal-only mode (no HTTP server) |
| `--debounce <dur>` | Debounce window for file events (default: 500ms) |

### `reqmd export csv` — export to CSV

One CSV file per document directory (`<dirname>-requirements.csv`).

```sh
reqmd export csv spec/ -o csv-out/
reqmd export csv spec/ -o csv-out/ --results ci-out/
```

With `--results`, adds `Verdict` and `Verdict Source` columns.

<details>
<summary>Example output</summary>

```text
exporting spec/ → csv-out/
  spec/02-system  → csv-out/system-requirements.csv   (11 rows)
  spec/03-software → csv-out/software-requirements.csv (6 rows)
  spec/04-tests   → csv-out/tests-requirements.csv     (7 rows)
  ...
  17 documents, 834 requirements exported
```

Each CSV has columns: `ID, Title, <schema attributes>, Body, Rationale`.

</details>

### `reqmd export html` — export to HTML

Standalone HTML files with trace links, document chain, theme toggle, search.

```sh
reqmd export html spec/ -o html-out/
reqmd export html spec/ -o html-out/ --results ci-out/ --results reviews/
```

With `--results`, renders color-coded verdict badges (pass/fail/skipped/inconclusive) on measure cards.

<details>
<summary>Example output</summary>

```text
exporting spec/ → html-out/
  spec/02-system  → html-out/system-requirements.html
  spec/03-software → html-out/software-requirements.html
  spec/04-tests   → html-out/tests-requirements.html
  ...
  17 documents exported
```

Each HTML file is standalone (no external JS), with card-based layout, trace links, document chain, search, and theme toggle.

<a href="/export-sample/02-system-requirements.html" target="_blank" rel="noopener"><strong>See a live HTML export →</strong></a> (from the reqmd project's own spec)

</details>

### `reqmd export graph` — export to LadybugDB graph

Creates a graph database with one node per requirement and `TracesTo` edges. Requires the `ladybug` build tag.

```sh
go build -tags ladybug -o reqmd ./cmd/reqmd
reqmd export graph spec/ -o graph-out/
```

<details>
<summary>Example output</summary>

```text
exporting spec/ → graph-out/reqmd-graph.lbug
  834 nodes, 1240 edges
  done
```

Query with the `lbug` CLI: `lbug graph-out/reqmd-graph.lbug`

</details>

### `reqmd baseline diff` — compare two git tags

Extracts the spec tree at each tag (no checkout), parses both, and produces a semantic diff.

```sh
reqmd baseline diff v1.0 v2.0
reqmd baseline diff v1.0 v2.0 --json
reqmd baseline diff HEAD~10 HEAD
```

Reports added, removed, and modified requirements (with attribute-level detail), schema changes, and submodule pin changes. Always exits 0.

<details>
<summary>Example output</summary>

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

</details>

### `reqmd repin` — update version pins

Updates `~N` version pins in trace references to match the upstream's current version. Dry-run by default; pass `--yes` to apply.

```sh
reqmd repin spec/                                    # dry-run: list proposed changes
reqmd repin spec/ --yes                              # apply changes
reqmd repin spec/ --yes --promote-unpinned           # also pin refs with no ~N
reqmd repin spec/ --json                             # machine-readable output
```

Rewrites only the `attr` blocks; surrounding Markdown is preserved. Idempotent — a second run after the first reports no changes. Predated pins (pin > upstream version) are surfaced but never auto-fixed.

<details>
<summary>Example output</summary>

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

</details>

| Flag | Description |
|------|-------------|
| `--yes`, `-y` | Apply changes without prompting |
| `--json` | Output as JSON |
| `--promote-unpinned` | Also pin trace refs that have no `~N` against a versioned upstream |
| `--dry-run` | Print the change list without applying (default; implicit when `--yes` is absent) |

## All commands at a glance

| Command | What it does | Key flags |
|---------|-------------|-----------|
| `check <dir>` | Validate + trace checks | `--json`, `--results`, `--relaxed-versions`, `-s` |
| `ls <dir>` | List all requirements | `--json` |
| `stats <dir>` | Stats per document | `--json` |
| `init <dir>` | Scaffold a new project | `--preset`, `--id-prefix`, `--force` |
| `serve <dir>` | Live HTML preview | `--addr`, `--results`, `--headless` |
| `export csv <dir>` | CSV export | `-o`, `--results` |
| `export html <dir>` | HTML export | `-o`, `--results` |
| `export graph <dir>` | Graph export (ladybug tag) | `-o` |
| `baseline diff <t1> <t2>` | Compare two git tags | `--json` |
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
| `level` | string | The V-model level for this document (e.g. `stakeholder-needs`, `system-requirements`, `software-requirements`, `test-specs`). Used for the document chain in HTML export. |
| `document-id` | string | A short identifier for this document directory (e.g. `system`, `stakeholder`, `tests`). Used in qualified references (`document-id/ID`). |
| `id-prefix` | string | The ID prefix for requirements in this directory (e.g. `SYS-`, `STK-GOAL-`, `TST-`). Requirements get IDs like `SYS-001`, `SYS-002`. |
| `upstream` | object | Declares which document directories are valid upstream sources for trace links. |
| `upstream.level` | string | The `x-reqmd.level` of the expected upstream. |
| `upstream.sources` | array | List of relative paths to upstream document directories. reqmd reads their `schema.yaml` to resolve trace targets. |
| `mandatory-disposition` | boolean | If `true`, every requirement in this directory must carry a `disposition` attribute (`implemented`, `deferred`, or `rejected`). Used for test specs and verification measures. |
| `external` | boolean | If `true`, requirements in this directory are external reference requirements (imported, not authored). They can be traced *to* but don't need upstream traces. |
| `additional-status-values` | array | Extend the built-in `status` enum (`[draft, approved]`) with custom values (e.g. `[review, withdrawn]`). Lowercase only, no built-in collision, no duplicates. |
| `ignore-status` | boolean | If `true`, all requirements in this directory are treated as coverage providers regardless of status. The status filter is hidden in HTML export. |

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

### Built-in attributes

reqmd recognizes these attribute names automatically (they don't need to be in your `properties`, but they can be):

| Attribute | Description |
|-----------|-------------|
| `trace` | Array of upstream IDs this requirement traces to. Supports `~N` version pins (e.g. `SYS-001~3`). |
| `requires-trace-from` | Declares that this requirement expects downstream coverage. Used for coverage checks. |
| `status` | `draft` or `approved`. Only `approved` counts as coverage. |
| `disposition` | `implemented`, `deferred`, or `rejected`. Required when `mandatory-disposition: true`. |
| `version` | Integer version of the requirement. Used by version-pin staleness checks. |
| `verify` | Verification method: `Test`, `Review`, `Inspection`, `Analysis`, or `Demonstration`. |
| `reqmd-suppress` | Array of check names to suppress (e.g. `[version-pin, missing-verdict]`). |
| `external` | If `true`, this is an external reference requirement (imported, not authored in this tree). |

## The requirement format

Each requirement is a level-2 heading, a YAML `attr` block in a fenced code block, and free-form prose:

<pre><code>## SYS-001  Temperature warning

&#96;&#96;&#96;attr
priority: High
status: approved
trace:
  - STK-003~2
verify: Test
version: 1
&#96;&#96;&#96;

The system shall display a warning when cabin temperature
exceeds 80°C for more than 5 seconds of continuous operation.

*Rationale:* Driver distraction from sudden thermal events
is a safety concern.</code></pre>

The heading text is the requirement ID plus a title. The `attr` block holds the structured attributes. The prose is the requirement body. The `Rationale:` paragraph is optional but shown in the HTML export.

## See also

- [Quickstart](/quickstart/) — a hands-on tour of every command.
- [FAQ](/faq/) — answers to common questions.
- [Use cases](/use-cases/) — what reqmd is for, and what it's not.