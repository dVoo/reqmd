---
title: "Step 2c — Repin"
description: "Bulk-update version-pin (~N) trace references to match the upstream's current version with reqmd repin --yes."
weight: 6
---

Bulk-update version-pin (`~N`) trace references to match the upstream's
current version. The `reqmd repin` command proposes or applies the
changes for you.

## When to use

After bumping a spec version ([Version pins](/quickstart/02b-version-pins/))
you'll see a wave of `version-pin: outdated` findings. `repin` rewrites
the affected `trace:` lines in place so the pins catch up to the
upstream — without you having to hand-edit every requirement.

## Dry-run by default

```sh
# See what would change — no files are written
reqmd repin 02-trace-your-spec/
# /path/to/dn.md  DN-001  UP-001 → UP-001~3
# /path/to/dn.md  DN-002  UP-002 → UP-002~7
# 2 outdated, 0 unpinned, 0 predated across 1 files.
```

The default mode is dry-run. The output is one line per affected
requirement (`file  REQ-ID  bare-ref → ref~newPin`) plus a one-line
summary. Nothing is written to disk.

## Apply

```sh
# Apply without prompting
reqmd repin 02-trace-your-spec/ --yes

# Or let repin prompt you on a TTY
reqmd repin 02-trace-your-spec/
# /path/to/dn.md  DN-001  UP-001 → UP-001~3
# 1 outdated, 0 unpinned, 0 predated across 1 files.
# Apply 1 changes across 1 files? [y/N] y
# applied 1 changes across 1 files
```

`repin` rewrites only the ```` ```attr ```` blocks in the affected
files; surrounding Markdown prose is preserved verbatim. The change
is idempotent — a second `repin` run after the first reports no
changes.

## Promote unpinned refs (opt-in)

By default `repin` only fixes outdated *pinned* refs (`~N` lower than
the upstream's `version`). To also pin refs that have no `~N` against
a versioned upstream, pass `--promote-unpinned`:

```sh
reqmd repin 02-trace-your-spec/ --yes --promote-unpinned
```

This converts "no claim" into "claimed at current version". It is
opt-in because the semantics change: a previously-unpinned ref no
longer means "verified against HEAD every time" — it now claims
verification against a specific version.

## Predated findings

Refs whose `~N` is *ahead* of the upstream's `version`
(`direction: "predated"`) are surfaced in the change list but
**never auto-fixed**. They indicate a data integrity error (a
downstream claiming verification against a version that does not
yet exist on the upstream) and must be fixed manually — either by
bumping the upstream or by correcting the pin.

## Machine-readable output

```sh
reqmd repin 02-trace-your-spec/ --json
```

The JSON shape is stable:

```json
{
  "deltas": [
    {
      "kind": "outdated",
      "req_id": "DN-001",
      "file": "spec/03-software/dn.md",
      "target_id": "UP-001",
      "source_ref": "UP-001~1",
      "old_pin": 1,
      "new_pin": 3,
      "new_version": 3
    }
  ],
  "by_file": { "spec/03-software/dn.md": 1 },
  "outdated": 1,
  "unpinned": 0,
  "predated": 0
}
```

`kind` is one of `outdated`, `unpinned`, or `predated`. Combined with
`--yes`, the JSON output also includes an `apply` block summarising
the rewrite (files touched, deltas applied, deltas skipped).

## Suppression

The per-requirement `reqmd-suppress: [version-pin]` opt-out is
respected by `repin` — suppressed nodes never appear in the change
list.

## What's next

Compare requirement baselines between git tags — go to
[Baseline diff](/quickstart/03b-baseline-diff/).
