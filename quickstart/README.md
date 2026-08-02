# reqmd Quickstart

A step-by-step guide to using `reqmd`. Each folder is one step — work through
them in order, from easy everyday use to more specific topics. Steps 2–9 and
11–12 all build on the **same example**: a car's *boot sequence* spec traced
through the V-model — stakeholder need → system requirement → software
component → tests → verification results.

For the rendered tutorial with full explanations, see the [site
quickstart](https://reqmd.dev/quickstart/).

## Steps

| Step | Folder | What you learn |
|------|--------|----------------|
| 1 | [`01-get-started/`](01-get-started/) | Scaffold a new project and run your first validation |
| 2 | [`02-trace-your-spec/`](02-trace-your-spec/) | Build a multi-level spec tree with trace links (stakeholder → system → software → tests) |
| 3 | [`03-export/`](03-export/) | Export requirements to CSV, HTML, and graph formats |
| 4 | [`04-live-preview/`](04-live-preview/) | Live-reloading HTML preview with `reqmd serve` |
| 5 | [`05-status-disposition/`](05-status-disposition/) | Status lifecycle (draft vs approved), disposition workflow (deferred/rejected) |
| 6 | [`06-ci-integration/`](06-ci-integration/) | JSON reports, exit codes, GitHub Actions integration |
| 7 | [`07-version-pins/`](07-version-pins/) | Version pinning with `~N`, stale-pin detection, `--relaxed-versions` |
| 8 | [`08-repin/`](08-repin/) | Bulk-update `~N` version pins to match upstream versions with `reqmd repin --yes` |
| 9 | [`09-baseline-diff/`](09-baseline-diff/) | Compare requirement baselines between git tags |
| 10 | [`10-submodule-configuration/`](10-submodule-configuration/) | Use stable `document-id` references when documents are assembled from git submodules (optional) |
| 11 | [`11-verification-results-ctrf/`](11-verification-results-ctrf/) | Load automated test results (CTRF) and run outcome-gated checks |
| 12 | [`12-review-documentation/`](12-review-documentation/) | Load manual review results (inspection, analysis, demonstration) |
| 13 | [`13-custom-templates/`](13-custom-templates/) | Define your own requirement schema with a custom `reqmd init` preset |
| 14 | [`14-variant-management/`](14-variant-management/) | One spec tree, many configurations with `--filter` |

## Suggested paths

**New users:** 1 → 2 → 3 → 4 → 5 → 11 → 12

**CI/automation focus:** 1 → 2 → 6 → 9 → 11

**Automotive SPICE focus:** 1 → 2 → 5 → 7 → 11 → 12

## Prerequisites

Build the `reqmd` binary:

```sh
go build -o reqmd ./cmd/reqmd
```

Each step's README shows the exact commands to run. Paths are relative to the
project root.
