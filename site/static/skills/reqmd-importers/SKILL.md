---
name: reqmd-importers
description: Write a reqmd importer — pull requirements out of DOORS, codeBeamer, a CSV, an API or a database into reqmd by emitting the exchange format (JSON) and letting reqmd-import render it into schema.yaml + Markdown.
---

# Writing a reqmd importer

## When to use

You have requirements in another system (DOORS, codeBeamer, Polarion, JAMA, a
CSV export, a REST API, a database) and want them validated, traceable and
reviewable in reqmd. Also relevant when a producer emits a bundle that
`reqmd-import` rejects and you need to know why.

Not for this: reading reqmd (that is `reqmd ls`/`export`), or writing Markdown
by hand (that is the authoring quickstart).

## Ground rules

1. **Your importer emits JSON, never Markdown or YAML.** `reqmd-import` is the
   only writer, so quoting, heading depth and schema rules exist once.
2. **Keep the source tool's own stable ID as `key`.** The rendered reqmd ID is
   `<x-reqmd.id-prefix><key>` (`DOORS-17`). Re-imports must never churn IDs, or
   every hand-written `trace:` breaks.
3. **Never invent a reqmd ID.** Supply `key`; the renderer stamps the ID.
4. **Structure is `kind: "section"`, content is `kind: "requirement"`.** A
   section has no fields, is never traced to and is never counted as coverage.
5. **Validate before you render.** `reqmd-import bundle.json --validate-only`
   reports every violation in one run and writes nothing.
6. The authoritative contract is `/exchange/v1/` (source:
   `site/content/exchange/v1.md`) plus `schema.json` next to it. Read it before
   inventing a field.

## Procedure

### 1. Decide the boundary

One reqmd document per module/tracker/spec, not one per item. Documents are
cheap; cross-document links need a `document-id`.

### 2. Choose the document shape

```python
doc = Document(
    key="brakes",                      # bundle-internal identity, unique, no "/"
    title="Brakes Module",             # required
    x_reqmd={                          # everything reqmd-specific
        "id-prefix": "DOORS-",
        "level": "stakeholder-requirements",
        "document-id": "brakes",       # how trace refs qualify this document
        "external": True,              # a mirror of content authored elsewhere
        "dir": "brakes",               # output directory; omit to use the key
    },
    properties={"priority": {"type": "string"}},   # reqmd-block JSON Schema
    required=["priority"],
)
```

### 3. Fetch and map items

```python
from reqmd_io import Bundle, Document, Requirement, Section

for obj in doors_module.objects:
    if obj.type == "Heading":
        doc.add(Section(key=obj.number, title=obj.heading,
                        body=obj.text or ""))
        continue
    doc.add(
        Requirement(
            key=obj.absolute_number,
            title=obj.heading,
            body=obj.text,                   # Markdown, no headings
            fields={"priority": obj.priority},
            links=obj.internal_link_keys,    # other items in this bundle
            refs=obj.external_link_ids,      # IDs owned elsewhere, verbatim
        ),
        parent=obj.parent_number or "",      # breadth-first is fine
    )
```

`doc.add(item, parent=key)` nests by key, so a flat table stays a flat loop.
The parent must already have been added — add parents before children, or
collect rows and nest in a second pass.

### 4. `links` versus `refs`

- `links` — keys of other items **in this bundle**. The renderer resolves and
  rewrites them; an unresolved link is an error. Use `"17"` for the same
  document and `"brakes/17"` across documents. They must target requirements.
- `refs` — requirement IDs owned by **someone else** (`SYS-BRAKE-001`). Written
  verbatim into `trace:`. Never resolved by the importer; `reqmd check` reports
  a missing target, which is the correct signal for an external document.

Both end up in the rendered `trace:` list.

### 5. Declare the attr schema

Every key you put in `fields` must appear in the document's `properties`, or the
renderer rejects the requirement. Values are strings, numbers, booleans or flat
lists of those. A JSON string stays a string even when it reads like `123`,
`null` or `true` — send a real number when you mean one.

`required` means "every requirement in this document carries it". Nothing is
inherited from a section: repeat the attribute on each requirement.

### 6. Render and validate

```sh
reqmd-import bundle.json --validate-only     # validate only
reqmd-import bundle.json -o spec/00-doors/   # render
reqmd-import bundle.json -o spec/00-doors/ --check   # render + reqmd check
```

Re-running is a no-op when nothing changed; files without the
`<!-- reqmd: generated -->` marker are never touched.

## Format reference

Document:

| Field | Required | Meaning |
| --- | --- | --- |
| `key` | yes | Document identity, unique in the bundle, no `/` or whitespace |
| `title` | yes | Level-1 heading |
| `body` | no | Preamble prose under the title |
| `items` | no | Nested items, in document order |
| `properties` / `required` | no | Attr-block JSON Schema (requirements only) |
| `x-reqmd` | no | `id-prefix`, `level`, `document-id`, `upstream`, `requires-trace-from`, `external`, `dir`, `file` |

Item:

| Field | Required | Meaning |
| --- | --- | --- |
| `kind` | yes | `"requirement"` or `"section"` |
| `key` | requirements | Stable source ID, unique per document, no whitespace |
| `title` | no | Heading text; omit for a prose block |
| `body` | no | Markdown, **no headings** |
| `fields` | no | Attr-block entries (rejected on sections) |
| `status` | no | Source vocabulary; map with `--status-map` |
| `links` | no | Keys of other items, must resolve to requirements |
| `refs` | no | External requirement IDs, verbatim |
| `children` | no | Nested items, document order |

Bundle: `{"$schema": "https://reqmd.dev/exchange/v1/schema.json", "documents": [...]}`.
Unknown fields are rejected everywhere.

## Rejections you will hit

All are reported together, addressed by item path (`brakes/Hydraulics/17`):

- `kind` missing or unknown; a requirement without `key`; duplicate keys
- `fields` on a section; an attribute not declared in `properties`; an
  unsatisfied `required`; a nested object/list as an attribute value; an
  attribute name outside `[A-Za-z0-9._-]`
- a `link` that does not resolve, or resolves to a section
- a title containing a line break; a document `key` containing `/`
- nesting deeper than six heading levels (flatten in the producer)
- two documents claiming the same output path

`body` containing a heading is a **producer obligation, not checked**: it will
silently re-parent the tree. Split structure into sections instead.

## Pitfalls

- **Reusing keys across source modules.** Keys are document-scoped, so `17` in
  two documents is fine — but a cross-document `link` must say `doc-key/17`.
- **`body` with Markdown headings.** The most common way to corrupt a tree.
- **Numbers typed as strings.** `fields={"line": "12"}` renders as `"12"`; if
  the schema says integer, pass `12`.
- **Sections with fields.** A source heading that carries its own metadata has
  no representation: put the metadata on the requirements beneath it.
- **Assuming the reqmd ID.** Always `<id-prefix><key>`; never hand-mint one.

## Worked example

`python/examples/csv_import.py` is a complete DOORS/codeBeamer reader; see
quickstart step 15 for the invocation, including `--kind-column "Object Type"`
to turn `Heading`/`Text` rows into sections.
