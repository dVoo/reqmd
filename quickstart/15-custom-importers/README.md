# Step 15 — Custom Importers

Import requirements from another tool (DOORS, codeBeamer, a CSV export, a
database) into a validated reqmd spec tree.

## The idea

Nobody should have to hand-write `attr` blocks, `schema.yaml`, or heading
levels. Instead, every producer emits **one JSON exchange format**
(`reqmd-import/v1`), and a single renderer turns it into reqmd documents:

```
DOORS / codeBeamer / CSV ──▶ your script ──▶ reqmd-import/v1 JSON ──▶ reqmd-import ──▶ .md + schema.yaml
source code (tree-sitter) ─▶ reqmd-code ──▶ reqmd-import/v1 JSON ──┘
```

The format is the extension point. There is only one renderer (Go), so a
Python importer cannot produce invalid Markdown or drift from the schema rules.

## 1. The built-in code importer

```sh
# reqmd-code extracts code symbols, reqmd-import renders them as reqmd docs
reqmd-code ./internal | reqmd-import - -o spec/03-software/imported
reqmd check spec/
```

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
  --dir brakes --id-prefix DOORS- --title "Brakes Module" \
  --level stakeholder-requirements --external \
  --id-column "Absolute Number" --title-column "Object Heading" \
  --body-column "Object Text" --parent-column Parent \
  --trace-column Links --attr priority=Priority

# exchange bundle ──▶ reqmd documents, then validate
./reqmd-import /tmp/brakes.json -o /tmp/doors-spec --check
```

`--check` runs `reqmd check` on the output for you (it needs `reqmd` on
`PATH`). Without it, run `reqmd check /tmp/doors-spec` yourself.

The result keeps the DOORS hierarchy (`##`, `###`, `####`) and its stable
object numbers as reqmd IDs (`DOORS-17`), so re-imports never churn IDs and
hand-written `trace:` links keep resolving:

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

Use the `reqmd_io` helper ([`python/`](../../python)) — it only builds the
exchange JSON, no YAML or Markdown:

```python
from reqmd_io import Bundle, Document, Requirement

doc = Document(id_prefix="CB-", dir="codebeamer", level="stakeholder-requirements",
               title="codeBeamer Requirements", external=True,
               properties={"priority": {"type": "string"}})

for item in my_codebeamer_pull():
    doc.add(Requirement(native_id=item.id, title=item.name, body=item.description,
                        parent=item.parent_id,
                        native_trace=[link.target for link in item.links],
                        attrs={"priority": item.priority}))

Bundle([doc]).write("codebeamer.json")
```

```sh
reqmd-import codebeamer.json -o spec/00-codebeamer/
```

Key rules of the format:

- `native_id` is the source tool's own stable ID. The reqmd ID becomes
  `<id_prefix><native_id>`. It must not contain whitespace — a key like
  `Item 42` would produce an ID nothing could `trace:` to, so it is rejected.
- `trace` holds fully-qualified reqmd IDs; `native_trace` holds native IDs of
  other items in the bundle (they are rewritten to full IDs for you).
- `attrs` take strings, numbers, booleans, or flat lists. A string stays a
  string even when it looks like a number, so pass a real number when you mean
  one. Names may contain only letters, digits, `_`, `-` and `.`. `status` and
  `trace` are reserved.
- `body` is Markdown and may contain headings (they become info items), but not
  a fence whose info string is `attr` directly under a heading — that pair would
  become a requirement of its own.
- `properties` declares the attr-block JSON Schema, so the generated
  `schema.yaml` accepts exactly the attributes you emit.
- `parent` nests by native ID, to any depth (each level adds one `#`).

## Validate

```sh
# without a reqmd binary on PATH:
./reqmd-import /tmp/brakes.json -o /tmp/doors-spec
./reqmd check /tmp/doors-spec
```

Expected: 3 valid requirements. The trace to `SYS-BRAKE-001` is reported as a
broken reference — correct, because that upstream requirement lives in another
document and is not in this scratch tree.
