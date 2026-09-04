---
title: "Step 11 — Verification results (CTRF)"
description: "Load automated test results in CTRF format and run outcome-gated checks against the boot-sequence spec."
weight: 11
---

Load automated test results in [CTRF format](https://ctrf.io/) and run outcome-gated checks against the [boot-sequence spec](/quickstart/02-trace-your-spec/).

## What's in this folder

```
11-verification-results-ctrf/
  boot-tests.ctrf.json    # CTRF report: TEST-001 passed
  coverage.json           # NOT a CTRF file; auto-skipped silently
```

## How CTRF maps to requirements

CTRF test entries bind to the spec via the `extra.x-reqmd` block (all fields optional):

```json
{
  "name": "Test_Boot_Time",
  "status": "passed",
  "extra": { "x-reqmd": { "id": "TEST-001" } }
}
```

| Field | What it does |
|---|---|
| `id` | Binds the result directly to a measure (requirement with a `verify:` attr). |
| `case` | A stable test-case identity. With `id`, it keys the result so several cases can attach to one measure; without `id`, it names a synthesized test case. |
| `verifies` | Upstream requirement IDs the test exercises. With `case` it synthesizes a test case that traces to them (tests-as-code). |
| `description` | Markdown body of a synthesized test case. |

A test with neither `id` nor `verifies` is skipped silently; a test that declares `x-reqmd` but binds nothing warns as `unbound-result` (suppress with `--ignore-unbound-results`).

CTRF `status` → reqmd `outcome`:

| CTRF status | reqmd outcome |
|---|---|
| `passed` | `pass` |
| `failed` | `fail` |
| `skipped` | `skipped` |
| `pending` / `other` | `inconclusive` |
| any + `flaky: true` | `inconclusive` |

## Run outcome-gated checks

```sh
reqmd check 02-trace-your-spec/ --results 11-verification-results-ctrf/
```

When results are loaded, reqmd runs two additional checks against each measure's **rolled-up evidence** (its own results plus results of its approved downstream test cases; strict aggregation — any fail → fail, else inconclusive, else skipped, else pass):

| Check | Level | When it fires |
|---|---|---|
| `missing-verdict` | WARNING | An approved measure has no evidence; draft downstream cases are reported as ignored |
| `failing-verdict` | ERROR | A measure's rolled-up verdict is `fail`; the message names the failing case(s) |

Draft and deferred/rejected measures are skipped. Result-attributed findings appear in the **Verification results** section.

The `coverage.json` file in this folder is silently skipped — reqmd auto-detects it's not a CTRF file (no `results.tests[]` object).

## Export with verdicts

```sh
# HTML with color-coded verdict badges and an expandable evidence list
reqmd export html 02-trace-your-spec/ --results 11-verification-results-ctrf/ -o html-out/

# CSV with Verdict, Verdict Source, and Verdict Cases columns
reqmd export csv 02-trace-your-spec/ --results 11-verification-results-ctrf/ -o csv-out/
```

## Where to look things up

- The `--results` flag and the `missing-verdict` / `failing-verdict` checks: [`reqmd check` in the cheat sheet](/cheat-sheet/#reqmd-check--validate-requirements-and-trace-links)

## What's next

CTRF covers automated tests. Manual verification methods (review, inspection, analysis) are documented as markdown — go to [Step 12](/quickstart/12-review-documentation/).
