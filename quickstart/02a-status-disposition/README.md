# Step 2a — Status & Disposition

Two built-in attributes control how requirements behave in trace coverage
and how deferred work is tracked.

## Status lifecycle

Every requirement has a `status` — either `draft` or `approved` (default when
omitted). Only `approved` requirements count as upstream coverage providers.

```markdown
## REQ-001: System boot
```attr
status: approved
trace: [STK-001]
```
```

A `draft` requirement can be referenced by a `trace` but does **not** satisfy
a `requires-trace-from:` expectation. This lets you stage work-in-progress
without breaking coverage checks.

Try it: in [Step 2](../02-trace-your-spec/), the `SYS-001` requirement shows:

```
⚠  SYS-001  no upstream trace: no approved requirement from "software" traces to this item (1 draft downstreams ignored)
```

The `(1 draft downstreams ignored)` message means `SW-001` traces to `SYS-001`
but `SW-001` is still `draft` — so the coverage chain is incomplete.

## Disposition workflow

When a requirement can't be immediately implemented, use `disposition` +
`disposition-reason` instead of leaving silent gaps:

| Disposition | What it signals | Rationale needed? |
|---|---|---|
| `implemented` | Actively developed | No |
| `deferred` | Accepted, postponed | **Yes** |
| `rejected` | Not accepted | **Yes** |

```markdown
## STAKE-007
```attr
status: approved
disposition: deferred
disposition-reason: "Deferred to Phase 2 per steering committee 2025-03-14"
```
The system shall support over-the-air firmware updates for all ECUs.
```

Disposition is orthogonal to trace coverage. Use `requires-trace-from: []` to
explicitly state that no downstream coverage is expected.

## Extending the status enum

When `draft`/`approved` is too narrow, add custom values:

```yaml
x-reqmd:
  additional-status-values: [review, in-progress]
```

## Opting out of the lifecycle

For legacy imports or generated specs:

```yaml
x-reqmd:
  ignore-status: true
```

All requirements count as coverage providers; the status filter is hidden in
HTML export.

## What's next

Now export your spec tree — go to [Step 3](../03-export/).