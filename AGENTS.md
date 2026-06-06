# ReqMD — AGENTS.md

## What this repo is

Spec + Go CLI for **reqmd**, a tool that validates and exports requirement specs stored in
Markdown files with embedded `attr` blocks (YAML) validated against JSON Schema 2020-12 (YAML).

## Layout

- `SPEC.md` — authoritative design spec
- `workspace.dsl` — C4 model (Structurizr DSL) for architecture visualization
- `spec/00-aspice/`, `spec/01-stakeholder/`, `spec/01a-aspice-stakeholder/`, `spec/02-system/`, `spec/03-software/`, `spec/04-tests/` — 6 doc dirs, 232 total reqs, V-model dogfood fixture
- `quickstart/` — three-document tutorial tree (stakeholder → system → software [+ tests])
- `internal/` — Go packages (model, parser, schema, exporter, reporter, graph, cli)
- `cmd/reqmd/main.go` — entry point for the `reqmd` binary (cobra subcommands live in `internal/cli/`)
- `go.mod` / `go.sum` — Go module (1.25)
- `.opencode/` — OpenCode tooling install (not part of the project)

## Key facts

- **Binary name is `reqmd`**, not `mdreq` (the repo name).
- **No `opencode.json`, no Makefile, no CI** yet. Adding any is greenfield.
- The Python `validate.py` in `spec/example/` is a **legacy prototype** — the Go CLI replaces it. Now removed.

## Build and run

```sh
go build -o reqmd ./cmd/reqmd          # build
go run ./cmd/reqmd check <root>       # check all docs under root
go run ./cmd/reqmd ls <root>          # list all requirements
go run ./cmd/reqmd stats <root>       # stats breakdown per doc
go run ./cmd/reqmd export csv <root>  # CSV export
go run ./cmd/reqmd export html <root> # HTML export
go run ./cmd/reqmd check --json <root>   # JSON output
go run ./cmd/reqmd check <file> -s <schema>  # single-file mode
```

## Commands (implemented)

| Command | Description |
|---------|-------------|
| `reqmd check <root>` | Recursive walk for `schema.yaml`, check all `.md` |
| `reqmd ls <root>` | Table of all requirements with all schema attributes |
| `reqmd stats <root>` | Attribute-value breakdown per document directory |
| `reqmd export csv <root> [-o <dir>]` | CSV export (`<dirname>-requirements.csv`) |
| `reqmd export html <root> [-o <dir>]` | HTML export with water.css CDN |
| `reqmd check --json <root>` | JSON validation report with requirements, pass/fail, trace checks |
| `reqmd ls --json <root>` | JSON list of requirement IDs with all attributes |
| `reqmd stats --json <root>` | JSON attribute-value breakdown per document |

Exit codes: 0 (all valid), 1 (validation errors), 2 (parse error).

Exit codes: 0 (all valid), 1 (validation errors), 2 (parse error).

## Architecture

```
cmd/reqmd/main.go → internal/cli (cobra commands)
  → parser.Discover(root) → []model.Document
    → goldmark AST parses each .md file
  → schema.Compile(schema, path) → validation against JSON Schema 2020-12
  → reporter.Report → stdout output + exit code
  → exporter.{CSV,HTML} → file output
```

- **Parser**: Recursive tree walk. Every dir with `schema.yaml` is a document dir.
  Files parsed in parallel via worker pool (`runtime.NumCPU()` goroutines).
- **Schema**: YAML → map → JSON marshal → `jsonschema.Schema.UnmarshalJSON` → `Resolve()`.
  Base URI set to `file://<absolute-path>/schema.yaml`.
- **Properties order**: Required fields first (in schema's `required` order),
  then optional fields (in schema definition order).
- **Trace ref checking**: Pass 2 graph-based trace checks (`graph.CheckResults()`) run after
  Pass 1 validation. Every value in the `trace` attribute is checked against all requirement IDs
  across all documents. Missing refs appear as WARNING-level results, cycles appear as ERROR-level.
  WARNING/INFO do not affect exit code; ERROR does.
- **Status lifecycle (built-in)**: `status` is a built-in enum `[draft, approved]`.
  Missing status defaults to `approved` (silent). Only `approved` requirements count
  as upstream coverage providers — a `draft` requirement can be referenced by a
  `trace` but does NOT satisfy a `requires-trace-from:` expectation. Extensions via
  `x-reqmd.additional-status-values: [review, ...]` (lowercase-only, no
  built-in collision, no duplicates). Per-document opt-out via
  `x-reqmd.ignore-status: true` (all requirements become coverage providers;
  status filter hidden in HTML export). The `checkNeedsCoverage` gate adds
  `(N draft downstreams ignored)` to the missing-coverage message when
  at least one inbound was filtered by the status check.
- **Version pins**: Trace refs may include a `~N` version pin (e.g. `UP-001~3`).
  When the upstream's `version` differs from the pin, `checkVersionPins()`
  emits a `version-pin` finding: `direction: "outdated"` (upstream newer than
  pin) is ERROR by default but can be demoted to WARNING via the
  `--relaxed-versions` CLI flag; `direction: "predated"` (pin ahead of upstream)
  is always ERROR. Unpinned refs and external/unversioned upstreams skip the
  check. Suppression: `reqmd-suppress: [version-pin]`.

## Testing

```sh
go test ./internal/...          # all unit tests (132 tests across 6 packages)
go test -v ./internal/parser/   # parser tests (most complex — 11 tests)
go test -v ./internal/schema/   # schema tests (Compile, Validate, Properties — 11 tests)
go test ./internal/reporter/    # reporter tests (ExitCode, Format, FormatList, FormatStats, Warnings, FormatJSON, FormatListJSON, FormatStatsJSON — 18 tests)
go test ./internal/exporter/    # exporter tests (CSV + HTML format — 5 tests)
go test ./internal/model/       # model tests (struct construction — 4 tests)
```

Tests use inline fixtures (no external files). No integration prerequisites.
Parser tests use `t.TempDir()` for file-based test cases.

## Test fixture

```sh
go run ./cmd/reqmd check spec/example/
# Expects: 3 requirements, IVI-FUN-003 missing required "asil"
# Exit code 1
```

## Go dependencies

| Package | Purpose |
|---------|---------|
| `github.com/spf13/cobra` | CLI framework |
| `github.com/yuin/goldmark` | Markdown AST parsing |
| `gopkg.in/yaml.v3` | YAML parsing |
| `github.com/google/jsonschema-go/jsonschema` | JSON Schema 2020-12 validation |

## Important gotchas

- Schema files are named `schema.yaml` (the spec previously referenced `doc.schema.yaml`).
- The `$id` field in schema files is a relative URI. It resolves against the schema's
  `file://` base URI set automatically by the compiler.

# Development rules

## Backward compatiblity

Do not care about backward compatibility but on writing clean code. Do not create code branches to be able to read old content versions.
Breaking changes are not a problem.

## Coding standards

Always comply to common best-practices and coding standards whenever possible

- DRY
- Clean Code
- KISS

## Test coverage

Whenever larger functionality is added, make sure to add unit-tests for
testing the functionality.

## Keep documentation up to date

After adding functionality:

- Check and update README.md
- Make sure that the specification (spec/reqs) is up to date
- Make sure that the software architecture (spec/workspace.dsl) is up to date
