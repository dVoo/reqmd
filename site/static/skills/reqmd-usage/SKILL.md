---
name: reqmd-usage
description: Operate an existing reqmd Markdown requirements tree. Use to validate a spec with reqmd check, interpret exit codes and JSON findings, trace a requirement, repair broken trace refs, fix missing coverage (requires-trace-from), manage draft/approved status and disposition, detect and bulk-update stale version pins with reqmd repin, scope checks with --filter, and roll up CTRF or manual verification results. Triggers include "reqmd check", "validate the spec", "trace", "coverage", "version pin", "repin", "failing-verdict", and wiring reqmd into CI.
---

# Using reqmd

Operational guide for an existing reqmd tree. For *writing* documents and
schemas, use the `reqmd-authoring` skill instead.

## When to use

- Validate a spec tree, or gate a PR/CI job on it.
- Add or repair a `trace:` link between requirements.
- Explain or fix a `requires-trace-from` coverage warning.
- Check whether a requirement is traced, approved, deferred, or verified.
- Find and update stale `~N` version pins.
- Load CTRF or manual verification results and act on verdicts.
- Slice a multi-variant spec with `--filter`.

## Ground rules

1. **`reqmd check` is the contract.** Run it after every change. Branch on the
   exit code, then read findings.
2. **Exit codes:** `0` = no ERROR, `1` = ERROR-level validation/trace failure,
   `2` = schema or file parse error. WARNING and INFO never change the exit code.
3. **Exit 0 does not mean trace-clean.** Broken references, missing coverage,
   untraced requirements, and `missing-verdict` are WARNING-level. Always inspect them
   (grep for `⚠` or use `--json`); do not report "spec is clean" from the exit
   code alone.
4. **Never change an existing requirement `id`.** IDs are trace targets.
   Deprecate instead: set `disposition: rejected` with a reason.
5. **`--results` is ephemeral.** Results are loaded per invocation and never
   written to the spec repo.
6. **Dry-run first.** `reqmd repin` mutates files only with `--yes`.
7. The binary is `reqmd`. The repo may be named something else.

## Command surface

| Need | Command |
|---|---|
| Validate whole tree | `reqmd check <root>` |
| Machine-readable findings | `reqmd check --json <root>` |
| Validate one file | `reqmd check <file.md> -s <schema.yaml>` |
| Check with V&V results | `reqmd check <root> --results <path> [--results <path>...]` |
| Demote outdated pins | `reqmd check <root> --relaxed-versions` |
| Scope to a subset | `reqmd check <root> --filter '<expr>'` |
| Cross-configuration trace check | `reqmd check <root> --disjoint-check <attr>` |
| List all content | `reqmd ls <root>` / `reqmd ls --json <root>` |
| Attribute breakdown | `reqmd stats <root>` / `reqmd stats --json <root>` |
| Dry-run pin updates | `reqmd repin <root>` |
| Apply pin updates | `reqmd repin <root> --yes` |
| Compare two baselines | `reqmd baseline diff <tag1> <tag2> [--json]` |
| Live preview while editing | `reqmd serve <root>` (`--headless` for terminal only) |
| Export | `reqmd export csv\|html <root> -o <dir>` (`--results` for verdicts) |
| Warehouse tables | `reqmd export json <root> -o <dir>` (`--results` for verdicts) |

## Procedure: the everyday loop

1. **Orient.** `reqmd ls <root>` lists every node in document order
   (`req` / `container` / `info` types) with attributes. `reqmd stats <root>`
   shows the attribute-value breakdown per document.
2. **Baseline.** `reqmd check <root>` — note existing warnings before you
   change anything, so you only fix what you touched.
3. **Edit** the requirement's heading or `attr` block (see `reqmd-authoring`).
4. **Re-check.** `reqmd check <root>`. Fix ERROR-level findings before continuing. Then scan
   warnings introduced by your edit.
5. **Report** using `--json` when a machine consumes the result.

For CI, run `reqmd check <root> --json` and gate on the exit code; the JSON
also carries `exit_code` and `summary.warnings`.

## Tracing

`trace:` points **downstream → upstream** (a requirement lists its parents).
The reverse direction is coverage (`requires-trace-from`).

To add or repair a trace:

1. **Confirm the target exists.** `reqmd ls <root> | grep '<TARGET-ID>'` or
   `reqmd ls --json <root> | jq -r '.documents[].rows[] | select(.id=="<TARGET-ID>")'`.
   A reference to a missing ID is a WARNING (`broken-ref`), not an error.
2. **Choose the reference form.**
   - `trace: [SYS-001]` — unqualified, fine when the ID is globally unique.
   - `trace: [system/SYS-001]` — qualified with the target document's
     `x-reqmd.document-id`. Prefer this when prefixes may collide, when
     documents move, or when assembling git submodules.
   - `trace: [SYS-001~3]` — pin the upstream version you verified against.
3. **Edit only the `attr` block** of the downstream requirement.
4. **Re-check.** An unqualified ref that matches IDs in two documents is an
   ERROR (`ambiguous reference`) — qualify it.

Related checks:

| Finding | Level | Fix |
|---|---|---|
| `broken reference` | WARNING | Correct the ID, or add the missing upstream requirement. |
| `ambiguous reference` | ERROR | Qualify as `document-id/ID`. |
| `circular dependency` | ERROR | Break the cycle; the trace graph must be a DAG. |
| `no upstream reference` (generic) | WARNING | Add `trace:`, or opt out with `requires-trace-from: []`. |

## Coverage (`requires-trace-from`)

Coverage asks: *does an approved requirement at each expected downstream
level trace to this one?* Declare expectations on the requirement, or once per
document via `x-reqmd.requires-trace-from` in `schema.yaml`.

- Tokens are `document-id`s or `level`s.
- **Only `status: approved` providers count.** A `draft` inbound is reported as
  `(N draft downstreams ignored)` and does not close coverage.
- `requires-trace-from: []` opts a requirement (or an empty document default,
  the whole document) out of downstream coverage.
- The declared attribute replaces the document default entirely — no merging.
- In an `external: true` document, untraced warnings are suppressed and
  version pins are out of scope.

When a coverage warning is legitimate, prefer adding the downstream
requirement. `requires-trace-from: []` is a contract, not a silencer — use it
only when no downstream coverage is genuinely expected.

## Status and disposition

- `status: draft | approved`; omitted means `approved`. Only `approved` counts
  as coverage.
- `disposition: implemented | deferred | rejected`; `deferred`/`rejected`
  require `disposition-reason` (missing = WARNING).
- When `x-reqmd.mandatory-disposition: true`, a missing `disposition` is an
  ERROR.
- Documents may extend status via `x-reqmd.additional-status-values` or opt
  out via `x-reqmd.ignore-status: true`.

## Version pins

A pin (`UP-001~2`) records the upstream version a downstream was verified
against. When `UP-001` bumps to `version: 3`, the pin is stale.

| Condition | Level | Action |
|---|---|---|
| `pin < upstream.version` (outdated) | ERROR | `reqmd repin <root> --yes`, or `--relaxed-versions` to demote to WARNING |
| `pin > upstream.version` (predated) | ERROR | Never auto-fixed — fix the pin or the version by hand |
| No pin, or upstream `version` unset/`0` | — | No check |
| Upstream `external: true` | — | Out of scope |

Workflow:

```sh
reqmd repin <root>                 # dry-run: list proposed changes
reqmd repin <root> --yes           # apply
reqmd repin <root> --json          # machine-readable change list
```

`repin` rewrites pins to the upstream's current version, touching only `attr`
blocks. It is idempotent. `--promote-unpinned` additionally pins refs that have
no `~N` at all — opt-in, because it converts "no claim" into a claim.
Per-requirement suppression: `reqmd-suppress: [version-pin]`.

## Verification results

`--results` loads CTRF reports (automated tests) and/or manual-results
directories (review/inspection/analysis) and runs outcome-gated checks.
Repeatable; each path is auto-detected.

```sh
reqmd check <root> --results ci-out/ctrf.json --results reviews/
reqmd check <root> --results ci-out/ --ignore-unbound-results
```

| Finding | Level | Meaning |
|---|---|---|
| `missing-verdict` | WARNING | An approved measure (has `verify:`) has no result, and none of its approved downstream cases do either. |
| `failing-verdict` | ERROR | The measure's rolled-up verdict is `fail`. |
| `unbound-result` | WARNING | A result declares `x-reqmd` metadata but binds to nothing. |

Verdicts roll up strictly: any `fail` → fail, else `inconclusive`, else
`skipped`, else `pass`. Draft measures and deferred/rejected measures are
skipped. Version pins also apply to result→measure traces, so a stale result
shows up as an ordinary `version-pin` finding. Result-attributed findings are
reported under `results` in JSON.

## Filtering and variants

`--filter '<expr>'` (expr-lang) is available on `check`, `ls`, `stats`,
`export csv|html`, `serve`, and `baseline diff`.

```sh
reqmd check <root> --filter '"Premium" in variant'
reqmd check <root> --filter 'variant == nil or "Premium" in variant'
reqmd check <root> --filter 'status == "approved"'
```

- A filter naming an attribute that exists in no `schema.yaml` is a compile-time
  ERROR — this is a typo guard, not a silent no-match.
- **Coverage is filter-aware:** a filtered-out requirement can neither require
  nor provide coverage, so per-variant checks don't produce false failures.
- Filter expressions can read built-ins `id`, `title`, `status`, `disposition`,
  `trace`, `version` in addition to schema attributes.
- To reject trace links between incompatible configurations, use
  `--disjoint-check <attr>` (or `x-reqmd.disjoint-check`): every trace link must
  share at least one value for the array attribute; empty/absent means "applies
  to all". Zero intersection is an ERROR.

## Reading `--json`

Top level: `documents[]`, `parse_errors[]`, `results[]`, `summary`, `version`,
`exit_code`. Each document has `path`, `requirements[]` (`id`, `valid`,
`checks[]`), and `req_count`. Each check has `level`, `message`, and often
`code`, `direction`, `source`.

```sh
# all ERROR findings, per requirement
reqmd check --json <root> \
  | jq -r '.documents[].requirements[] | .id as $id
      | .checks[]? | select(.level=="ERROR") | "\($id): \(.message)"'

# any findings at all (warnings included)
reqmd check --json <root> \
  | jq -r '[.documents[].requirements[].checks[]?] | length'

# verification-result findings
reqmd check --json <root> --results ci-out/ | jq -r '.results[]? | "\(.id): \(.message)"'
```

## Exports and preview

- `reqmd export html <root> -o out/` — standalone per-document HTML with trace
  links, doc chain, sidebar TOC, search, theme toggle.
- `reqmd export csv <root> -o out/` — one CSV per document, document order.
- Add `--results` to either for verdict badges/columns.
- `reqmd export json <root> -o out/` — four flat NDJSON tables for reporting
  platforms (Databricks, Snowflake, BigQuery): `requirements.ndjson` (one row
  per content node; built-ins promoted to top-level keys, other attributes in
  `attrs`), `traces.ndjson` (one row per declared trace ref, with `pin` and
  `resolved`), `results.ndjson` (one row per verdict; needs `--results`), and
  `findings.ndjson` (one row per check finding). Output is a pure function of
  tree + results, so repeated exports are byte-identical; a dangling ref appears
  in `traces` with `resolved:false` and in `findings`.
- `reqmd serve <root>` — watch + live-reload preview; `--headless` for a
  terminal-only rebuild loop; `--filter` to preview one configuration.
- `reqmd baseline diff <tag1> <tag2>` — semantic diff of requirements, schemas,
  and submodule pins. It always exits 0; it is informational.

## Suppression

Per-requirement `reqmd-suppress: [<check>]` accepts:
`broken-ref`, `circular`, `untraced`, `no-downstream`, `disposition-reason`,
`mandatory-disposition`, `id-prefix`, `version-pin`,
`requires-trace-from-coverage`, `missing-verdict`, `failing-verdict`,
`disjoint-attribute`.

Suppress sparingly and locally. Prefer fixing the data or declaring an explicit
`requires-trace-from: []` over silencing a check the whole tree needs.

## Pitfalls

- **Exit 0 with warnings.** The single most common mistake. Broken refs and
  missing coverage pass CI. Check warnings explicitly.
- **Draft upstreams.** A `draft` inbound does not satisfy coverage even though
  the `trace:` is valid.
- **Predated pins** are data-integrity errors and are never auto-fixed.
- **Unqualified refs across submodules** are ambiguous; qualify with
  `document-id/ID`.
- **`h1` is the document title**, never a requirement — don't look for it in
  `reqmd ls` requirements or trace to it.
- **`--filter` scopes coverage too** — not just reporting.
- Editing `external: true` or generated/imported documents by hand is usually
  wrong; regenerate them.

## Verification

- `reqmd check <root>` exits 0 and `--json` has `exit_code: 0` with zero
  ERROR-level checks.
- No new WARNING was introduced by the change (compare against the baseline).
- Any trace you added resolves: `reqmd ls <root> | grep '<TARGET-ID>'`.
- If you ran `repin`, a second `reqmd repin <root>` reports "no version-pin
  changes needed".
