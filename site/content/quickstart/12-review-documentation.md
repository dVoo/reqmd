---
title: "Step 12 — Review documentation"
description: "Load manual verification results (review, inspection, analysis, demonstration) recorded as markdown with a user-supplied schema."
weight: 12
---

Load manual verification results (review, inspection, analysis, demonstration) recorded as markdown with a user-supplied schema.

## What's in this folder

```
12-review-documentation/
  schema.yaml                    # schema for manual results
  design-review-results.md       # VR-001: design review for TEST-002 (pass)
```

## Manual results format

Manual results are markdown files with `attr` blocks, just like requirements — but the schema declares result-specific attributes:

```yaml
# schema.yaml
x-reqmd:
  level: verify-results
required: [outcome, verifier, verified-at]
properties:
  outcome:
    type: string
    enum: [pass, fail, skipped, inconclusive]
  verifier:
    type: string
  evidence:
    type: string
  verified-at:
    type: string
```

Each result is a heading with an `attr` block:

````markdown
## VR-001: Boot Sequence Design Review Result
```attr
status: approved
outcome: pass
verifier: "A. Reviewer"
evidence: minutes/2026-07-15-boot-review.md
verified-at: "2026-07-15"
trace: [TEST-002]
```
````

The `trace` attribute links the result to the measure it verifies.

## Scaffold a results directory

Use the built-in `results` preset:

```sh
reqmd init my-reviews/ --preset results --id-prefix VR
```

This creates `schema.yaml` + `results.md` with one example result.

## Run checks with manual results

```sh
reqmd check 02-trace-your-spec/ --results 12-review-documentation/
```

## Combine automated + manual results

```sh
reqmd check 02-trace-your-spec/ \
  --results 11-verification-results-ctrf/ \
  --results 12-review-documentation/
```

Both result sources are merged — if a measure has results from both CTRF and manual sources, the latest `verified-at` wins.

## Export with manual verdicts

```sh
reqmd export html 02-trace-your-spec/ \
  --results 12-review-documentation/ -o html-out/
```

## That's it

You've worked through the full reqmd workflow:

1. Scaffold a project
2. Build a traced V-model spec tree
3. Export, preview, and diff it
4. Run it in CI with version pins
5. Load automated test results (CTRF)
6. Load manual review results

For the full spec, see the [README](https://github.com/dVoo/reqmd) and [AGENTS.md](https://github.com/dVoo/reqmd).

## Where to look things up

- The `results` preset and `--results` flag: [`reqmd init`](/cheat-sheet/#reqmd-init--scaffold-a-new-project) and [`reqmd check`](/cheat-sheet/#reqmd-check--validate-requirements-and-trace-links) in the cheat sheet

## What's next

The V-model is fully verified. Now extend reqmd to your own document types — go to [Step 13](/quickstart/13-custom-templates/).
