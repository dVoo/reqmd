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

Each requirement has a `trace` attribute listing its upstream IDs. reqmd validates that every trace target exists:

````markdown
## SYS-001: Boot Sequence
```attr
status: approved
requires-trace-from: [software, tests]
trace: [stakeholder/STK-GOAL-001]
```
````

The `requires-trace-from` attribute declares which downstream levels must trace back to this requirement. reqmd checks coverage both ways.

> **Hint:** The tokens here (`software`, `tests`) are `document-id`s. A `requires-trace-from` token can also name a `level` (e.g. `software-requirements`) to match every document at that layer — see the [x-reqmd options](/cheat-sheet/#x-reqmd-options) and the [level FAQ](/faq/#what-is-the-level-directive-in-x-reqmd-for).

## Validate the full tree

```sh
reqmd check 02-trace-your-spec/
```

Expected output:

```
Summary: 5 total, 5 valid, 0 invalid, 0 parse errors
```

`SW-001` is still `draft`, so you'll also see the status-aware coverage message described in [Step 5](/quickstart/05-status-disposition/).

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
