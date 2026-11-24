# Step 4 — Live Preview

Serve a live-reloading HTML preview while you edit your spec tree.

## Start the server

```sh
reqmd serve 02-trace-your-spec/
```

This opens a browser at `http://localhost:8080` showing the project home —
document catalog, project-wide coverage and verification, actionable exception
queues, and cross-document search. Every document page has the same Document,
Table, Trace, and Coverage views as `reqmd export html`; every time you save a
`.md` or `schema.yaml` file, the browser auto-refreshes via SSE (Server-Sent
Events).

Every file event goes through the `--debounce` window (default `500ms`) before a
rebuild runs. A rebuild re-parses the tree and re-runs the graph checks, but it
does **not** re-run JSON Schema validation — run `reqmd check` to validate
attribute values.

## Flags

| Flag | Default | What it does |
|------|---------|--------------|
| `--addr` | `localhost:8080` | Listen address |
| `--headless` | `false` | Terminal-only mode (no HTTP server) |
| `--no-open` | `false` | Don't open a browser on start |
| `--debounce` | `500ms` | Debounce window for file-change events |
| `--results` | _(none)_ | Load verification results (CTRF `.ctrf.json` or manual-results dirs). Repeatable. Result file changes also trigger rebuild. |

`serve` always serves the whole tree: there is no `--filter`/`--relaxed-filter`
flag. Filtering happens in the served HTML — the toolbar writes `f=attr:value`
URL state that narrows the views.

## Usage

```sh
# Default — opens browser
reqmd serve 02-trace-your-spec/

# Terminal-only re-check loop (no HTTP server)
reqmd serve 02-trace-your-spec/ --headless

# With verification results — verdict badges on measure cards
reqmd serve 02-trace-your-spec/ --results 11-verification-results-ctrf/
reqmd serve 02-trace-your-spec/ --results 12-review-documentation/
```

## With verification results

Pass `--results` (repeatable) to load CTRF or manual verification results.
Verdict badges appear on measure cards (pass / fail / skipped / inconclusive).
Result files are watched alongside spec files — editing a CTRF JSON or
manual-results markdown triggers a rebuild after the debounce window and a
browser refresh.

```sh
reqmd serve 02-trace-your-spec/ --results 11-verification-results-ctrf/
```

The `failing-verdict` graph check runs on every rebuild: a `fail` verdict
on a measure produces an ERROR-level finding, turning the status line red.

## What's next

Now look at the status lifecycle that makes `SW-001` a draft — go to
[Step 5](../05-status-disposition/).