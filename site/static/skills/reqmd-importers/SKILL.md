---
name: reqmd-importers
description: Write a Python importer that pulls requirements from another tool (DOORS, codeBeamer, Polarion, Jama, a CSV export, a REST API, a database) into a validated reqmd spec tree. Use when asked to import, mirror, sync, or convert external requirements into reqmd Markdown, or when working with the reqmd-import/v1 exchange format, reqmd_io, reqmd-import, or reqmd-code bundles. Triggers include "write an importer", "import DOORS into reqmd", "DOORS/codeBeamer CSV to requirements", "mirror an ALM into the spec", "exchange format", "reqmd-import", "reqmd_io", "migrate our requirements to Markdown".
---

# Writing a reqmd importer

How to get requirements out of another tool and into a reqmd tree **without
hand-writing Markdown or a schema**.

## When to use

- Mirroring requirements from an ALM (DOORS, codeBeamer, Polarion, Jama) into
  reqmd as generated, diff-reviewed documents.
- One-time migration (`migrate our legacy spec to Markdown`).
- A nightly/CI sync from a CSV, REST API, database, or ReqIF export.
- Debugging a bundle that `reqmd-import` rejected.

Not for: authoring requirements by hand (use `reqmd-authoring`), or operating
an existing tree (use `reqmd-usage`).

## Ground rules

1. **You write JSON. You do not write Markdown or YAML.** The pipeline is
   `producer → reqmd-import/v1 JSON → reqmd-import → .md + schema.yaml`. A
   second Markdown writer in Python *will* drift from the renderer on quoting,
   heading depth, and schema rules. There is exactly one renderer; use it.
2. **One-way, content down.** The ALM is authoritative for what it owns; reqmd
   is authoritative downstream. Never write imported prose back into the ALM.
   Pushing *computed* facts (coverage, verdicts) back is a separate, safe job —
   they are recomputed every run so they cannot conflict.
3. **Keep the source tool's own stable ID.** `native_id` is the DOORS absolute
   number or the codeBeamer item ID, not a hash and not a row number. The
   rendered ID becomes `<id-prefix><native-id>`. Re-imports must never churn
   IDs, or every `trace:` pointing at an imported requirement breaks.
4. **Never invent a reqmd ID.** Supply `native_id`; the renderer stamps it.
5. **Treat the renderer's rejections as bugs in your script**, not obstacles.
   They exist to stop a malformed document at the boundary.
6. **Imports land in a dedicated directory** that nothing hand-edits.

## The pipeline

```
ALM / CSV / API ──▶ your Python script ──▶ bundle.json ──▶ reqmd-import ──▶ spec/00-doors/
   (fetch + field map)      (reqmd_io)      (v1 format)      (renders)         (.md + schema.yaml)
                                                                                    │
                                                                            reqmd check spec/
```

## Procedure

### 1. Decide the boundary

Write down, before coding: which documents you produce, and which of them is
authoritative where. A useful default is one imported document per source
module, marked `external: true` (a proxy for content authored elsewhere).

### 2. Choose the document shape

One `Document` per output directory:

| Field | Notes |
|---|---|
| `dir` | Output subdirectory. `""` = the output root. Must not escape it. |
| `id_prefix` | Prepended to every `native_id`. Use the source tool's tag: `DOORS-`, `CB-`. |
| `level` | `x-reqmd.level`, e.g. `stakeholder-requirements`. Enables level-typed coverage. |
| `document_id` | Stable id for qualified refs (`document_id/ID`) — survives moves and submodule assembly. |
| `title` / `file` | Level-1 heading; `.md` filename (default `requirements.md`). `file` must be a plain file name — no `/`, no `..`. |
| `external` | `true` for a mirror of content authored elsewhere (suppresses untraced noise, pins out of scope). |
| `upstream` | `{"level": "...", "sources": ["../01-stakeholder/"]}` for doc-chain navigation. |
| `properties` / `required` | The attr-block JSON Schema — see step 5. |

A requirement-less document is valid and emits only `schema.yaml`. Use one as
an umbrella (shared level + upstream) when producing many sibling documents.

### 3. Fetch and map items

```python
from reqmd_io import Bundle, Document, Requirement

doc = Document(id_prefix="DOORS-", dir="brakes", file="package.md",
               title="Brakes Module", level="stakeholder-requirements",
               external=True,
               properties={"priority": {"type": "string"}})

for item in fetch_doors_module():          # your code: REST / DXL / CSV
    doc.add(Requirement(
        native_id=item.absolute_number,
        title=item.heading,
        body=item.text,
        parent=item.parent_number or "",   # native id, same document
        attrs={"priority": item.priority},
    ))

Bundle([doc]).write("brakes.json")
```

- **`body` is Markdown.** Convert plain text to paragraphs if the source has
  newlines; a raw multi-line string is written verbatim into the document.
  Headings inside a body are fine — they become info items under the
  requirement — but a fence whose info string is `attr` directly under a
  heading is rejected, because in reqmd that pair *is* a requirement: an
  imported description could otherwise add requirements of its own.
- **`title` is the heading text** after the ID; omit it for a bare heading.
- **`status` defaults to `approved`**, which is what makes imported items
  satisfy downstream coverage. Set `draft` only if the source really is a draft.
- **Keep provenance**: add `attrs={"x-reqmd.external-id": item.id,
  "x-reqmd.external-url": item.web_url}` so reviewers get a deep link back — and
  declare both in `properties`, or the generated schema rejects them.

### 4. Map links to `trace` vs `native_trace`

This is the payoff: relations become validated graph edges.

- `trace` — **fully-qualified reqmd IDs**, passed through verbatim. Use for
  links to hand-authored requirements (`"SYS-BRAKE-001"`, `"system/SYS-001"`).
- `native_trace` — **native IDs of other items in this bundle**. The renderer
  rewrites each to that item's full ID. Use for links to sibling imported items,
  including items in a *different* document of the same bundle.

When one source column mixes both, classify (the shipped CSV example does
exactly this):

```python
import re
REQMD_ID = re.compile(r"^(?:[a-z0-9_-]+/)?[A-Z][A-Z0-9-]+(?:~\d+)?$")

trace, native_trace = [], []
for link in split_links(item.links):       # your separator
    (trace if REQMD_ID.match(link) else native_trace).append(link)
```

`native_id`s must be **unique across the whole bundle** — that is what makes
cross-document `native_trace` unambiguous. If two source modules can reuse the
same key, prefix it with the module.

### 5. Declare the attr schema

Every attribute you emit must be declared, or the generated `schema.yaml`
rejects it (`additionalProperties: false`):

```python
properties={name: {"type": "string"} for name in attr_columns}
# tighter, when you know the domain:
properties={"priority": {"type": "string", "enum": ["High", "Medium", "Low"]}}
required=["priority"]          # only if every item really has it
```

Attr values may be strings, numbers, booleans, or **flat lists** of those.
Nested maps/lists are rejected — flatten them (`x.y` keys) instead.
`status` and `trace` are reserved: they have dedicated fields.

Attribute names may only contain letters, digits, `_`, `-` and `.`. A name with
anything else (a colon, say) is rejected, because reqmd reads the attr block
line by line and would silently file the value under a different name.

Send a JSON number or boolean when you mean one: a JSON *string* stays a string,
even when it reads like a number, `null`, or `true`.

### 6. Render and validate

```sh
pip install ./python                  # from the reqmd repo (not on PyPI)
go build -o reqmd-import ./reqmd-import/cmd/reqmd-import

python my_importer.py -o brakes.json
./reqmd-import brakes.json -o spec/00-doors/          # render
reqmd check spec/                                     # validate the whole tree
```

Add `--check` to `reqmd-import` to run `reqmd check` on the output
automatically (it needs `reqmd` on `PATH`).

Review the diff like any other change: the generated documents are ordinary
Markdown in the repo, so the PR diff *is* the review of the import.

## Format reference

`Requirement` fields (Python) / JSON keys:

| Field | Meaning |
|---|---|
| `native_id` / `native-id` | **Required.** Source tool's stable key. No whitespace — the rendered ID must be referenceable by `trace:`. |
| `title`, `body` | Heading text; Markdown body after the attr block (single-line title). |
| `status` | Built-in status; omitted → `approved`. |
| `parent` / `parent` | Native id of the parent **in the same document**. |
| `trace` | Fully-qualified reqmd IDs, verbatim. |
| `native_trace` / `native-trace` | Native ids resolved inside the bundle. |
| `attrs` | Everything else for the attr block. |

Bundle JSON shape:

```json
{
  "format": "reqmd-import/v1",
  "documents": [
    { "dir": "brakes", "id-prefix": "DOORS-", "level": "...",
      "properties": { "...": { "type": "string" } },
      "requirements": [
        { "native-id": "17", "title": "Brake request", "body": "...",
          "parent": "1", "native-trace": ["1"], "trace": ["SYS-BRAKE-001"],
          "attrs": { "priority": "High" } }
      ] }
  ]
}
```

Renderer guarantees worth relying on:

- Requirements may be listed in **any order**: children are reordered directly
  under their parents, to any depth (`##`, `###`, `####`, …).
- Unknown JSON fields and trailing data are **rejected** — a typo fails loudly.
- Duplicate `native_id`s, unknown `parent`/`native-trace`, reserved attrs,
  parent cycles, escaping `dir`s, and two documents claiming one output path are
  all rejected with a precise message.
- Content that would corrupt a document is rejected too, so a data-quality bug
  in the source tool fails here instead of silently changing the tree: an ID
  containing whitespace, a `file` that is not a plain file name, a line break in
  a title or notice, an attribute name outside `A-Z a-z 0-9 _ - .`, and a body
  that puts an `attr` fence directly under a heading.
- Output is byte-deterministic and idempotent. Files carry a
  `<!-- reqmd: generated -->` marker; **content without that marker is never
  overwritten**, and unchanged files are not rewritten.

## Worked example: DOORS / codeBeamer CSV

A generic reader for both tools ships in the repo — no code needed for the
common case:

```sh
python python/examples/csv_import.py rows.csv -o doors.json \
  --dir brakes --id-prefix DOORS- --title "Brakes Module" \
  --level stakeholder-requirements --external \
  --id-column "Absolute Number" --title-column "Object Heading" \
  --body-column "Object Text" --parent-column Parent \
  --trace-column Links --attr priority=Priority

reqmd-import doors.json -o spec/00-doors/ --check
```

Use it as the template for a source with different column names, or when you
need to fetch from an API instead of a file. See `quickstart/15-custom-importers/`
for the end-to-end walkthrough.

## Pitfalls

- **Do not hand-write `.md` or `schema.yaml`.** If the bundle format cannot
  express what you need, extend producers' use of `attrs`/`properties` — or
  raise it as a format change. Adding a second renderer is the failure mode this
  architecture exists to prevent.
- **A hand-authored `schema.yaml` in the output directory is never
  overwritten** (that is the marker guard). If your import "did nothing" to the
  schema, check for an existing hand-written document at that path — and give
  imports their own directory instead.
- **Colliding `native_id`s across documents.** Uniqueness is bundle-wide, not
  per document. Namespace the key.
- **An ALM key with a space or a slash in it.** DOORS and ReqIF frequently hand
you keys like `Item 42`. A space makes the rendered ID un-referenceable, so it
  is rejected: map the source key to a stable, reqmd-shaped one (usually
  `ReqIF.ForeignID`) and keep the original in `attrs`
  ("x-reqmd.external-id") for the way back.
- **`parent` is document-local.** Cross-document parents are rejected; express
  cross-document relationships as traces, not nesting.
- **Declaring an attribute you never emit, or emitting one you never declare.**
  Both fail validation. Keep the two in sync — derive `properties` from the same
  mapping you build `attrs` from.
- **Assuming `trace` resolves.** A ref to a requirement outside the tree is a
  WARNING (`broken reference`), so exit code 0 does not mean the link is good.
  Import the upstream document too, or fix the ID.
- **Assuming `--check` found the binary.** `--check` shells out to `reqmd` on
  `PATH`; run `reqmd check` yourself if it is not installed.
- **`external: true` does not mean "ignore me".** It suppresses untraced noise
  and pins; the document is still validated and still participates in coverage.
- **Never edit a generated document by hand.** The next import overwrites it.
  Fix the producer, or make the requirement hand-authored in a different
  document.

## Verification

1. `reqmd-import bundle.json -o <dir>` succeeds (no rejection messages).
2. `reqmd check <root>` exits 0 with **0 invalid and 0 parse errors**; read the
   warnings — broken refs introduced by the import show up there.
3. Re-run the importer, then confirm the tree is unchanged: a second
   `reqmd-import` writes nothing (byte-identical output is not rewritten).
   Every ID in the tree is identical across the two runs.
4. Spot-check one rendered document: hierarchy is nested (`##`/`###`), the attr
   block is valid YAML, and `trace`/`native_trace` links resolved to full IDs.
5. `reqmd ls <root> | grep '<ID>'` finds an imported requirement you expect.
