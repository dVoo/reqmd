# ReqMD — AGENTS.md

## What this repo is

Spec + Go CLI for **reqmd**, a tool that validates and exports requirement specs stored in
Markdown files with embedded `attr` blocks (YAML) validated against JSON Schema 2020-12 (YAML).

## Layout

- `SPEC.md` — authoritative design spec (if absent, the spec tree under `spec/` is the source of truth)
- `spec/workspace.dsl` — C4 model (Structurizr DSL) for architecture visualization
- `spec/00-aspice/`, `spec/01-stakeholder/`, `spec/01a-aspice-stakeholder/`, `spec/02-system/`, `spec/03-software/`, `spec/04-tests/` — 6 doc dirs, 233 total reqs, V-model dogfood fixture
- `quickstart/` — step-by-step tutorial (01-get-started, 01a-ci-integration, 02-trace-your-spec, 02a-status-disposition, 02b-version-pins, 03-export, 03a-live-preview, 03b-baseline-diff, 04-custom-templates, 05-verification-results-ctrf, 06-review-documentation)
- `internal/` — Go packages (model, parser, schema, exporter, reporter, graph, diff, cli)
- `cmd/reqmd/main.go` — entry point for the `reqmd` binary (cobra subcommands live in `internal/cli/`)
- `go.mod` / `go.sum` — Go module `reqmd` (1.25)
- `go.work` / `go.work.sum` — Go workspace linking `reqmd` (root) and `reqmd-import` so `go build ./...` and `go test ./...` from the repo root cover both modules. Both keep independent `go.mod` files and dependency sets.
- `reqmd-import/` — separate Go module (`reqmd-import`) for the extraction tool that imports source code as proxy requirements. Independent toolchain (tree-sitter); shares the spec *format* with reqmd but no Go code. Build with `go build ./reqmd-import/cmd/reqmd-import`. Modeled in `spec/workspace.dsl` as the "Extraction Tool" softwareSystem.
- `.opencode/` — OpenCode tooling install (not part of the project)

## Key facts

- **Binary name is `reqmd`**, not `mdreq` (the repo name).
- **No `opencode.json`, no Makefile, no CI** yet. Adding any is greenfield.
- The Python `validate.py` in `spec/example/` is a **legacy prototype** — the Go CLI replaces it. Now removed.

## Build and run

```sh
go build -o reqmd ./cmd/reqmd          # build reqmd CLI
go build -o reqmd-import ./reqmd-import/cmd/reqmd-import  # build extraction CLI
go run ./cmd/reqmd check <root>       # check all docs under root
go run ./cmd/reqmd ls <root>          # list all requirements
go run ./cmd/reqmd stats <root>       # stats breakdown per doc
go run ./cmd/reqmd export csv <root>  # CSV export
go run ./cmd/reqmd export html <root> # HTML export
go run ./cmd/reqmd check --results <path> <root> # check with ephemeral V&V results
go run ./cmd/reqmd export html <root> --results <path> -o docs/ # HTML with verdict badges
go run ./cmd/reqmd init <dir> --preset results    # scaffold manual results dir
go run ./cmd/reqmd init <dir> --preset ./my-preset/ # scaffold with custom preset
go run ./cmd/reqmd check --json <root>   # JSON output
go run ./cmd/reqmd check <file> -s <schema>  # single-file mode
go run ./cmd/reqmd repin <root>           # dry-run: list version-pin changes
go run ./cmd/reqmd repin <root> --yes     # apply version-pin changes in place
go run ./cmd/reqmd repin <root> --json    # machine-readable change list
```

## Commands (implemented)

| Command | Description |
|---------|-------------|
| `reqmd check <root>` | Recursive walk for `schema.yaml`, check all `.md` |
| `reqmd ls <root>` | Table of all requirements with all schema attributes |
| `reqmd stats <root>` | Attribute-value breakdown per document directory |
| `reqmd export csv <root> [-o <dir>] [--results <path>...]` | CSV export (`<dirname>-requirements.csv`). `--results` adds `Verdict` and `Verdict Source` columns from ephemeral verification results. |
| `reqmd export html <root> [-o <dir>] [--results <path>...]` | HTML export with water.css CDN. `--results` renders color-coded verdict badges (pass/fail/skipped/inconclusive) on measure cards. |
| `reqmd export graph <root> [-o <dir>] [--results <path>...]` | LadybugDB graph export (requires `-tags ladybug`). `--results` includes `RESULT:` nodes with `outcome`/`source` properties for graph traversal from requirements to verification results. |
| `reqmd check --json <root>` | JSON validation report with requirements, pass/fail, trace checks |
| `reqmd check --results <path> <root>` | Load ephemeral verification results (CTRF `.ctrf.json` or manual-results dirs with `schema.yaml`) and run outcome-gated checks (`missing-verdict`, `failing-verdict`). `--results` is repeatable; auto-detects CTRF vs manual by extension + shape. |
| `reqmd ls --json <root>` | JSON list of requirement IDs with all attributes |
| `reqmd repin <root> [-y/--yes] [--json] [--promote-unpinned]` | Update version-pin (~N) trace refs to the upstream's current version. Dry-run by default; `--yes` applies. `--promote-unpinned` also pins refs that have no ~N suffix against a versioned upstream. Predated findings (pin > upstream) are surfaced but never auto-fixed. |
| `reqmd stats --json <root>` | JSON attribute-value breakdown per document |
| `reqmd baseline diff <tag1> <tag2>` | Compare requirements between two git tags; also reports added/removed/updated submodules (flags: `--json`) |
| `reqmd init <dir>` | Scaffold a new requirements directory. Presets: `generic` (default), `aspice`, `results` (manual verification results), or a custom preset directory path. Flags: `--preset`, `--id-prefix`, `--id`, `--title`, `--level`, `--force` |

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
- **Repin (`reqmd repin`)**: Proposes/applies `~N` version-pin updates against upstream
  versions. Pure-`graph.RepinDeltas` computes a sorted list of `RepinDelta`
  (kind: `outdated` | `unpinned` | `predated`); `internal/repin.Apply` rewrites the
  source files in place, scoped to ```attr blocks so surrounding Markdown prose is
  preserved verbatim. The Pass 2 `OutboundPins` cache is augmented with
  `OutboundRefs` (the full source-form ref) so deltas round-trip without re-parsing
  YAML. Predated findings are surfaced but never auto-fixed — they are data-
  integrity errors. `--promote-unpinned` adds `unpinned` deltas for refs without
  any `~N` against a versioned upstream (opt-in because it converts "no claim" into
  "claimed at current version"). Default behavior is dry-run; `--yes` applies.
- **Verification results (`reqmd check --results`)**: Ephemeral verification
  results close the right side of the V-model. Results are loaded per
  invocation (never persisted in the spec repo) from CTRF JSON reports
  (automated tests) or manual-results markdown dirs (review/inspection/
  analysis). `internal/verify` parses both into a `measureID → latestResult`
  map (latest by CTRF `tests[].stop` or manual `verified-at`), synthesizes
  one pseudo-requirement per result (`RESULT:<id>`, `trace: [MEASURE-ID~N]`,
  `outcome: pass|fail|skipped|inconclusive`), and appends them to the doc
  slice before `graph.New`. Two new graph checks fire when result nodes are
  present: `missing-verdict` (approved measure with no result, WARNING) and
  `failing-verdict` (latest result is fail, ERROR). The existing
  `version-pin` check applies to result→measure traces via `~N` pins,
  reusing stale-verdict detection with no new logic. CTRF→measure mapping
  via the `x-reqmd.id` extra field per test. Suppression:
  `reqmd-suppress: [missing-verdict]`, `[failing-verdict]`.
- **Baseline diff (`reqmd baseline diff`)**: Loads the spec tree at two git tags
  via `git archive | tar` (no checkout), parses both with the standard pipeline,
  and produces a semantic diff of requirements (added/removed/modified with
  attribute-level detail) and schemas (new/removed properties, changed required
  fields). Output: colored text by default, `--json` for structured. Uses
  `github.com/r3labs/diff/v3` for structured diffing.
- **Submodule diff**: `git ls-tree -t <tag>` extracts submodule commit pins. The
  diff reports added, removed, and updated submodules (same-commit skipped).
  No new dependencies; pure stdlib `os/exec`. Output: short hashes in text,
  full SHA in JSON. Section is hidden when no submodules exist.



```sh
go test ./...                         # all tests, both modules (via go.work)
go test ./internal/...                # reqmd unit tests (10 packages)
go test ./reqmd-import/...            # reqmd-import unit tests
go test -v ./internal/parser/   # parser tests (most complex)
go test -v ./internal/schema/   # schema tests (Compile, Validate, Properties)
go test ./internal/reporter/    # reporter tests (ExitCode, Format, FormatList, FormatStats, Warnings, FormatJSON, FormatListJSON, FormatStatsJSON)
go test ./internal/exporter/    # exporter tests (CSV + HTML format)
go test ./internal/model/       # model tests (struct construction)
go test ./internal/diff/        # baseline diff tests
go test ./internal/verify/      # CTRF parser, manual-results loader, merge, synth
go test ./internal/repin/        # repin tests (Build, Apply, format, whole-ref match safety)
```

Tests use inline fixtures (no external files). No integration prerequisites.
Parser tests use `t.TempDir()` for file-based test cases.

## Go dependencies

| Package | Purpose |
|---------|---------|
| `github.com/spf13/cobra` | CLI framework |
| `github.com/yuin/goldmark` | Markdown AST parsing |
| `gopkg.in/yaml.v3` | YAML parsing |
| `github.com/google/jsonschema-go/jsonschema` | JSON Schema 2020-12 validation |
| `github.com/r3labs/diff/v3` | Structured diffing for baseline comparison |
| `github.com/mattn/go-isatty` | TTY detection for interactive `repin` prompt |
| `github.com/fsnotify/fsnotify` | Filesystem watching for `serve` live-reload |
| `golang.org/x/sync` | Parallel document loading via `errgroup` |

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

## Version control

This project uses Jujutsu (jj) as its own VCS. For each new feature create a new changeset.
Do not use Git for the tool's own history.

Note: the `baseline diff` command and `internal/parser/git.go` shell out to `git` to
read document trees (the spec repos reqmd operates on), which are typically git-backed.
That is expected: jj for the tool, git for the document trees the tool validates/diffs.

## Keep documentation up to date

After adding functionality:

- Check and update README.md
- Make sure that the specification (the `spec/` tree) is up to date
- Make sure that the software architecture (spec/workspace.dsl) is up to date
