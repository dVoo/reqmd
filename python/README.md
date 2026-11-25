# reqmd-io

Build [exchange-format](../site/content/exchange/v1.md) bundles from Python.

This library only emits the JSON exchange format — it contains no Markdown or
YAML logic. `reqmd-import` is the single renderer that turns a bundle into
reqmd documents (`schema.yaml` + `.md`), so an importer written in Python
cannot drift from the Go renderer.

The contract: [exchange format v1](/site/content/exchange/v1.md) ·
[`schema.json`](../site/static/exchange/v1/schema.json).

## Install

```sh
pip install reqmd-io
```

Or from the repo root:

```sh
pip install ./python
```

## Use

```python
from reqmd_io import Bundle, Document, Requirement, Section

doc = Document(
    key="brakes",
    title="Brakes Module",
    x_reqmd={"id-prefix": "DOORS-", "level": "stakeholder-requirements",
             "external": True, "dir": "brakes"},
    properties={"priority": {"type": "string", "enum": ["High", "Low"]}},
)

doc.add(Section(title="Hydraulics", body="The hydraulic circuit."))
doc.add(Requirement(key="17", title="Brake request",
                    body="The pedal shall request braking.",
                    fields={"priority": "High"}))
doc.add(Requirement(key="42", title="Brake pressure",
                    links=["17"], refs=["SYS-001"],
                    fields={"priority": "Low"}), parent="17")

Bundle([doc]).write("brakes.json")
```

```sh
reqmd-import brakes.json --validate-only   # report every violation, write nothing
reqmd-import brakes.json -o spec/00-doors/
reqmd check spec/
```

Key points:

- `kind` separates content from structure: `Requirement` gets a reqmd block and
  is validated, traced and counted; `Section` is a heading with prose and
  children. Both are the same wire `kind` — folder-versus-leaf is derived from
  whether a section has children, so adding a subsection never changes a type.
- `key` is the source tool's own stable ID (a DOORS object number, a codeBeamer
  item ID), unique per document. The reqmd ID becomes
  `<x_reqmd["id-prefix"]><key>`, so re-imports never churn IDs and
  hand-authored `trace:` links keep resolving. It must not contain whitespace.
- `links` name other items by key (`"17"`, or `"brakes/17"` across documents)
  and are rewritten to full requirement IDs. `refs` hold IDs owned elsewhere
  and are passed through verbatim.
- `doc.add(item, parent=key)` nests by key, so a flat source table stays a flat
  loop. The parent must already have been added.
- `fields` are the reqmd-block entries (sections never carry them). Values take
  strings, numbers, booleans or flat lists of those; a string stays a string
  even when it looks like a number or `null`, so pass a real number or boolean
  when you mean one. Names may only contain letters, digits, `_`, `-` and `.`.
- `body` is Markdown and **must not contain headings** — structure belongs in
  sections.
- `status` is your own vocabulary. `reqmd-import --status-map "In Review=review"`
  translates it; an unmapped value falls back to reqmd's default and is kept as
  `x-reqmd.source-status`.
- `properties` declares the reqmd-block JSON Schema; a new attribute means
  adding it there.

A complete, runnable importer is in [`examples/csv_import.py`](examples/csv_import.py)
— one generic reader for DOORS and codeBeamer CSV exports, with
`--kind-column` to map a row-type column onto requirement versus section.

## Test

```sh
cd python && python -m unittest discover -s tests
```

## Releasing

`reqmd-io` versions **independently of the reqmd binaries**: bump it only for
a new feature or an exchange-format change, not for every reqmd release. The
version in `pyproject.toml` is the single source of truth.

```sh
# 1. bump the version in pyproject.toml (PyPI never allows a reuse)
# 2. land it on main, then tag with the py-v prefix (NOT the bare v prefix):
git tag py-v0.1.0 && git push origin py-v0.1.0
```

The `py-v*` tag triggers `.github/workflows/publish-python.yml`, which runs the
tests, builds the sdist and wheel, runs `twine check`, and fails if the tag and
the packaged version disagree. Publishing uses PyPI Trusted Publishing, so no
API token secret is stored.

To rehearse before a real release, build locally and upload to TestPyPI:

```sh
cd python
python -m build && twine check dist/*
twine upload -r testpypi dist/*
```

First-time setup (once, on PyPI): add a pending publisher for the project
`reqmd-io` with owner `dVoo`, repository `reqmd`, workflow
`publish-python.yml`, and environment `pypi`.
