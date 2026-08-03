---
title: "FAQ"
description: "Frequently asked questions about reqmd — what it's for, how it works, and how it fits into your workflow."
weight: 6
---

Questions about reqmd, grouped by topic: what it is, how to use it, and how it fits your workflow.

## What is reqmd?


{{< details summary="What does reqmd do?" open="true" card="true" >}}
reqmd is a command-line tool for writing, linking, and checking requirements and verification measures in plain Markdown. You write each requirement as a heading, a small YAML block, and some prose. reqmd checks every cross-reference — broken links, missing coverage, circular dependencies, stale version pins — and tells you what's wrong. It also exports to HTML, CSV, and graph formats.

The key idea: your spec lives in text files, in Git, next to your code (or in its own repo). The validation runs in milliseconds. No database, no server, no lock-in.
{{< /details >}}

{{< details summary="What can I use reqmd for?" card="true" >}}
Any project where you need to write requirements and trace them to each other, to tests, and to reviews:

- **Spec-first development** — write the requirement, then the test, then the code, all in one pull request.
- **Higher-level specification** — stakeholder needs, system requirements, software design, and test specs traced across a V-model. The spec does *not* need to live next to the code; it can be a standalone repo for requirements and traceability at any level.
- **Verification and validation tracking** — load test results from CI (CTRF JSON) and manual review results (Markdown), and get verdict badges showing which measures pass, fail, or are still missing.
- **CI-gated traceability** — every pull request that touches a spec runs `reqmd check` and gets a pass/fail answer.
- **Live review** — `reqmd serve` gives you a live-reloading HTML preview while you edit.
{{< /details >}}

{{< details summary="Does reqmd only handle requirements, or also verification measures?" card="true" >}}
Both. reqmd treats verification measures — tests, reviews, inspections, analyses, demonstrations — as first-class objects. A test specification is a requirement with a `verify` attribute. A test result is loaded as a pseudo-requirement with an `outcome` (pass/fail/skipped/inconclusive) that traces back to the measure it verifies.

This means the full traceability chain is: **stakeholder need → system requirement → software requirement → test specification → test result**. Every link is checked. If an approved measure has no result, reqmd warns (`missing-verdict`). If the latest result is a fail, reqmd errors (`failing-verdict`).
{{< /details >}}

{{< details summary="Does the spec need to live next to the code?" card="true" >}}
No. reqmd works on any directory of Markdown files. Many teams keep the spec in the same Git repository as the code so that spec changes and code changes land in the same pull request. But a standalone requirements repository works just as well — point `reqmd check` at whatever directory contains your spec tree.

For higher-level requirement specification and traceability (stakeholder needs, system architecture, supplier specs), a separate repo is often the better choice. The spec is reviewed by people who don't work in the codebase — product managers, QA, safety engineers — and a dedicated repo with its own review cycle keeps things clean.
{{< /details >}}

{{< details summary="How is reqmd different from a wiki or a Google Doc?" card="true" >}}
A wiki stores text. reqmd stores text *and checks it*. In a wiki, a broken trace link is a human problem — someone has to notice it. In reqmd, a broken link is a CI failure — the tool catches it on every commit. You also get structured export (HTML with trace navigation, CSV, graph), version pinning, status lifecycle, and verification result tracking. None of that exists in a wiki.
{{< /details >}}

{{< details summary="How is reqmd different from OpenFastTrace?" card="true" >}}
Both are static-spec tools with a strong emphasis on bidirectional traceability. The differences:

- **Data format.** reqmd is Markdown-only with a YAML block; OFT uses its own `.xml` formats.
- **Runtime.** reqmd is a single Go binary; OFT is a Java project.
- **Verification.** reqmd has first-class CTRF ingestion and outcome-gated checks (`missing-verdict`, `failing-verdict`); OFT treats verification as separate artifacts.
- **Fit.** OFT is a stronger fit for traditional spec-by-spec coverage analysis; reqmd is a stronger fit for git-native, text-first, CI-driven workflows.

Both are good tools. Pick by your team's existing workflow.
{{< /details >}}

{{< details summary="How is reqmd different from Jama / Polarion / DOORS?" card="true" >}}
Those are database-backed requirements-management suites. They've been the default in regulated industries for 20 years, and they solve a real problem — managed workflows, audit trails, role-based access. But they come with structural problems that get worse as your team gets more agile:

- **Slow.** Every interaction is a web UI round-trip. A spec review that takes 10 minutes in a Git diff takes an hour in DOORS.
- **Disconnected from the code.** The spec is in a database; the code is in Git. There is no native connection. When a developer renames a function, the spec doesn't know.
- **No real branching.** "Baselines" are snapshots, not branches. You can't `git checkout -b`, try a change, and merge it. Two teams working on the spec simultaneously means lock-then-edit.
- **No submodules.** A multi-team spec is one monolithic database. You can't split it into per-team repos and compose them.
- **Not CI-native.** Running "does the spec validate?" in CI requires a REST API call, authentication, and a custom script. With reqmd it's `reqmd check spec/` — one binary, 16 ms, exit code 0/1/2.
- **Not developer-friendly.** Developers don't open DOORS. They open their editor. If the spec is in a database, the developer never reads it, and the spec drifts from the code.
- **Vendor lock-in.** Your spec is in a proprietary database schema. Exporting it is a project. With reqmd, your data is plain text — diffable, portable, yours.

**Pick reqmd** when you want the spec to live where the developers work — in Git, next to the code, with branching, submodules, CI, and pull-request review as the natural workflow. **Pick an RM suite** when you specifically need named approvers with electronic signatures and a regulated audit trail that an auditor reads from the tool. Some teams use both: reqmd for the engineering spec, the RM suite as the system of record. See the [compare page](/compare/) for the full side-by-side.
{{< /details >}}

## How it works


{{< details summary="How does the GitHub review workflow work?" card="true" >}}
reqmd is designed for Git-based review. The spec lives in a Git repository (either next to the code or in a dedicated repo). Every change to a requirement, a trace link, a schema, or a verification result is a commit. Here's the typical workflow:

1. **Author.** Open a branch. Edit the `.md` files — add a requirement, update a trace link, change a status. Commit and push.
2. **Check locally.** Run `reqmd check spec/` before pushing. If it exits non-zero, fix the issues.
3. **Open a pull request.** The diff shows exactly what changed: the requirement text, the trace link, the schema. Reviewers see the same diff they'd see for a code change.
4. **CI runs the check.** A GitHub Action (or GitLab CI job) runs `reqmd check --json spec/` on the PR. If it fails, the PR is blocked. The JSON output can be posted as a comment or an annotation.
5. **Review.** Reviewers read the diff, check the trace links, run `reqmd serve spec/` for a live preview if they want to see the rendered spec with links resolved.
6. **Merge.** Once the check passes and the review is approved, merge. The spec is now consistent — no broken links, no missing coverage, no stale version pins.

The key advantage over a database-backed RM tool: the review happens in the same pull request interface the team already uses for code. The diff is the review artifact. The check is the gate. No separate tool, no separate login, no separate review state to track.
{{< /details >}}

{{< details summary="How does the trace check work?" card="true" >}}
reqmd builds a trace graph from every `trace:` link in every requirement, then checks it in one pass:

- **Broken references** — a `trace:` pointing to an ID that doesn't exist anywhere.
- **Missing coverage** — a requirement with `requires-trace-from` that no downstream requirement traces to.
- **Circular dependencies** — a chain of traces that loops back to itself.
- **Version pin staleness** — a downstream pinned to `SYS-001~3` but the upstream is now version 5.
- **Missing verdicts** — an approved measure with no verification result.
- **Failing verdicts** — the latest result for a measure is `fail`.

Broken references, cycles, and failing verdicts are errors (exit code 1). Missing coverage and missing verdicts are warnings (exit code 0, but shown in the output).
{{< /details >}}

{{< details summary="How does the status lifecycle work?" card="true" >}}
Every requirement has a `status`: `draft` (work in progress) or `approved` (signed off). Only `approved` requirements count as coverage providers — a `draft` requirement can be referenced by a trace but does *not* satisfy a `requires-trace-from` expectation.

Requirements can also carry a `disposition`: `implemented` (actively developed), `deferred` (parked, with a reason), or `rejected` (not doing it, with a reason). When a document directory sets `mandatory-disposition: true`, every requirement must have a disposition.
{{< /details >}}

{{< details summary="How do I load verification results?" card="true" >}}
Two formats, both loaded with the `--results` flag (repeatable):

- **CTRF JSON** — from automated test runs. The `x-reqmd.id` field in each test maps it to a requirement ID. reqmd reads the pass/fail/skipped verdict and the timestamp.

reqmd synthesizes one pseudo-requirement per result and runs two new checks: `missing-verdict` (approved measure, no result) and `failing-verdict` (latest result is fail). The HTML export shows color-coded verdict badges on every measure card.
{{< /details >}}

{{< details summary="Is there a web UI?" card="true" >}}
There is no separate web UI; the HTML export *is* the web UI. Run `reqmd serve spec/` for a live-reloading preview, or `reqmd export html spec/ -o docs/` and host the resulting `docs/` directory as a static site.

The HTML export includes card-based layout, upstream/downstream trace links, a document chain tab strip for V-model navigation, status filters, theme toggle, and (with `--results`) color-coded verdict badges on measure cards.
{{< /details >}}

{{< details summary="Does reqmd work without Git?" card="true" >}}
Yes, for everything except `baseline diff`. The spec tree is plain files; no Git required to validate, list, or export. `baseline diff` is the one command that calls `git archive` to read a tagged snapshot — every other command operates on whatever directory you point it at.
{{< /details >}}

{{< details summary="How do I use git submodules for configuration management?" card="true" >}}
Configuration management — the discipline of controlling, versioning, and composing a multi-team spec across component boundaries — is a core systems engineering practice. In reqmd, it's built on Git submodules. Each team owns their own requirements repo as a git submodule, and a top-level repo assembles them into a single tree that `reqmd check` validates as a whole.

A typical layout:

```
spec-monorepo/
  stakeholder/        ← submodule: git@github.com:team/stakeholder-spec.git
  system/             ← submodule: git@github.com:team/system-spec.git
  software/           ← submodule: git@github.com:team/software-spec.git
  tests/              ← submodule: git@github.com:team/test-spec.git
```

Each submodule is a normal reqmd document directory with its own `schema.yaml`. The top-level repo just holds the submodule references — no spec content of its own. When a team updates their spec, they push to their repo and the top-level repo bumps the submodule pin.

**Cross-repo trace links** work out of the box. A requirement in `system/` can trace to a requirement in `stakeholder/` by ID — reqmd resolves it across the submodule boundary because all submodules are on disk under the top-level directory. Use qualified references (`document-id/ID`) if two submodules share ID prefixes.

**The baseline is the submodule pins.** When you need to compare two baselines, `reqmd baseline diff v1.0 v2.0` reads the submodule pins at each tag and reports what changed — added, removed, and modified requirements across all submodules, plus which submodule pins were updated. No separate configuration management tool needed; git is the tool.

**Review workflow** works the same as a single repo. A PR that bumps a submodule pin shows the exact diff of the submodule's spec changes. Reviewers see what changed, CI runs `reqmd check` on the assembled tree, and the merge updates the baseline.
{{< /details >}}

{{< details summary="How do I avoid hard-coded parent paths in submodule schemas?" card="true" >}}
Use a stable `x-reqmd.document-id` for every independently maintained document and omit `upstream.sources` when the parent path is chosen by the integrating repository. The document ID stays stable even when a submodule is mounted at a different directory.

```
# system/schema.yaml
x-reqmd:
  document-id: system
  level: system-requirements
  upstream:
    level: stakeholder-needs
    # sources intentionally omitted
```

Reference requirements by document ID in Markdown attributes:

```
trace:
  - stakeholder/STK-001

requires-trace-from: [software]
```

Run `reqmd check` from the assembled root containing all submodules. ReqMD discovers every schema below that root and resolves `document-id/requirement-id` references. A submodule validated by itself cannot resolve references to documents outside its checkout.

`upstream.sources` remains available for HTML document-chain navigation and path-based boundary inference; add it only when the assembled repository has a stable layout. See the [`document-id` quickstart example](/quickstart/10-submodule-configuration/).
{{< /details >}}

{{< details summary="How do I link requirements down to the actual code?" card="true" >}}
Use the companion [`reqmd-import`](/import/) tool. It walks your Go or Python source code with a tree-sitter grammar and writes per-package `.md` requirements into your spec tree. Every function, type, and method gets a requirement ID in the same namespace as your spec.

The extracted code requirements are marked `external: true` and live in an `imported/` sub-tree. They don't need upstream traces (they're not authored by hand), but they *do* trace to the real implementation. The result: the same `reqmd check` that catches a broken spec reference also catches a broken implementation reference — if a function is renamed or deleted, the trace from the spec to that function becomes a broken reference.

The full traceability chain becomes: **stakeholder need → system requirement → software requirement → test specification → test result → code implementation**. Every link is a plain-text reference that reqmd validates mechanically.

`reqmd-import` is a separate Go module with its own release cadence. It supports Go and Python today; Rust and Zig are reserved for future work.
{{< /details >}}
 

## Schema & data


{{< details summary="What is the schema.yaml for?" card="true" >}}
Each document directory has a `schema.yaml` that declares the attributes a requirement in that directory carries — their names, types, and which are required. It's a standard [JSON Schema 2020-12](https://json-schema.org/draft/2020-12/schema) document written in YAML, so you get validation, enums, and type checking for free. The `x-reqmd` extension block adds reqmd-specific options: the document's `level` in the V-model, its `id-prefix`, the `upstream` level it traces from, and check toggles like `disjoint-check`.

One schema per directory means each team or V-model layer can define its own attributes independently. See the [cheat sheet](/cheat-sheet/) for the full `x-reqmd` reference and a complete example.
{{< /details >}}

{{< details summary="Can reqmd import or export Excel, Word, or ReqIF?" card="true" >}}
Not built in (yet). The data model — plain Markdown with a YAML `attr` block per requirement — is simple enough that a one-time script can convert from or to any of these formats. For imports, write a script that walks the source file and emits one `.md` per requirement; for exports, `reqmd ls --json` gives you a flat JSON list you can reshape into a spreadsheet or document.

The companion [`reqmd-import`](/import/) tool is the closest existing roundtrip — it extracts requirement IDs from Go and Python source via tree-sitter. A future `reqmd-import reqif` subcommand is the most likely path for native ReqIF support.
{{< /details >}}

{{< details summary="Does reqmd work with my issue tracker (Jira, GitHub Issues, Linear)?" card="true" >}}
No native integration. The common pattern is to put the issue-tracker key in a custom attribute (e.g. `jira: PROJ-123`) and search via `reqmd ls --json` or `reqmd stats --json`. The issue tracker is a downstream consumer of the spec, not a source.
{{< /details >}}


## Variants & configuration


{{< details summary="Can I manage multiple product variants in one spec tree?" card="true" >}}
Yes. reqmd deliberately has no built-in "variants" feature — you declare a `variant` array attribute in `schema.yaml` like any other custom attribute, tag each requirement, and scope every command with `--filter "<expr>"`. One spec tree serves any number of configurations (Base, Premium, Sport — or platforms, regions, owners) with no extra files, branches, or vocabulary.
{{< /details >}}

{{< details summary="How do I include requirements that apply to all variants?" card="true" >}}
A requirement with no `variant` attribute is treated as "common to all" — but the bare filter `"X" in variant` matches only requirements explicitly tagged X. Use the `or variant == nil` pattern to build the whole configuration view:

```sh
reqmd check spec/ --filter '"Premium" in variant or variant == nil'
```

This matches Premium-specific requirements plus the common-to-all ones. Use bare `"X" in variant` when you want only that configuration's specific requirements.
{{< /details >}}

{{< details summary="Is coverage checking variant-aware?" card="true" >}}
Yes — when `--filter` is active, coverage (`requires-trace-from`) is evaluated within the filtered subset only. A requirement excluded by the filter can neither require nor provide coverage. This prevents false failures: a Base-only requirement isn't reported under-covered just because its only provider is a Premium test that was never meant to cover it. Without `--filter`, coverage is computed across the whole tree exactly as before.
{{< /details >}}

{{< details summary="How do I catch a trace link between incompatible variants?" card="true" >}}
Use `--disjoint-check <attr>` — generalized, not variant-specific:

```sh
reqmd check spec/ --disjoint-check variant
```

For every trace link it verifies the two requirements share at least one value of the named array attribute. Zero intersection is an ERROR, e.g. a `variant: [Sport]` requirement tracing to a `variant: [Base]` target — no real configuration contains both. Requirements with an empty or absent value are exempt ("applies to all"). Declare it once in `schema.yaml` via `x-reqmd.disjoint-check: variant` to enforce it on every check.
{{< /details >}}

{{< details summary="How do I produce a &quot;what Premium adds over Base&quot; report?" card="true" >}}
`baseline diff` has a same-commit two-view mode that compares two filtered views of the same tree — no divergent branches or tags needed:

```sh
reqmd baseline diff \
  --filter-a 'variant == nil or "Base" in variant' \
  --filter-b '"Premium" in variant' \
  HEAD
```

The output is a real semantic diff (added / removed / modified with attribute-level detail) — the configuration-management evidence artifact, generated from a single commit. You can also filter a normal two-tag diff: `reqmd baseline diff v1.0.0 v1.1.0 --filter '"Premium" in variant'`.
{{< /details >}}

{{< details summary="Can I validate every variant in CI?" card="true" >}}
Yes — run one `reqmd check` per configuration in a CI matrix. Each job validates its view with `--filter "${{ matrix.filter }}" --json`; the JSON summary records the exact filter in a `"filter"` field. Add a `--disjoint-check variant` job to catch cross-configuration trace mistakes. A complete example is in [Step 14 — Variant management](/quickstart/14-variant-management/).
{{< /details >}}

{{< details summary="What does the filter expression language support?" card="true" >}}
`--filter` uses `expr-lang`, evaluated against each requirement's attribute map. Supported subset:

- `==`, `!=` — equality, e.g. `status == "approved"`
- `in` — array membership, e.g. `"Premium" in variant`
- `and`, `or`, `not` — boolean composition
- `contains`, `startsWith`, `endsWith` — string helpers, e.g. `id startsWith "SYS-"`
- `variant == nil` — attribute absence / "common to all"

Built-in variables are always available: `id`, `title`, `status`, `disposition`, `trace`, `version`. A filter referencing an attribute not declared in any `schema.yaml` is a compile-time ERROR — reqmd fails fast on typos instead of silently returning no matches.
{{< /details >}}

{{< details summary="Do I need separate branches for each variant?" card="true" >}}
No — that's the point. Variant differences live in the `variant` attribute of individual requirements, not in forked spec trees. The filter creates the views; `baseline diff --filter-a/--filter-b` creates the "variant A vs variant B" evidence; the CI matrix validates each view. One shared tree, one review process, no divergent branches to reconcile.
{{< /details >}}

{{< details summary="How do I choose between variant attributes, branches, submodules, and forks?" card="true" >}}
Three axes of divergence, one decision flow:

```
Will the two things reconcile (merge or die)?
  yes → branch
  no → Do they release and review independently
       (separate cadence, no shared check)?
    yes → fork (separate repo)
    no  → variant attribute (one tree, one check, one review)
```

The primary discriminator is the **independence test**: if the two things release and review independently (separate cadence, no shared `check`), they're a fork; if they must ship together from one commit and share one review, they're variants. The reqmd-native corollary is the **validation test**: if they must pass `check` against each other in one tree, they're variants. **Submodules** are orthogonal — an ownership/assembly mechanism that composes with all three, not a fourth axis.

For the full decision guide, a life-like topology showing all four mechanisms at once, and the fork-vs-submodule comparison, see the [Structure page](/structure/).
{{< /details >}}

{{< details summary="When should I use a branch, and when is it the wrong tool?" card="true" >}}
A branch is for change-in-progress that reconciles — it merges back or is discarded. It's the wrong tool when the difference is permanent: long-lived branches for variants lose single-tree validation and review (see [Do I need separate branches for each variant?](/faq/#do-i-need-separate-branches-for-each-variant)), and long-lived branches that are really a separate product bit-rot and can't release independently (that's a fork). The one-line rule: **if your branch will never merge and never die, it isn't a branch.**

See [Branches](/structure/#branches) on the Structure page.
{{< /details >}}

{{< details summary="When should I fork vs depend on a library at a pinned version?" card="true" >}}
Discriminate by whether you intend to *modify* the library's requirements or *depend on* them at a version. **Fork** if you own and evolve a copy — you take the library's requirements as your starting point and change them, can track upstream (host fork or raw), and pay merge-conflict tax on local edits. **Submodule** if you consume the library at a pinned commit and don't modify its requirements — your system requirements `trace:` up with version pins (`LIB-001~2`), and `reqmd repin` updates the pins when you bump. No conflicts, because you never edited upstream.

For the full comparison table and the fork-vs-submodule decision, see [Forks](/structure/#forks) on the Structure page.
{{< /details >}}

{{< details summary="What's the difference between a raw-git fork and a GitHub/GitLab fork?" card="true" >}}
Same axis — a separate repo that releases and reviews independently. A host fork (GitHub/GitLab "Fork" button) records the parent-link metadata and gives a sync UI ("Sync fork" / "Update now") plus the contribute-back-via-PR flow. A raw `git clone` + upstream remote gives the same git capabilities without the host record. Under the hood both are `git fetch upstream && git merge`; reqmd sees both as a repo to validate independently.

See [Forks](/structure/#forks) on the Structure page.
{{< /details >}}


## CI & performance


{{< details summary="Can I run reqmd in CI without a database?" card="true" >}}
Yes, no database, no daemon, no state. Each `check` is self-contained. Pin the version in CI. The recommended CI invocation:

```yaml
- run: go install github.com/dVoo/reqmd/cmd/reqmd@<pinned-version>
- run: reqmd check --json requirements/ > report.json
- run: reqmd check requirements/ --results ci-artifacts/
```

Use `--json` for programmatic parsing, plain output for human review.
{{< /details >}}

{{< details summary="Does reqmd scale to my 50,000-requirement spec?" card="true" >}}
Yes. Based on the [measured linear scaling](/benchmarks/) (~50 req/ms on 12 cores), 50,000 requirements would take ~2.5 seconds for `check`. Even 720,000 requirements finish in under 10 seconds. The bottleneck is file I/O, not the graph. See the [benchmarks page](/benchmarks/) for the full numbers.
{{< /details >}}

{{< details summary="What do the check / warning / error symbols mean?" card="true" >}}
Three severity levels:

- `✅` — passes schema validation and all required trace checks.
- `⚠` — passes schema validation but has a warning (broken reference, untraced, missing verdict). Warnings do not affect the exit code.
- `❌` — fails schema validation or has an error (cycle, missing required attribute, failing verdict). Errors exit non-zero.

Only errors affect the exit code. Warnings are informational.
{{< /details >}}

{{< details summary="What's the deal with the ladybug build tag?" card="true" >}}
`reqmd export graph` writes a [LadybugDB](https://ladybugdb.com/) database from your spec tree — requirements, traces, and verification results become a graph you can query with Cypher. Load the database in the [Ladybug explorer](https://docs.ladybugdb.com/visualization/lbug-explorer/) to run sophisticated trace checks that go beyond reqmd's built-in pass: multi-hop coverage paths, cross-V-model impact analysis, "which tests cover this stakeholder need through any chain of traces," and custom queries over the full traceability graph.

The LadybugDB driver pulls in a Cgo dependency, so `export graph` is gated behind a build tag:

```sh
go build -tags ladybug -o reqmd ./cmd/reqmd
```

Without the tag, `export graph` is not built. The rest of the CLI works exactly the same.
{{< /details >}}


## Still have questions?

- Read the [Quickstart](/quickstart/) — most "how do I…" questions are answered there.
- Read the [Cheat sheet](/cheat-sheet/) — full command overview and schema reference.
- Read the [Use cases](/use-cases/) page for what reqmd does and what it doesn't.
- See the [Compare](/compare/) page for how reqmd stacks up against other tools.
- Open an issue on [GitHub](https://github.com/dVoo/reqmd).