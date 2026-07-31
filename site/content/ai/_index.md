---
title: "AI & agents"
description: "Why reqmd's plain-markdown + JSON format is the right substrate for AI-assisted spec workflows, and how an agent drives the loop."
weight: 6
---

reqmd is designed for the agent era. Every artifact an agent needs — the spec, the schema, the trace graph, the validation result, the verdict — is a plain text file on disk or a JSON blob on stdout. There is no database to query, no REST API to authenticate against, no UI to drive. The agent can read, write, and validate specs as fast as it can read and write code.

This page is written for two audiences. **If you're building an agent**, you'll find the loop, the AGENTS.md recipe, and the patterns below. **If you're evaluating reqmd against UI-based ALM tools**, the comparison table at the end is the short version; the rest is the long version with receipts.

## The agent loop

A spec-writing agent drives a tight four-step loop. Each step is fast enough to sit in the same tool call budget as reading a file:

```
   ┌─────────────────────────────────────────────────────┐
   │  1. READ  cat spec/02-system/features.md           │
   │  2. EDIT  write the new SYS-007 heading + attr     │
   │  3. CHECK reqmd check spec/  (50 ms, exits 0/1/2)  │
   │  4. JSON  reqmd check --json spec/  (parse verdict)│
   └─────────────────────────────────────────────────────┘
                            ↓
                     iterate on ❌ findings
```

The check is the contract. The agent doesn't read the spec to "understand" it — it reads the JSON verdict, sees the failed checks, and fixes them. The same loop works for:

- Writing a new requirement (edit, check, fix `id-prefix mismatch` + `missing upstream`).
- Closing a trace gap (read `requires-trace-from`, write a downstream requirement, check again).
- Bumping a spec version (read all `~N` pins against the upstream, run `reqmd repin` to fix the `outdated` findings, or use `--relaxed-versions` for the ones that don't matter).
- Adding a verification result (read CTRF JSON or a manual-results file, run `check --results`, fix any `failing-verdict` or `missing-verdict`).

All four exit with one of three codes: `0` clean, `1` validation errors, `2` parse error. The agent branches on the code, reads the JSON, and knows what to do.

## Why plain text beats a database

UI-based ALM tools — Jama, Polarion, Siemens Teamcenter, IBM DOORS, codeBeamer — model requirements as rows in a relational schema. The agent doesn't see those rows; it sees a REST API and a web UI. Three things go wrong:

1. **The prose is hidden.** A requirement is `id`, `title`, `description`, `priority`, `status`, `parent_id`, and a foreign-key chain to a hundred other tables. The actual *content* of the requirement — the prose, the rationale, the body — is a single text column the agent has to fetch and re-format every time.
2. **The trace graph is opaque.** "Where does SYS-007 trace to?" requires either a UI session or a paginated REST call that returns IDs but not context. The agent has to follow references by hand, across multiple round-trips, to recover what reqmd makes a single 50 ms read.
3. **Review is impossible.** When the agent writes a new requirement through the API, the change shows up in the UI but the diff is unreadable. A reviewer looking at the agent's PR sees a JSON blob in the audit log, not a Markdown paragraph they can mark up in GitHub.

reqmd's model is the opposite. A requirement is a level-2 heading, a YAML `attr` block, and free-form prose. The agent edits text. The diff is text. The review is text. The check is a binary that returns text. The whole pipeline is in text.

## The format is the API

Every operation the agent needs maps to one CLI invocation. The full command surface:

| Need | Command | Output |
|---|---|---|
| Validate a spec | `reqmd check spec/` | text + exit code 0/1/2 |
| Get structured findings | `reqmd check --json spec/` | JSON |
| Validate with V&V results | `reqmd check spec/ --results ci-out/` | text + exit code |
| Validate single file | `reqmd check file.md -s schema.yaml` | text + exit code |
| Demote version-pin errors | `reqmd check spec/ --relaxed-versions` | text + exit code |
| List all requirements | `reqmd ls spec/` | table |
| Same, structured | `reqmd ls --json spec/` | JSON |
| Coverage stats | `reqmd stats spec/` | text |
| Same, structured | `reqmd stats --json spec/` | JSON |
| Scaffold a new doc dir | `reqmd init my-feat/ --id-prefix SYS` | scaffolded dir |
| Scaffold ASPICE template | `reqmd init my-feat/ --preset aspice` | scaffolded dir |
| Scaffold results dir | `reqmd init my-results/ --preset results --id-prefix VR` | scaffolded dir |
| Update version pins (dry-run) | `reqmd repin spec/` | change list |
| Apply version pins | `reqmd repin spec/ --yes` | applied changes |
| Promote unpinned refs | `reqmd repin spec/ --yes --promote-unpinned` | applied changes |
| Live HTML preview | `reqmd serve spec/` | HTTP + SSE |
| Export to HTML | `reqmd export html spec/ -o docs/` | HTML files |
| Export to CSV | `reqmd export csv spec/ -o csv/` | CSV files |
| Export to graph DB | `reqmd export graph spec/ -o g.lbug` (needs `-tags ladybug`) | LadybugDB |
| Compare two git tags | `reqmd baseline diff v1.0 v2.0` | text diff |
| Same, structured | `reqmd baseline diff v1.0 v2.0 --json` | JSON |

That's the entire surface. No API key, no auth, no rate limit, no SDK to install. The agent reads `reqmd --help` once and has the full vocabulary. See the [cheat sheet](/cheat-sheet/) for example outputs of every command.

## The schema is the contract

Each document directory has a `schema.yaml` that defines the required attributes and the trace topology. The agent must read it before adding a new requirement — rejected schemas are config errors. The key fields:

```yaml
# Standard JSON Schema 2020-12 fields:
$schema: "https://json-schema.org/draft/2020-12/schema"
$id: "my-requirements"          # relative URI, resolves against file path
title: "My Requirements"        # shown in HTML export and stats
type: object                    # always object (attributes of a requirement)
required: [priority]            # attributes that must be present
properties:
  priority:
    type: string
    enum: [Critical, High, Medium, Low]
additionalProperties: false     # reject attributes not in properties

# reqmd-specific extension:
x-reqmd:
  level: system-requirements    # V-model level for this document
  document-id: system           # short identifier for qualified refs
  id-prefix: SYS-              # prefix for requirement IDs (SYS-001, SYS-002)
  upstream:                     # valid upstream sources for trace links
    level: stakeholder-needs
    sources:
      - ../01-stakeholder/
  mandatory-disposition: true  # every requirement needs a disposition
  external: false               # external reference reqs (imported, not authored)
  additional-status-values: [review]  # extend [draft, approved] with custom values
  ignore-status: false           # all reqs count as coverage regardless of status
```

### Built-in attributes (recognized automatically)

| Attribute | Description |
|-----------|-------------|
| `trace` | Array of upstream IDs. Supports `~N` version pins (e.g. `SYS-001~3`). |
| `requires-trace-from` | Declares expected downstream coverage. |
| `status` | `draft` or `approved`. Only `approved` counts as coverage. |
| `disposition` | `implemented`, `deferred`, or `rejected`. Required when `mandatory-disposition: true`. |
| `version` | Integer version. Used by version-pin staleness checks. |
| `verify` | Verification method: `Test`, `Review`, `Inspection`, `Analysis`, `Demonstration`. |
| `reqmd-suppress` | Array of check names to suppress (e.g. `[version-pin, missing-verdict]`). |
| `external` | If `true`, this is an external reference requirement. |

See the [cheat sheet](/cheat-sheet/) for the full schema reference and every `x-reqmd` option.

## A copy-pasteable AGENTS.md

Drop this into the root of a spec repo. The agent reads it once per session and has everything it needs:

````markdown
# Spec workflow

This repo is a [reqmd](https://reqmd.dev) spec. The agent drives it as follows.

## What reqmd is

Single-binary Go CLI. Validates a directory of `.md` files against per-directory
JSON Schema (YAML) with bidirectional trace checks. Exit codes: 0 (clean),
1 (validation errors), 2 (parse error). See <https://reqmd.dev/llms.txt>
for a machine-readable index of the docs.

## Hard rules

1. **Read before write.** Never add a requirement without first reading
   the parent document and at least one downstream that traces to it.
2. **Edit, don't rewrite.** A change to a requirement is a heading-level
   edit, never a file rewrite. Preserve the prose, the rationale, the
   history.
3. **Run `reqmd check` after every edit.** It runs in 50 ms on this
   repo. If it exits non-zero, read the JSON output and fix the
   findings before continuing. Don't batch edits.
4. **Never touch `id` once it's set.** IDs are the trace target. If
   you need to change a requirement's identity, write a new one and
   deprecate the old with `disposition: rejected`.
5. **Use `requires-trace-from: []` to opt out of coverage**, not a
   missing `trace:`. A missing attribute is ambiguous; an explicit
   empty list is a contract.
6. **Check the schema before adding a new attribute.** Read
   `<doc>/schema.yaml` — if the attribute isn't in `properties`, it
   will be rejected (`additionalProperties: false`).

## The loop

```
  read   spec/<doc>/<file>.md          # the current state
  edit   write the change (heading + attr + prose)
  check  reqmd check spec/             # ~50 ms, exits 0/1/2
  json   reqmd check --json spec/      # parse findings, fix ❌
```

The check is the contract. Don't ship an edit that exits non-zero.

## All commands

| Need | Command |
|---|---|
| Validate the whole tree | `reqmd check spec/` |
| JSON for parsing | `reqmd check --json spec/` |
| Validate with V&V results | `reqmd check spec/ --results ci-artifacts/` |
| List all requirements | `reqmd ls spec/` |
| Coverage stats | `reqmd stats spec/` |
| Add a new doc dir | `reqmd init my-feat/ --id-prefix SYS` |
| Update version pins | `reqmd repin spec/ --yes` |
| Live HTML preview | `reqmd serve spec/` |
| Export to HTML | `reqmd export html spec/ -o docs/` |
| Export to CSV | `reqmd export csv spec/ -o csv/` |
| Compare two tags | `reqmd baseline diff v1.0 v2.0` |
| Find broken trace refs | `reqmd check spec/ \| grep "broken reference"` |

## Schema

Each document dir has a `schema.yaml`. Key `x-reqmd` fields:
- `id-prefix: SYS-` — prefix for requirement IDs
- `upstream.sources: [../01-stakeholder/]` — where trace links can point
- `mandatory-disposition: true` — every requirement needs a disposition
- `external: true` — imported reference requirements, not authored

Built-in attributes: `trace`, `requires-trace-from`, `status`, `disposition`,
`version`, `verify`, `reqmd-suppress`, `external`.

## Status & disposition

- `status: draft` — work in progress, doesn't satisfy coverage.
- `status: approved` — signed off, counts as coverage provider.
- `disposition: implemented` — actively developed.
- `disposition: deferred` — accepted, postponed; needs `disposition-reason`.
- `disposition: rejected` — not accepted; needs `disposition-reason`.

## Version pins

If a requirement depends on a specific upstream version, pin the trace:
`trace: [UP-001~3]`. When the upstream `version` is bumped, the
`version-pin` check fires. Use `reqmd repin spec/ --yes` to bulk-update
all outdated pins. Demote outdated pins with `--relaxed-versions`
if the change is non-substantive.

## What NOT to do

- Don't introduce a new field in `attr` without checking
  `<doc>/schema.yaml` first. Rejected schemas are config errors.
- Don't drop the YAML fence language (`attr`) — it's how the parser
  finds the attribute block.
- Don't add a `trace:` without confirming the target ID exists in
  some document. Use `reqmd ls spec/ | grep ID` to verify.
````

The file is plain markdown, lives in the repo, and is read by humans and agents. One source of truth.

## Patterns

Six patterns the agent can mix and match.

### 1. Spec-from-issue

A user files an issue. The agent reads it, drafts a stakeholder-level
requirement, writes it to the right document, checks, and posts a PR.
The PR carries the new requirement, the trace to the parent goal, and
the rationale — all in one diff.

```sh
reqmd check spec/                 # baseline: clean
$EDITOR spec/01-stakeholder/      # write the new goal
reqmd check spec/                 # expect: 0 new findings
reqmd check --json spec/          # post as a PR comment
```

### 2. Coverage repair

The CI check fails with a `requires-trace-from` warning. The agent
reads the JSON, finds the requirement with no downstream coverage, and
either adds a downstream (preferred) or annotates the upstream with
`requires-trace-from: []` (acceptable when there's a real reason).

### 3. Spec validation in the agent loop

The agent writes a spec for a feature it's about to implement. The
spec goes in a doc dir. The agent runs `reqmd check` after every
edit. By the time the implementation lands, the spec already passes.

```sh
$EDITOR spec/02-system/features.md   # draft the feature
reqmd check spec/                    # find issues
$EDITOR spec/02-system/features.md   # fix
reqmd check spec/                    # clean
```

### 4. Version-pin watch

The agent is asked to bump the version of an upstream requirement.
It runs `reqmd repin` to bulk-update all outdated pins:

```sh
reqmd repin spec/                    # dry-run: see what would change
reqmd repin spec/ --yes              # apply
reqmd repin spec/ --json             # machine-readable output
```

Or, to just report which downstreams will need re-verification without
fixing:

```sh
reqmd check spec/ --relaxed-versions | grep "version-pin"
```

### 5. Verification roll-up

The agent has a CTRF report from a CI run. It passes the path to
`reqmd check` and the spec is checked with the latest verdicts:

```sh
reqmd check spec/ --results ./ctrf.json
# exits non-zero if any measure's latest result is "fail"
reqmd check --json spec/ --results ./ctrf.json | jq '.documents[].requirements[] | select(.checks[]?.level == "ERROR")'
```

The `failing-verdict` and `missing-verdict` checks behave like the
trace checks: ERROR fails CI, WARNING is informational, and the JSON
output is shaped for parsing.

### 6. Closing the spec → code loop with `reqmd-import`

The agent has a function ready to ship but no ID in the spec for it.
Run the source-code extractor to generate per-package `IMP-...`
requirements, then re-check:

```sh
reqmd-import extract ./internal ./spec/03-software/imported
reqmd check spec/                              # pass
```

Each generated requirement carries `x-reqmd.source-file` and
`x-reqmd.source-line` provenance, so the agent can answer "where is
this requirement implemented?" by reading the file. The companion
[`reqmd-import`](/import/) page has the full reference; the takeaway
is that the spec and the code now share an ID space, and the same
`reqmd check` loop that catches broken upstream refs also catches
broken implementation refs.

## Why this is better than UI-based ALM tools

A spec row in a UI-based ALM is an object the agent can't see.
A spec paragraph in reqmd is a file the agent can read. That's the
whole difference, but the consequences compound:

| Concern | UI-based ALM (Jama, Polarion, Teamcenter, DOORS) | reqmd |
|---|---|---|
| Spec format | Rows in a relational schema, prose in a text column | Plain Markdown + YAML |
| Agent reads spec | REST GET, paginated, returns IDs not context | `cat spec/<doc>/<file>.md` |
| Agent writes spec | REST POST, returns a JSON audit log entry | `echo "## REQ-001" > spec/<doc>/<file>.md` |
| Trace graph | REST GET, foreign keys, paginated joins | `reqmd ls spec/` (one read) |
| Review of an agent's change | Replay the REST call in the UI; diff is invisible | Standard Git diff, in the PR, in Markdown |
| Spec + code in one PR | No — spec and code live in different systems | Yes — same repo, same PR, same CI |
| Cost of a typo | Server round-trip, possible transaction | `sed -i` in your editor |
| Cost of a schema change | Migration script + admin review | Edit `schema.yaml`, check, done |
| Verification roll-up | REST query, custom report, export to PDF | `reqmd check spec/ --results ctrf.json` |
| Baseline diff | UI tool, "compare two baselines" wizard | `reqmd baseline diff v1 v2` |
| Version pin staleness | Report generator + email | `reqmd check spec/` (built-in check) |
| Setup time | Days (license, install, integrate, train) | 30 seconds (`go install`) |
| Lock-in | Vendor-specific data model, hard to extract | Plain text, you already have it |

The last row is the one that matters. A spec written in a UI-based
ALM is, at the byte level, a vendor's view of your requirements. A
spec written in reqmd is, at the byte level, a Markdown file in a
Git repo. The agent doesn't have to choose between "use the tool"
and "own your data." It gets both.

## See also

- [Quickstart](/quickstart/) — a 10-minute tour of the loop.
- [Cheat sheet](/cheat-sheet/) — full command reference with example outputs and schema.yaml options.
- [Use cases](/use-cases/) — what reqmd is designed for, and what is out of scope.
- [<code>reqmd-import</code>](/import/) — the source-code traceability tool that closes the loop from code to spec.
- [<code>llms.txt</code>](/llms.txt) — a machine-readable index of this site for agents.
- [<code>llms-full.txt</code>](/llms-full.txt) — the full plain-text dump of every page.
- [FAQ](/faq/) — the questions an integrator asks.