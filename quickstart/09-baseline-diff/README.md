# Step 9 — Baseline Diff

Compare requirement baselines between two git tags — no checkout needed.

## How it works

`reqmd baseline diff <tag1> <tag2>` extracts the repository at each tag via
`git archive | tar`, parses both versions through the standard pipeline, and
produces a semantic diff:

- **Requirements**: added (`+`), removed (`-`), modified (`~`) with
  attribute-level detail. A requirement that moved to another section is a
  modification as well: it is reported as a `parent:` delta, because the
  section it sits under is part of what the requirement says. Unchanged
  requirements are not listed.
- **Schemas**: new/removed top-level keys, changed values such as `required`
- **Submodules**: added/removed/updated submodule pins, when the tree has
  submodules (read with `git ls-tree` at each tag)

The diff result itself is informational, not validation: an empty diff, a huge
one, or newly added requirements never changes the exit code. Real failures do
exit non-zero — a missing git repository, an unreadable tag, a parse error, or
an invalid `--filter` expression.

## Try it

This repo uses Jujutsu, not Git, so you can't run `baseline diff` directly on
it. But for a Git-backed spec repo:

```sh
# Compare two release tags
reqmd baseline diff v1.0.0 v1.1.0
```

Example output (illustrative — there is no tag pair with a usable diff in this
repo; the shapes are the ones the command prints):

```text
=== Baseline Diff: v1.0.0 → v1.1.0 ===

--- 02-system/ ---
  + SYS-FMT-004  (added)
  ~ SYS-FMT-001  (modified)
      priority: "High" → "Critical"
      trace: [STK-GOAL-001] → [STK-GOAL-001, SAFE-003]
  ~ SYS-FMT-002  (modified)
      parent: "Traceability" → "Format"

--- 04-tests/ ---
  - TEST-003  (removed)

--- Schema Changes ---
02-system/
  ~ required: [priority, status] → [priority, risk, status]

Summary: 1 added, 1 removed, 2 modified, 1 unchanged
```

One block per document directory (`--- <dir>/ ---`), one line per changed
requirement, then indented lines for each attribute delta: `key: old → new`, or
`key: value (new)` / `key: value (removed)` when the attribute appeared or
vanished. A `parent:` delta means the requirement now lives under a different
section even though its attributes did not change.

## JSON output for CI

```sh
reqmd baseline diff v1.0.0 v1.1.0 --json
```

Produces a structured report suitable for release-note generation or CI
gating.

## Submodule changes

When the repo has git submodules, the diff also reports submodule pin changes:

```text
--- Submodule Changes ---
  vendor/spec-a:  a1b2c3d  →  e4f5g6h  (updated)
  vendor/spec-b:  f7g8h9i  →  (removed)
  vendor/spec-c:  (added)   →  j0k1l2m
```

Short hashes are 7 characters. This section is hidden when no submodules exist,
and it appears before the `Summary:` line.

## What's next

Now close the right side of the V-model — load verification results for the
boot tests. Go to [Step 11](../11-verification-results-ctrf/).