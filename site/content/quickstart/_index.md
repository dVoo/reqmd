---
title: "Quickstart"
description: "Install reqmd and take a step-by-step tour of the workflow — from your first requirement to a fully traced, verified, and exported spec tree."
weight: 2
---

In the next 10 minutes you can go from a fresh `reqmd` install to a fully traced V-model spec tree, validated, verified, and exported.

The steps are ordered from **easy, everyday use to more specific topics**. Steps 2–9 and 11–12 all build on the **same example**: a car's *boot sequence* spec traced through the V-model — stakeholder need → system requirement → software component → tests → verification results.

## reqmd in one glance

Five commands cover the everyday loop:

```sh
reqmd init my-project/            # scaffold a new spec tree
reqmd check my-project/           # validate every requirement + trace link
reqmd ls my-project/              # list all requirements
reqmd stats my-project/           # attribute-value breakdown per document
reqmd export html my-project/ -o out/   # share as a browsable HTML report
```

Command details, flags, and reference tables live in the [cheat sheet](/cheat-sheet/) — the quickstart links to it instead of repeating it.

## The steps

| Step | Page | What you'll learn |
|------|------|-------------------|
| 0 | [Installation](/quickstart/00-installation/) | `go install`, build from source, or prebuilt binary |
| 1 | [Get started](/quickstart/01-get-started/) | Scaffold a new project and run your first validation |
| 2 | [Trace your spec](/quickstart/02-trace-your-spec/) | Build the four-level V-model boot-sequence example with trace links |
| 10 | [Submodule-friendly configuration](/quickstart/10-submodule-configuration/) | Portable `document-id` references for specs assembled from git submodules (optional) |
| 3 | [Export](/quickstart/03-export/) | CSV, HTML, and graph export formats |
| 4 | [Live preview](/quickstart/04-live-preview/) | `reqmd serve` with live-reloading HTML |
| 5 | [Status & disposition](/quickstart/05-status-disposition/) | The `draft`/`approved` lifecycle and `deferred`/`rejected` workflow |
| 6 | [CI integration](/quickstart/06-ci-integration/) | JSON reports, exit codes, GitHub Actions |
| 7 | [Version pins](/quickstart/07-version-pins/) | Pin traces to a specific upstream version; detect when stale |
| 8 | [Repin](/quickstart/08-repin/) | Bulk-update `~N` version pins with `reqmd repin --yes` |
| 9 | [Baseline diff](/quickstart/09-baseline-diff/) | Compare requirements between two git tags |
| 11 | [Verification results (CTRF)](/quickstart/11-verification-results-ctrf/) | Load automated test results and run outcome-gated checks |
| 12 | [Review documentation](/quickstart/12-review-documentation/) | Load manual review, inspection, and analysis results |
| 13 | [Custom templates](/quickstart/13-custom-templates/) | Define your own schema with a custom `reqmd init` preset |
| 14 | [Variant management](/quickstart/14-variant-management/) | One spec tree, many configurations with `--filter` |

## Suggested paths

**New users:** `00` → `01` → `02` → `03` → `04` → `05` → `11` → `12`

**CI/automation focus:** `00` → `01` → `02` → `06` → `09` → `11`

**Status & lifecycle focus:** `00` → `01` → `02` → `05` → `07` → `11` → `12`

**Product-line / variant focus:** `00` → `01` → `02` → `03` → `09` → `14`

## Prerequisites

Build the `reqmd` binary:

```sh
go build -o reqmd ./cmd/reqmd
```

Each step's page shows the exact commands to run. Paths are relative to the project root. The full source for every step lives in the [`quickstart/`](https://github.com/dVoo/reqmd/tree/main/quickstart) directory of the repo.

**Beyond the spec itself:** the companion [`reqmd-import`](/import/) tool extracts requirement IDs from Go and Python source via tree-sitter. After you've written the spec for a feature, run `reqmd-import extract` and the implementation gets IDs in the same namespace — closing the loop from spec to code.
