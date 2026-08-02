---
title: "Step 9 — Baseline diff"
description: "Compare requirement baselines between two git tags — no checkout needed."
weight: 9
---

Compare requirement baselines between two git tags — no checkout needed.

## How it works

`reqmd baseline diff <tag1> <tag2>` extracts the repository at each tag via `git archive | tar`, parses both versions through the standard pipeline, and produces a semantic diff:

- **Requirements**: added (`+`), removed (`-`), modified (`~`) with attribute-level detail
- **Schemas**: new/removed properties, changed required fields
- **Submodules**: added/removed/updated submodule pins

The command always exits `0` — the diff is informational, not validation.

## Try it

For a Git-backed spec repo (like the one holding your traced tree):

```sh
# Compare two release tags
reqmd baseline diff v1.0.0 v1.1.0
```

Example output:

```
Requirements
  + IVI-FUN-004  added in v1.1.0
  - IVI-FUN-002  removed in v1.1.0
  ~ IVI-FUN-001
      priority: medium -> high
      trace: ["SYS-001"] -> ["SYS-001","SAFE-003"]

Schemas
  ~ ivi-requirements
      + property: owner
      - property: priority
      required: ["priority","maturity","status","verify"] -> ["priority","maturity","status","verify","owner"]
```

## JSON output for CI

```sh
reqmd baseline diff v1.0.0 v1.1.0 --json
```

Produces a structured report suitable for release-note generation or CI gating.

## Submodule changes

When the repo has git submodules, the diff also reports submodule pin changes:

```
--- Submodule Changes ---
  vendor/spec-a:  a1b2c3d  →  e4f5g6h  (updated)
  vendor/spec-b:  f7g8h9i  →  (removed)
  vendor/spec-c:  (added)   →  j0k1l2m
```

This section is hidden when no submodules exist.

## Where to look things up

- All `baseline diff` flags and output shapes: [`reqmd baseline diff` in the cheat sheet](/cheat-sheet/#reqmd-baseline-diff--compare-two-git-tags)

## What's next

Now close the right side of the V-model — load verification results for the boot tests. Go to [Step 11](/quickstart/11-verification-results-ctrf/).
