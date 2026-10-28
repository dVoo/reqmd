# reqmd-io

Build [`reqmd-import/v1`](../reqmd-import) exchange bundles from Python.

This library only emits the JSON exchange format — it contains no Markdown or
YAML logic. `reqmd-import` is the single renderer that turns a bundle into
reqmd documents (`schema.yaml` + `.md`), so an importer written in Python
cannot drift from the Go renderer.

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
from reqmd_io import Bundle, Document, Requirement

doc = Document(id_prefix="DOORS-", dir="brakes", level="stakeholder-requirements",
               title="Brakes Module", external=True,
               properties={"priority": {"type": "string", "enum": ["High", "Low"]}})

doc.add(Requirement(native_id="17", title="Brake request",
                    body="The pedal shall request braking.",
                    attrs={"priority": "High"}))
doc.add(Requirement(native_id="42", title="Brake pressure", parent="17",
                    native_trace=["17"], trace=["SYS-001"],
                    attrs={"priority": "Low"}))

Bundle([doc]).write("brakes.json")
```

```sh
reqmd-import brakes.json -o spec/00-doors/
reqmd check spec/
```

Key points:

- `native_id` is the source tool's own stable ID (a DOORS object number, a
  codeBeamer item ID). The reqmd ID becomes `<id_prefix><native_id>`, so
  re-imports never churn IDs and hand-authored `trace:` links keep resolving.
  It must not contain whitespace (an `Item 42` key is rejected: nothing could
  ever `trace:` to the ID it produces).
- `trace` holds fully-qualified reqmd IDs. `native_trace` holds native IDs of
  other items in the same bundle and is rewritten to that item's full ID.
- `attrs` take strings, numbers, booleans, or flat lists of those. A string
  stays a string even when it looks like a number or `null`, so pass a real
  number or boolean when you mean one. Names may only contain letters, digits,
  `_`, `-` and `.`. The built-ins `status` and `trace` are reserved (there are
  dedicated fields).
- `body` is Markdown and may contain headings (they become info items), but not
  a fence whose info string is `attr` directly under a heading — that pair would
  become a requirement of its own.
- `properties` declares the attr-block JSON Schema; regenerating with a new
  attr means adding it there.

A complete, runnable importer is in [`examples/csv_import.py`](examples/csv_import.py)
— one generic reader for DOORS and codeBeamer CSV exports.

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
