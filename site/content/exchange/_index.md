---
title: "Exchange format"
description: "The JSON contract between requirement producers (ALM exports, source-code extractors, scripts) and reqmd — field reference, rules, and the published JSON Schema."
---

Every importer feeds reqmd through one contract: a JSON **exchange bundle**.
A producer writes it, `reqmd-import` renders it into `schema.yaml` plus
Markdown. Because the format is neutral — no reqmd-only required fields, no
YAML to get right — a producer can be a few dozen lines of Python, a Go binary,
or an export script against a database.

## The two directions

```text
DOORS / codeBeamer / CSV ──▶ your producer ──┐
source code ──▶ reqmd-code ──────────────────┼─▶ exchange bundle ──▶ reqmd-import ──▶ .md + schema.yaml
                                             ┘
```

- [**Exchange format v1 →**](/exchange/v1/) — the field reference, the rules,
  and two complete examples.
- [JSON Schema](/exchange/v1/schema.json) — load it into your producer's tests
  and validate a bundle before you ever run `reqmd-import`.
- [`reqmd-code` and ALM imports](/import/) — the built-in producer and the
  CSV example.

## Try it

```sh
reqmd-import bundle.json -o spec/alm --validate-only   # check the bundle, write nothing
reqmd-import bundle.json -o spec/alm                   # render it
reqmd check spec/                                      # validate the result
```
