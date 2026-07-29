# Step 2b — Version Pins

Pin trace references to a specific upstream version and detect when the
upstream changes.

## How version pins work

A downstream requirement can pin the upstream version it was last verified
against using the `~N` suffix:

```markdown
## UP-001
```attr
status: approved
version: 3
```
```

```markdown
## DN-001
```attr
status: approved
trace: [UP-001~2]   # verified against UP-001 v2, but UP-001 is now v3
```
```

When the upstream `version` is bumped, the pin becomes stale and reqmd flags
the downstream as needing re-verification.

## Findings

| Condition | Meaning | Level | Demote with |
|-----------|---------|-------|-------------|
| `pin < upstream.version` | Outdated — upstream bumped, downstream needs re-verify | **ERROR** | `--relaxed-versions` |
| `pin > upstream.version` | Predated — downstream claims a non-existent version | **ERROR** | (never — data integrity) |
| `pin == upstream.version` | Current | — | — |
| No pin | No version check applies | — | — |

## Try it

The spec tree in [Step 2](../02-trace-your-spec/) doesn't use version pins, so
no version-pin findings fire. To see it in action, add `version: 1` to an
upstream requirement and `~1` to a downstream trace, then bump the upstream
to `version: 2`:

```sh
reqmd check 02-trace-your-spec/
# ❌ DN-001  version-pin: outdated — UP-001 is at version 2, pinned to ~1

# Demote to WARNING (CI-friendly: fails on predated, warns on outdated)
reqmd check --relaxed-versions 02-trace-your-spec/
# ⚠ DN-001  version-pin: outdated — UP-001 is at version 2, pinned to ~1
```

## Suppression

Suppress the check per-requirement:

```yaml
reqmd-suppress:
  - version-pin
```

## Stale verdicts

Version pins also apply to result→measure traces: pin a result with
`MEASURE-ID~3` against a measure now at `version: 4` and the `outdated` finding
fires. This is how stale-verdict is detected — no new check, just the
existing version-pin check on a new edge type. See [Step 5](../05-verification-results-ctrf/).

## What's next

Now export your spec tree — go to [Step 3](../03-export/).