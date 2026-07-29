# Step 3a — Live Preview

Serve a live-reloading HTML preview while you edit your spec tree.

## Start the server

```sh
reqmd serve 02-trace-your-spec/
```

This opens a browser at `http://localhost:8080` showing the HTML export with
live reload. Every time you save a `.md` or `schema.yaml` file, the browser
auto-refreshes via SSE (Server-Sent Events).

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
reqmd serve 02-trace-your-spec/ --results 05-verification-results-ctrf/
reqmd serve 02-trace-your-spec/ --results 06-review-documentation/
```

## With verification results

Pass `--results` (repeatable) to load CTRF or manual verification results.
Verdict badges appear on measure cards (pass / fail / skipped / inconclusive).
Result files are watched alongside spec files — editing a CTRF JSON or
manual-results markdown triggers an immediate rebuild and browser refresh.

```sh
reqmd serve 02-trace-your-spec/ --results 05-verification-results-ctrf/
```

The `failing-verdict` graph check runs on every rebuild: a `fail` verdict
on a measure produces an ERROR-level finding, turning the status line red.

## What's next

Now compare two baselines — go to [Step 3b](../03b-baseline-diff/).