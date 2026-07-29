---
title: "Quickstart"
description: "Install reqmd and take a seven-step tour of the workflow — from your first requirement to a fully traced V-model spec tree, validated and exported."
weight: 2
---

In the next 10 minutes you can go from a fresh `reqmd` install to a fully traced V-model spec tree, validated and exported. Work through the steps in order — lettered steps are optional detours.

## The seven steps

| Step | Folder | What you'll learn |
|------|--------|-------------------|
| 0 | [Installation](/quickstart/00-installation/) | `go install`, build from source, or prebuilt binary |
| 1 | [Get started](/quickstart/01-get-started/) | Scaffold a new project and run your first validation |
| 1a | [CI integration](/quickstart/01a-ci-integration/) | JSON reports, exit codes, GitHub Actions |
| 2 | [Trace your spec](/quickstart/02-trace-your-spec/) | Build a four-level V-model with trace links |
| 2a | [Status & disposition](/quickstart/02a-status-disposition/) | The `draft`/`approved` lifecycle and `deferred`/`rejected` workflow |
| 2b | [Version pins](/quickstart/02b-version-pins/) | Pin traces to a specific upstream version; detect when stale |
| 2c | [Repin](/quickstart/02c-repin/) | Bulk-update `~N` version pins with `reqmd repin --yes` |
| 3 | [Export](/quickstart/03-export/) | CSV, HTML, and graph export formats |
| 3a | [Live preview](/quickstart/03a-live-preview/) | `reqmd serve` with live-reloading HTML |
| 3b | [Baseline diff](/quickstart/03b-baseline-diff/) | Compare requirements between two git tags |
| 4 | [Custom templates](/quickstart/04-custom-templates/) | Define your own schema with a custom `reqmd init` preset |
| 5 | [Verification results (CTRF)](/quickstart/05-verification-results-ctrf/) | Load automated test results and run outcome-gated checks |
| 6 | [Review documentation](/quickstart/06-review-documentation/) | Load manual review, inspection, and analysis results |

## Suggested paths

**New users:** `00` → `01` → `02` → `03` → `04` → `05` → `06`

**CI/automation focus:** `00` → `01` → `01a` → `02` → `03b` → `05`

**Status & lifecycle focus:** `00` → `01` → `02` → `02a` → `02b` → `05` → `06`

## Prerequisites

Build the `reqmd` binary:

```sh
go build -o reqmd ./cmd/reqmd
```

Each step's page shows the exact commands to run. Paths are relative to the project root. The full source for every step lives in the [`quickstart/`](https://github.com/dVoo/reqmd/tree/main/quickstart) directory of the repo.

**Beyond the spec itself:** the companion [`reqmd-import`](/import/) tool extracts requirement IDs from Go and Python source via tree-sitter. After you've written the spec for a feature, run `reqmd-import extract` and the implementation gets IDs in the same namespace — closing the loop from spec to code.
