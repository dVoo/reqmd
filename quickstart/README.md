# reqmd Quickstart

A step-by-step guide to using `reqmd`. Each folder is one step — work through
them in order. Lettered steps (e.g. `02a`) are optional detours into specific
features.

## Steps

| Step | Folder | What you learn |
|------|--------|----------------|
| 1 | [`01-get-started/`](01-get-started/) | Scaffold a new project and run your first validation |
| 1a | [`01a-ci-integration/`](01a-ci-integration/) | JSON reports, exit codes, GitHub Actions integration |
| 2 | [`02-trace-your-spec/`](02-trace-your-spec/) | Build a multi-level spec tree with trace links (stakeholder → system → software → tests) |
| 2a | [`02a-status-disposition/`](02a-status-disposition/) | Status lifecycle (draft vs approved), disposition workflow (deferred/rejected) |
| 2b | [`02b-version-pins/`](02b-version-pins/) | Version pinning with `~N`, stale-pin detection, `--relaxed-versions` |
| 2c | [`02c-repin/`](02c-repin/) | Bulk-update `~N` version pins to match upstream versions with `reqmd repin --yes` |
| 3 | [`03-export/`](03-export/) | Export requirements to CSV, HTML, and graph formats |
| 3a | [`03a-live-preview/`](03a-live-preview/) | Live-reloading HTML preview with `reqmd serve` |
| 3b | [`03b-baseline-diff/`](03b-baseline-diff/) | Compare requirement baselines between git tags |
| 4 | [`04-custom-templates/`](04-custom-templates/) | Define your own requirement schema with a custom `reqmd init` preset |
| 5 | [`05-verification-results-ctrf/`](05-verification-results-ctrf/) | Load automated test results (CTRF) and run outcome-gated checks |
| 6 | [`06-review-documentation/`](06-review-documentation/) | Load manual review results (inspection, analysis, demonstration) |

## Suggested path

**New users:** 1 → 2 → 3 → 4 → 5 → 6

**CI/automation focus:** 1 → 1a → 2 → 3b → 5

**Automotive SPICE focus:** 1 → 2 → 2a → 2b → 5 → 6

## Prerequisites

Build the `reqmd` binary:

```sh
go build -o reqmd ./cmd/reqmd
```

Each step's README shows the exact commands to run. Paths are relative to the
project root.