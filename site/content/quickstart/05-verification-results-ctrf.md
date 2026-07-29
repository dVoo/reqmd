---
title: "Step 5 — Verification results (CTRF)"
description: "Load automated test results in CTRF format and run outcome-gated checks against your spec tree."
weight: 11
---

Load automated test results in [CTRF format](https://ctrf.io/) and run outcome-gated checks against your spec tree.

## What's in this folder

```
05-verification-results-ctrf/
  boot-tests.ctrf.json    # CTRF report: TEST-001 passed
  coverage.json           # NOT a CTRF file; auto-skipped silently
```

## How CTRF maps to requirements

CTRF test entries map to measures (requirements with a `verify` attribute) via the `x-reqmd.id` extra field:

```json
{
  "name": "Test_Boot_Time",
  "status": "passed",
  "extra": { "x-reqmd": { "id": "TEST-001" } }
}
```

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
reqmd check 02-trace-your-spec/ --results 05-verification-results-ctrf/
```

When results are loaded, reqmd runs two additional checks:

| Check | Level | When it fires |
|---|---|---|
| `missing-verdict` | WARNING | An approved measure (`verify:` attr) has no result |
| `failing-verdict` | ERROR | A measure's latest result has outcome `fail` |

The `coverage.json` file in this folder is silently skipped — reqmd auto-detects it's not a CTRF file (no `results.tests[]` object).

## Export with verdicts

```sh
# HTML with color-coded verdict badges (green=pass, red=fail, etc.)
reqmd export html 02-trace-your-spec/ --results 05-verification-results-ctrf/ -o html-out/

# CSV with Verdict and Verdict Source columns
reqmd export csv 02-trace-your-spec/ --results 05-verification-results-ctrf/ -o csv-out/
```

## What's next

CTRF covers automated tests. Manual verification methods (review, inspection, analysis) are documented as markdown — go to [Step 6](/quickstart/06-review-documentation/).
