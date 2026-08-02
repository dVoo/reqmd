---
title: "Step 4 — Live preview"
description: "Serve a live-reloading HTML preview of the boot-sequence spec while you edit it."
weight: 4
---

Serve a live-reloading HTML preview of the [boot-sequence spec](/quickstart/02-trace-your-spec/) while you edit it.

## Start the server

```sh
reqmd serve 02-trace-your-spec/
```

This opens a browser at `http://localhost:8080` showing the HTML export with live reload. Every time you save a `.md` or `schema.yaml` file, the browser auto-refreshes via SSE (Server-Sent Events).

## Flags

| Flag | Default | What it does |
|------|---------|--------------|
| `--addr` | `localhost:8080` | Listen address |
| `--headless` | `false` | Terminal-only mode (no HTTP server) |
| `--no-open` | `false` | Don't open a browser on start |
| `--debounce` | `500ms` | Debounce window for file-change events |
| `--results` | _(none)_ | Load verification results (CTRF `.ctrf.json` or manual-results dirs). Repeatable. Result file changes also trigger rebuild. |

## Usage

```sh
# Default — opens browser
reqmd serve 02-trace-your-spec/

# Terminal-only (remote server, CI)
reqmd serve 02-trace-your-spec/ --headless --addr 0.0.0.0:8080

# With verification results — verdict badges on measure cards
reqmd serve 02-trace-your-spec/ --results 11-verification-results-ctrf/
reqmd serve 02-trace-your-spec/ --results 12-review-documentation/
```

## With verification results

Pass `--results` (repeatable) to load CTRF or manual verification results. Verdict badges appear on measure cards (pass / fail / skipped / inconclusive). Result files are watched alongside spec files — editing a CTRF JSON or manual-results markdown triggers an immediate rebuild and browser refresh.

```sh
reqmd serve 02-trace-your-spec/ --results 11-verification-results-ctrf/
```

The `failing-verdict` graph check runs on every rebuild: a `fail` verdict on a measure produces an ERROR-level finding, turning the status line red.

## Where to look things up

Every `reqmd serve` flag in one table: [`reqmd serve` in the cheat sheet](/cheat-sheet/#reqmd-serve--live-reloading-html-preview)

## What's next

Now look at the status lifecycle that makes `SW-001` a draft — go to [Step 5](/quickstart/05-status-disposition/).
