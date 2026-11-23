---
title: "Import an ALM or your code"
description: "Close the traceability gap: extract requirement proxies from source code, or pull requirements out of DOORS, codeBeamer, or any CSV/API — both through one shared exchange format."
weight: 5
---

Requirements rarely start (or end) inside reqmd. Two halves of the same
problem:

- **Code** — the functions, types, and methods that *implement* your spec need
  proxy requirements so traces resolve end to end.
- **ALM tools** — the DOORS or codeBeamer items that product managers review
  need to land in the spec tree as generated, reviewed-as-diffs documents.

Both are handled the same way: a small producer emits one JSON bundle in the
[exchange format](/exchange/v1/), and a single renderer (`reqmd-import`) writes
the actual reqmd documents. Nobody but the renderer writes `attr` blocks or
`schema.yaml`.

```
DOORS / codeBeamer / CSV ──▶ your script ──▶ exchange JSON ──┐
                                                                     ├─▶ reqmd-import ──▶ .md + schema.yaml ──▶ reqmd check
source code (tree-sitter) ─▶ reqmd-code ─────▶ exchange JSON ─┘
```

Why one renderer: a second Markdown writer in Python (or Java, or DXL) would
drift from the quoting, heading, and schema rules. The JSON format is the
stable contract; the renderer is the only code that knows the file syntax.

## Source code: `reqmd-code`

Walks your C, C++, Go, Python, or Rust code, parses each file with a
tree-sitter grammar, and emits every symbol as a proxy requirement. The source
path is recorded as provenance, the symbol kind is preserved, and the doc
comment becomes child section items under the symbol. Symbols are grouped into
one `## <path>` section per source file, and methods nest under their parent
type via `children`. The result is a spec tree where code and spec share an ID
space.

```sh
# Build (both halves)
go build -o reqmd-code ./reqmd-code/cmd/reqmd-code
go build -o reqmd-import ./reqmd-import/cmd/reqmd-import

# Extract code and render it into the spec tree
reqmd-code ./internal | reqmd-import - -o spec/03-software/imported

# Validate the whole tree, generated documents included
reqmd check spec/
```

Requirements are rendered one `.md` file per package (`package.md`), under an
`IMP-<package>-` ID prefix. Given `internal/cli/check_test.go`:

````markdown
# Package: cli

<!-- reqmd: generated -->

## internal/cli/check_test.go
### IMP-cli-BenchmarkRunValidationPipeline-80e6286: BenchmarkRunValidationPipeline
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/check_test.go
x-reqmd.source-line: 19
x-reqmd.symbol-kind: function
trace: []
```

BenchmarkRunValidationPipeline measures the cost of the full validation
pipeline against the spec/ fixture.
````

The ID format is `IMP-<package>-<symbol>-<7charhash>`. The hash is computed over
`(package, symbol)` so IDs survive line-number shifts — a refactor that moves
code around does not invalidate the trace.

### Options

```sh
reqmd-code ./internal                       # bundle to stdout
reqmd-code ./internal -o imported.json      # bundle to a file
reqmd-code --lang go ./internal             # restrict to one language
reqmd-code --heuristic-traces ./internal    # also mine bare IDs from prose
reqmd-code --id-prefix MY- ./src            # custom ID prefix (default IMP-)
```

`--heuristic-traces` looks for bare IDs in doc comments and prose (e.g.
`// implements SYS-001`) and records them under `x-reqmd.heuristic-trace`
— a separate attribute from the hand-asserted `trace:`, so a guess never
masquerades as a claim.

### Languages

| Language | Status | Tree-sitter grammar | Notes |
|---|---|---|---|
| **Go** | Stable | `tree-sitter-go` | Functions, methods, types, constants, variables. Method receivers are part of the ID (`Foo.Bar`). |
| **Python** | Stable | `tree-sitter-python` | Functions, classes, methods, constants. Package falls back to the module (file) name. |
| **C** | Stable | `tree-sitter-c` | Functions, struct/enum/union types, typedefs, global variables, `#define` macros. Function prototypes and function-local declarations are skipped. |
| **C++** | Stable | `tree-sitter-cpp` | Functions, class/struct member functions (inline and declared), class/struct/enum/union types, `using` aliases, typedefs, globals, macros. Method receivers carry the enclosing class. |
| **Rust** | Stable | `tree-sitter-rust` | Functions, `impl`/`trait` methods, struct/enum/trait/union types, type aliases, constants, statics. Method receivers carry the impl'd type or trait. |
| Zig | Not implemented | — | The registry slot is reserved; contributions welcome. |

To add a language: implement the `lang.Language` interface in a new package
under `reqmd-code/internal/lang/<lang>/` and self-register it via
`lang.Register(&impl{})` in an `init()` block. The blank import in
`cmd/reqmd-code/main.go` pulls all plugins into the binary.

## ALM tools and anything else: the exchange format

Any tool that can print JSON can feed reqmd. A producer writes one JSON bundle;
`reqmd-import` renders it into `schema.yaml` plus Markdown. The format is
neutral: nothing reqmd-specific is required of you, and the few reqmd-only
decisions (ID prefix, V-model level, upstream wiring) live in an optional
`x-reqmd` block.

A minimal bundle — one section grouping two requirements:

```json
{
  "$schema": "https://reqmd.dev/exchange/v1/schema.json",
  "documents": [
    {
      "key": "brakes",
      "title": "Brakes Module",
      "properties": {"priority": {"type": "string", "enum": ["High", "Low"]}},
      "x-reqmd": {"id-prefix": "DOORS-", "level": "stakeholder-requirements",
                  "external": true, "dir": "brakes"},
      "items": [
        {"kind": "section", "title": "Hydraulics",
         "children": [
           {"kind": "requirement", "key": "17", "title": "Brake request",
            "body": "The pedal shall request braking.", "fields": {"priority": "High"}},
           {"kind": "requirement", "key": "42", "title": "Brake pressure",
            "fields": {"priority": "Low"}, "links": ["17"], "refs": ["SYS-BRAKE-001"]}
         ]}
      ]
    }
  ]
}
```

```sh
reqmd-import brakes.json --validate-only      # report every violation, write nothing
reqmd-import brakes.json -o spec/00-doors/ --check
```

The format deliberately keeps the *source tool's* stable IDs:

- `kind` is `requirement` (content: an attr block, validated, traced, counted)
  or `section` (structure only: a heading, its prose, its children). A section
  that has children renders as a folder; one without renders as a prose block.
- `key` is the DOORS object number or codeBeamer item ID, unique per document.
  The reqmd ID becomes `<x-reqmd.id-prefix><key>` (`DOORS-17`), so re-imports
  never churn IDs and hand-written `trace:` links keep resolving.
- `children` nest items in document order. There is no parent pointer to get
  wrong: a missing parent, an out-of-order child and a cycle are unrepresentable.
- `links` name other items by key (`"17"`, or `"brakes/17"` across documents)
  and are rewritten to full requirement IDs; `refs` hold IDs owned elsewhere and
  are passed through verbatim.
- `fields` are the attr-block entries (sections never carry them). A JSON
  string stays a string even when it reads like a number, `null` or `true`, so
  send a real number or boolean when you mean one.
- `body` is Markdown and **must not contain headings** — structure belongs in
  sections. Doc comments with headings are split into sections by `reqmd-code`
  for exactly this reason.
- `status` is your own vocabulary. `--status-map "In Review=review"` translates
  it; an unmapped value falls back to reqmd's default and is kept as
  `x-reqmd.source-status`.
- `properties` / `required` describe the attr block's JSON Schema, which the
  renderer writes into the document's `schema.yaml`.
- `x-reqmd.external: true` marks the document a proxy for content authored
  elsewhere.

The renderer never normalizes, reorders or drops: anything that would corrupt a
document — a dangling link, a section with fields, an undeclared attribute, a
line break in a title, a nesting deeper than six heading levels — is rejected,
and every violation is reported in one run with an item path.

Full field reference, every rule and two complete examples:
[exchange format v1](/exchange/v1/) · machine-readable
[`schema.json`](/exchange/v1/schema.json).

## The trace gap

Plain Markdown traceability has a directional problem. A software requirement
can `trace: [system-requirements/SYS-001]` and reqmd validates that SYS-001
exists upstream. But the *implementation* — the function that actually
implements the requirement — has no native concept of a reqmd ID. Without
help you maintain a mapping table by hand, use an IDE plugin, or give up and
let the spec stop at the test layer.

`reqmd-code` produces that mapping automatically. It emits the symbols as an
exchange bundle; `reqmd-import` renders the bundle, and the IDs then exist in
the spec tree. The agent and the human can grep them, link them, and
CI enforces that they still resolve. When a function is renamed or moved, the
next run updates the file — the ID is a hash of `(package, symbol)`, not a
position, so the spec tree still points at the same requirement.

## Where it fits in the workflow

```sh
# 1. Author the spec by hand
reqmd init my-feature/
$EDITOR my-feature/stakeholder/goals.md
$EDITOR my-feature/system/features.md
reqmd check my-feature/                       # pass

# 2. Import the upstream requirements from the ALM (optional)
python my_doors_import.py -o doors.json
reqmd-import doors.json -o my-feature/00-doors/

# 3. Write the code, then extract proxies and re-validate
reqmd-code ./src | reqmd-import - -o my-feature/03-software/imported
reqmd check my-feature/                       # still pass

# 4. CI runs the same check on every PR
#    - run: reqmd check spec/
```

The agent — and the human — can then answer questions like:

- *What functions implement SYS-001?* — `grep -r "SYS-001" spec/03-software/imported/`
- *What spec IDs does `foo/bar.go` correspond to?* — `reqmd ls spec/ | grep "bar.go"`
- *What code is untested?* — load CTRF results with `reqmd check spec/ --results ctrf.json`; the `missing-verdict` check fires on measures with no test.

## What it doesn't do

- **Guess traceability.** Generated items start with `trace: []`. You annotate
  the source, run `--heuristic-traces`, declare links in your importer, or
  hand-fill the attr block.
- **Refresh on every commit.** You re-run the producer when you want new
  output. Incremental sync is a per-tool concern that would add drift for a
  workflow that runs once per PR or release.
- **Read binary artefacts.** `reqmd-code` reads source, not compiled binaries
  or runtime behaviour; add a runtime instrumentation step if you need that.
- **Sync content back into the ALM.** See the authority rule above.

## Architecture

```
  walkdir (per-file)                    JSON bundle (exchange v1)
    → language plugin (tree-sitter)            │
      → symbols                                ▼
        → trace inference                  reqmd-import
          → JSON bundle  ────────────────▶  ├─ schema.yaml per document
                                           ├─ .md per document (hierarchical headings)
                                           └─ idempotent write-if-changed, marker-guarded
```

Each stage is independently testable. The renderer's idempotency guarantee
(byte-identical output for unchanged input, hand-authored files untouched) is
what makes both import paths safe to re-run in CI. The tree-sitter grammars are
loaded once at startup and per-file parsing is dispatched to a worker pool
sized to `runtime.NumCPU()`.

## See also

- [Quickstart](/quickstart/) — including step 15, a custom CSV importer.
- [Use cases](/use-cases/) — where the importers sit in the engineering workflow.
- [Compare](/compare/) — how reqmd compares to other tools.
- Sources: [`reqmd-code`](https://github.com/dVoo/reqmd/tree/main/reqmd-code),
  [`reqmd-import`](https://github.com/dVoo/reqmd/tree/main/reqmd-import),
  [`python/`](https://github.com/dVoo/reqmd/tree/main/python).
