---
title: "Use cases"
description: "Where reqmd fits — from writing your first requirement to CI-gated traceability and live review."
weight: 3
---

reqmd is a tool for teams who need to write, link, and verify requirements. This page walks through the real workflows where reqmd adds value, then clearly states what is out of scope.

## What reqmd does

reqmd reads a tree of Markdown files — each requirement is a heading, a YAML `attr` block, and some prose — and checks every cross-reference, every status, every version pin. It produces HTML, CSV, and graph exports, loads verification results from CI or manual reviews, and closes the loop from spec to code.

You write requirements the way you already write documentation: in a text editor, in Git, in a pull request. reqmd adds the trace checks that a human reviewer would do — but mechanically, on every commit.

## Use cases

<div class="usecase-grid">
<article class="usecase">
<h3>Spec-first development</h3>
<p>Write the requirement, then the test, then the code. Traceability is enforced by the parser, not remembered by humans. The same pull request carries the spec change, the trace update, and the implementation — reviewed in one diff.</p>
<a href="/quickstart/02-trace-your-spec/">Step 2 — Trace your spec →</a>
</article>
<article class="usecase">
<h3>Requirements with status and lifecycle</h3>
<p>Every requirement is <code>draft</code> or <code>approved</code>, and carries a disposition: <code>implemented</code>, <code>deferred</code>, or <code>rejected</code> (with a mandatory reason). The schema enforces the lifecycle; the trace checks enforce that a <code>draft</code> requirement cannot satisfy an upstream coverage expectation.</p>
<a href="/quickstart/05-status-disposition/">Step 5 — Status &amp; disposition →</a>
</article>
<article class="usecase">
<h3>Traceability in CI</h3>
<p>Drop <code>reqmd check</code> into GitHub Actions or GitLab CI. Exit code 0/1/2 maps to pass/fail/parse-error. <code>--json</code> output is shaped for parsing. Per-check suppression and a <code>baseline diff</code> for comparing two tagged releases. Five lines of YAML, done.</p>
<a href="/quickstart/06-ci-integration/">Step 6 — CI integration →</a>
</article>
<article class="usecase">
<h3>Version pinning and impact tracking</h3>
<p>A downstream requirement can pin the upstream version it was verified against (<code>SYS-001~3</code>). When the upstream <code>version</code> is bumped, the pin becomes stale and the downstream needs re-verification. This is the engineering work of tracking what a version bump actually breaks.</p>
<a href="/quickstart/07-version-pins/">Step 7 — Version pins →</a>
</article>
<article class="usecase">
<h3>Verification result roll-up</h3>
<p>Load test results from CI (CTRF JSON) or manual reviews (Markdown). The <code>--results</code> flag runs <code>missing-verdict</code> and <code>failing-verdict</code> checks — an approved measure with no result, or a latest result that is fail. HTML export shows color-coded verdict badges on every measure.</p>
<a href="/quickstart/11-verification-results-ctrf/">Step 11 — CTRF results →</a>
</article>
<article class="usecase">
<h3>Product-line requirements with variants</h3>
<p>One spec tree, many configurations. Tag requirements with a <code>variant</code> attribute and scope every command — <code>check</code>, <code>ls</code>, <code>stats</code>, <code>export</code>, <code>baseline diff</code>, <code>serve</code> — with <code>--filter</code>. Coverage is evaluated within the view, cross-configuration trace mistakes fail via <code>--disjoint-check</code>, and a CI matrix validates every configuration on every pull request.</p>
<a href="/quickstart/14-variant-management/">Step 14 — Variant management →</a>
</article>
<article class="usecase">
<h3>Live stakeholder review</h3>
<p>Run <code>reqmd serve</code> while editing. The browser refreshes on every save, with trace links resolved and verdict badges color-coded. Reviewers read the spec as a website — with search, filters, and a light/dark toggle. No server framework, no SPA build.</p>
<a href="/quickstart/04-live-preview/">Step 4 — Live preview →</a>
</article>
<article class="usecase">
<h3>Submodule-based spec integration</h3>
<p>Compose specs from multiple git submodules. Use <code>document-id/ID</code> qualified references and <code>baseline diff</code> to see what changed between tags. Merge specs across teams without losing provenance.</p>
<a href="/quickstart/09-baseline-diff/">Step 9 — Baseline diff →</a>
</article>
<article class="usecase">
<h3>Source-code traceability</h3>
<p>The companion <code>reqmd-import</code> tool walks C, C++, Go, Python, or Rust source with tree-sitter and writes per-package <code>.md</code> requirements into your spec tree. Every function and type gets an ID in the same namespace as your spec — the same <code>reqmd check</code> that catches broken upstream refs also catches broken implementation refs.</p>
<a href="/import/">reqmd-import →</a>
</article>
<article class="usecase">
<h3>Custom schemas per document</h3>
<p>Each document directory carries its own <code>schema.yaml</code>. Define required fields, enums, ID prefixes, and per-directory metadata. Use <code>reqmd init</code> presets or build your own. The data model adapts to any V-model workflow.</p>
<a href="/quickstart/13-custom-templates/">Step 13 — Custom templates →</a>
</article>
<article class="usecase">
<h3>AI-assisted spec authoring</h3>
<p>An agent reads <code>llms.txt</code>, edits <code>*.md</code> in the same tool-call budget as code, runs <code>reqmd check</code>, parses the JSON verdict, and iterates. The check is 16 ms even on the bundled 249-req spec. Drop in a copy-pasteable <code>AGENTS.md</code> and you're done.</p>
<a href="/ai/">AI &amp; agents →</a>
</article>
</div>

## What reqmd is good at

These are the things reqmd does well, and that show up in every use case above:

**Bidirectional trace verification.** Every requirement ID across every document, every `trace` link, every `requires-trace-from` expectation, every `~N` version pin is checked mechanically. Broken references, circular dependencies, missing coverage — all caught on every commit.

**Status lifecycle and disposition.** A requirement is either in progress (`draft`) or signed off (`approved`); either actively developed (`disposition: implemented`), or parked with a reason (`deferred` / `rejected`, both with mandatory reasons). The lifecycle is enforced by the schema and the trace checks.

**Documented rationale.** Every requirement can carry a `Rationale:` paragraph alongside the body. The HTML export shows it. Reviewers see the *why*, not just the *what* — the engineering work of being able to answer "why is this requirement here?" six months from now.

**Version pinning.** A downstream requirement can pin the upstream version it was verified against. When the upstream version is bumped, the pin becomes stale and the downstream needs re-verification. This is the engineering work of tracking what a version bump actually breaks.

**CI-gated quality.** Stable exit codes, JSON output, per-check suppression. Every pull request that touches a spec, the agent (or the human) runs `reqmd check` and gets a pass/fail answer. No manual review of trace quality is required because the check is mechanical.

**Live HTML review.** A traceable HTML export with bidirectional links, document chain, theme toggle, and search. Reviewers can read the spec as a website. The engineering work is making the spec legible to the people who didn't write it.

## What is out of scope for reqmd

reqmd is a **requirements authoring and traceability tool**. It is not a project management tool, a risk management tool, or a process compliance framework. The following categories of work are explicitly out of scope — reqmd will not grow features for them:

**Project management.** Sprint planning, task assignment, effort estimation, milestone tracking, burndown charts. reqmd stores requirements, not tasks or schedules. Use a project tracker (Jira, Linear, GitHub Issues) for that.

**Risk management.** Risk registers, risk assessment matrices, risk mitigation tracking. reqmd can *store* a requirement that says "mitigate risk X" — but it does not model risk as a first-class object, does not compute risk priorities, and does not track mitigation status. Use a dedicated risk tool or a spreadsheet.

**Process compliance and audit trails.** Named approvers, sign-off records, electronic signatures, audit logs, regulatory submission packages. reqmd produces a *validated artifact* (the checked spec), not a *compliance record*. If your process requires named approvers and audit trails, use a tool designed for that.

**Supplier monitoring and procurement.** Supplier scorecards, purchase-order tracking, supplier qualification records. Not a spec problem.

**Measurement and process improvement.** KPI dashboards, process metrics, maturity assessments, CMMI levels. reqmd validates artifacts, not the organization that produced them.

**Notification and review threading.** Active notification registries, affected-party tracking, review-thread tools, comment management. The HTML export is a *passive publication* — it shows the spec, it does not manage a conversation about it. Use a code review tool or a wiki.

**Configuration and build management.** Version control of binaries, artifact repositories, build pipelines, release packaging. reqmd works *with* your VCS (it reads git tags for baseline diffs), but it is not a configuration management tool.

If your team does this work, reqmd can store the *requirements that result from it* — e.g. "REQ-001: the build pipeline shall produce a signed artifact." But reqmd does not *do* the work. No tool should pretend to conjure organizational process from a text file.

## See also

- [Quickstart](/quickstart/) — try the workflow in 10 minutes.
- [Benchmarks](/benchmarks/) — performance on a real 249-requirement spec and on synthetic 35-doc corpora up to 720k reqs.
- [reqmd-import](/import/) — the source-code traceability tool that closes the gap from code to spec.
- [Compare](/compare/) — how reqmd stacks up against other requirements tools.
- [AI &amp; agents](/ai/) — why reqmd's plain-text format is the right substrate for AI-assisted spec workflows.