---
title: "Step 15 — Custom importers"
description: "Import requirements from DOORS, codeBeamer, or any CSV/API — and extract proxy requirements from source code — through one shared exchange format."
weight: 15
---

Import requirements from another tool (DOORS, codeBeamer, a CSV export, a
database) into a validated reqmd spec tree.

## The idea

Nobody should have to hand-write `attr` blocks, `schema.yaml`, or heading
levels. Instead, every producer emits **one JSON exchange format**, and a single
renderer turns it into reqmd documents:

```text
DOORS / codeBeamer / CSV ──▶ your script ──▶ exchange bundle ──▶ reqmd-import ──▶ .md + schema.yaml
source code (tree-sitter) ─▶ reqmd-code ────────────────────────┘
```

The format is the extension point. There is only one renderer (Go), so a Python
importer cannot produce invalid Markdown or drift from the schema rules.

The full contract — every field, every rule the renderer enforces, and two
complete examples — lives on the [exchange format page](/exchange/v1/). The
machine-readable form is
[`schema.json`](/exchange/v1/schema.json); load it into your producer's tests.

## 1. The built-in code importer

```sh
# reqmd-code extracts code symbols, reqmd-import renders them as reqmd docs
reqmd-code ./internal --lang go | reqmd-import - -o spec/03-software/imported
reqmd check spec/
```

Each package becomes a document grouped into one section per source file, with
doc comments rendered as prose sections under their symbol.

## 2. A CSV importer (DOORS / codeBeamer)

`brakes.csv` in this folder is a typical DOORS export. The repo ships a generic
reader for both tools — it maps columns to fields, so no code changes per tool:

```sh
# build the renderer and the importer
go build -o reqmd-import ./reqmd-import/cmd/reqmd-import
go build -o reqmd ./cmd/reqmd

# CSV ──▶ exchange bundle
python python/examples/csv_import.py quickstart/15-custom-importers/brakes.csv \
  -o /tmp/brakes.json \
  --key brakes --dir brakes --id-prefix DOORS- --title "Brakes Module" \
  --level stakeholder-requirements --external \
  --id-column "Absolute Number" --title-column "Object Heading" \
  --body-column "Object Text" --parent-column Parent \
  --trace-column Links --attr priority=Priority \
  --kind-column "Object Type"

# check the bundle without writing anything, then render
./reqmd-import /tmp/brakes.json --validate-only
./reqmd-import /tmp/brakes.json -o /tmp/doors-spec --check
```

`--validate-only` reports every violation in one run and writes nothing — the
fastest loop when you are writing your own producer. `--check` runs `reqmd
check` on the output afterwards (it needs `reqmd` on `PATH`).

`--kind-column "Object Type"` turns DOORS' `Heading` and `Text` rows into
sections: a heading groups the requirements under it, a text row is a leaf prose
block. Both are the same `kind` in the format, because folder-versus-leaf is
derived from whether a section has children.

The result keeps the DOORS hierarchy and its stable object numbers as reqmd IDs
(`DOORS-17`), so re-imports never churn IDs and hand-written `trace:` links keep
resolving:

````markdown
## DOORS-1: Brakes module
```attr
status: approved
priority: High
trace: []
```

The vehicle shall decelerate on demand.
````

## 3. Writing your own importer

Use the `reqmd_io` helper ([`python/` in the repo](https://github.com/dVoo/reqmd/tree/main/python)) — it only builds the
exchange JSON, no YAML or Markdown:

```python
from reqmd_io import Bundle, Document, Requirement, Section

doc = Document(
    key="codebeamer",
    title="codeBeamer Requirements",
    x_reqmd={"id-prefix": "CB-", "level": "stakeholder-requirements", "external": True},
    properties={"priority": {"type": "string"}},
)

doc.add(Section(title="Brakes", body="Braking subsystem."))
for item in my_codebeamer_pull():
    doc.add(
        Requirement(
            key=item.id,
            title=item.name,
            body=item.description,
            fields={"priority": item.priority},
            links=[link.target_key for link in item.links],
            refs=[link.external_id for link in item.external_links],
        ),
        parent=item.parent_id,
    )

Bundle([doc]).write("codebeamer.json")
```

```sh
reqmd-import codebeamer.json -o spec/00-codebeamer/
```

Key rules of the format:

- **`kind`** is `"requirement"` (content: gets an `attr` block, is validated,
  traced and counted) or `"section"` (structure only). It is mandatory.
- **`key`** is the source tool's own stable ID, unique within the document, and
  required on requirements. The reqmd ID becomes `<x-reqmd.id-prefix><key>`.
  Re-imports never churn IDs.
- **`links`** name other items by key — `"17"` in the same document,
  `"brakes/17"` across documents. They must resolve, and may only target
  requirements. **`refs`** hold requirement IDs owned elsewhere
  (`SYS-BRAKE-001`) and are passed through verbatim.
- **`children`** nest items, in document order. A child cannot reference a
  missing parent, cannot come before it, and cannot form a cycle — the shape
  makes it unrepresentable.
- **`fields`** are the attr-block attributes (requirements only; a section with
  `fields` is rejected). Values are scalars or flat lists. Names may contain
  only letters, digits, `_`, `-` and `.`.
- **`body`** is Markdown and must not contain headings. Structure belongs in
  sections, where it can be nested and navigated.
- **`properties`** / **`required`** declare the attr-block JSON Schema, so the
  generated `schema.yaml` accepts exactly the attributes you emit.
- **`x-reqmd`** holds everything reqmd-specific (`id-prefix`, `level`,
  `document-id`, `upstream`, `requires-trace-from`, `external`, `dir`, `file`).
  A producer that is not reqmd-aware omits it, and requirement IDs become the
  bare keys.

The renderer rejects a bundle it cannot render faithfully — an undeclared
attribute, a dangling link, a section with fields, a title with a line break, a
nesting deeper than six heading levels — and reports **all** violations at once,
each addressed by item path (`brakes/Hydraulics/17`). It never silently
normalizes, reorders or drops anything.

## Validate

```sh
# without a reqmd binary on PATH:
./reqmd-import /tmp/brakes.json -o /tmp/doors-spec
./reqmd check /tmp/doors-spec
```

Expected: 3 valid requirements. The trace to `SYS-BRAKE-001` is reported as a
broken reference — correct, because that upstream requirement lives in another
document and is not in this scratch tree.
