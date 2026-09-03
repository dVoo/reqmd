---
title: "Step 2 — Trace your spec"
description: "Build the four-level V-model boot-sequence example with trace links between levels."
weight: 2
---

Build a multi-level spec tree with trace links between levels.

This is the heart of reqmd — and it's the **example that runs through the rest of the quickstart**. The spec tree in this step is a car's *boot sequence* traced through the V-model:

```
02-trace-your-spec/
  stakeholder/          # Stakeholder needs (top of V)
    schema.yaml
    goals.md             # STK-GOAL-001: Fast Boot
  system/               # System requirements
    schema.yaml          # upstream: ../stakeholder/
    features.md          # SYS-001 traces to STK-GOAL-001
  software/             # Software requirements
    schema.yaml          # upstream: ../system/
    components.md        # SW-001 traces to SYS-001
  tests/                 # Test specifications (bottom of V)
    schema.yaml          # upstream: ../software/
    boot-test.md         # TEST-001/002 trace to SW-001 and SYS-001
```

Steps 3–11 keep working on this same tree: exporting it, serving it, pinning its versions, and finally verifying its tests.

## How trace links work

Each requirement has a `trace` attribute listing its upstream IDs. reqmd validates that every trace target exists.

Downstream coverage expectations are declared **once per document** in the `x-reqmd.requires-trace-from` key of `schema.yaml`, and every requirement in that document that does not declare its own `requires-trace-from` attribute inherits it:

```yaml
# system/schema.yaml
x-reqmd:
  level: system-requirements
  document-id: system
  requires-trace-from: [software, tests]   # SYS-001 inherits this
```

A requirement may override the document default by declaring the attribute itself, and `requires-trace-from: []` opts a single requirement out. Here the `tests/` layer sets an empty document default — a bottom-of-V layer expects no further downstream coverage. reqmd checks coverage both ways.

> **Hint:** The tokens above (`software`, `tests`) are `document-id`s. A `requires-trace-from` token can also name a `level` (e.g. `software-requirements`) to match every document at that layer — see the [x-reqmd options](/cheat-sheet/#x-reqmd-options) and the [level FAQ](/faq/#what-is-the-level-directive-in-x-reqmd-for).

## Validate the full tree

```sh
reqmd check 02-trace-your-spec/
```

Expected output:

```
Summary: 5 total, 5 valid, 0 invalid, 0 parse errors, 1 warnings
```

The single warning is `SYS-001`: its only `software`-level inbound (`SW-001`) is still `draft`, and a draft requirement does not close a coverage expectation. Promote `SW-001` to `approved` to make it green — see [Step 5](/quickstart/05-status-disposition/).

## List all requirements

```sh
reqmd ls 02-trace-your-spec/
```

## Stats per document

```sh
reqmd stats 02-trace-your-spec/
```

## Where to look things up

- The `trace` / `requires-trace-from` semantics: [x-reqmd options in the cheat sheet](/cheat-sheet/#x-reqmd-options)
- The `check`, `ls`, and `stats` flags: [`reqmd check`](/cheat-sheet/#reqmd-check--validate-requirements-and-trace-links), [`reqmd ls`](/cheat-sheet/#reqmd-ls--list-all-requirements), [`reqmd stats`](/cheat-sheet/#reqmd-stats--attribute-value-breakdown-per-document)

## What's next

The spec tree validates. Now export it for stakeholders — go to [Step 3](/quickstart/03-export/).
