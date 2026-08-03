---
title: "FAQ"
description: "Frequently asked questions about reqmd — what it's for, how it works, and how it fits into your workflow."
weight: 6
---

Questions about reqmd, grouped by topic: what it is, how to use it, and how it fits your workflow.

## What is reqmd?

<div class="faq">
<details open>
<summary>What does reqmd do?</summary>
<div class="faq-body">
<p>reqmd is a command-line tool for writing, linking, and checking requirements and verification measures in plain Markdown. You write each requirement as a heading, a small YAML block, and some prose. reqmd checks every cross-reference — broken links, missing coverage, circular dependencies, stale version pins — and tells you what's wrong. It also exports to HTML, CSV, and graph formats.</p>
<p>The key idea: your spec lives in text files, in Git, next to your code (or in its own repo). The validation runs in milliseconds. No database, no server, no lock-in.</p>
</div>
</details>

<details>
<summary>What can I use reqmd for?</summary>
<div class="faq-body">
<p>Any project where you need to write requirements and trace them to each other, to tests, and to reviews:</p>
<ul>
  <li><strong>Spec-first development</strong> — write the requirement, then the test, then the code, all in one pull request.</li>
  <li><strong>Higher-level specification</strong> — stakeholder needs, system requirements, software design, and test specs traced across a V-model. The spec does <em>not</em> need to live next to the code; it can be a standalone repo for requirements and traceability at any level.</li>
  <li><strong>Verification and validation tracking</strong> — load test results from CI (CTRF JSON) and manual review results (Markdown), and get verdict badges showing which measures pass, fail, or are still missing.</li>
  <li><strong>CI-gated traceability</strong> — every pull request that touches a spec runs <code>reqmd check</code> and gets a pass/fail answer.</li>
  <li><strong>Live review</strong> — <code>reqmd serve</code> gives you a live-reloading HTML preview while you edit.</li>
</ul>
</div>
</details>

<details>
<summary>Does reqmd only handle requirements, or also verification measures?</summary>
<div class="faq-body">
<p>Both. reqmd treats verification measures — tests, reviews, inspections, analyses, demonstrations — as first-class objects. A test specification is a requirement with a <code>verify</code> attribute. A test result is loaded as a pseudo-requirement with an <code>outcome</code> (pass/fail/skipped/inconclusive) that traces back to the measure it verifies.</p>
<p>This means the full traceability chain is: <strong>stakeholder need → system requirement → software requirement → test specification → test result</strong>. Every link is checked. If an approved measure has no result, reqmd warns (<code>missing-verdict</code>). If the latest result is a fail, reqmd errors (<code>failing-verdict</code>).</p>
</div>
</details>

<details>
<summary>Does the spec need to live next to the code?</summary>
<div class="faq-body">
<p>No. reqmd works on any directory of Markdown files. Many teams keep the spec in the same Git repository as the code so that spec changes and code changes land in the same pull request. But a standalone requirements repository works just as well — point <code>reqmd check</code> at whatever directory contains your spec tree.</p>
<p>For higher-level requirement specification and traceability (stakeholder needs, system architecture, supplier specs), a separate repo is often the better choice. The spec is reviewed by people who don't work in the codebase — product managers, QA, safety engineers — and a dedicated repo with its own review cycle keeps things clean.</p>
</div>
</details>

<details>
<summary>How is reqmd different from a wiki or a Google Doc?</summary>
<div class="faq-body">
<p>A wiki stores text. reqmd stores text <em>and checks it</em>. In a wiki, a broken trace link is a human problem — someone has to notice it. In reqmd, a broken link is a CI failure — the tool catches it on every commit. You also get structured export (HTML with trace navigation, CSV, graph), version pinning, status lifecycle, and verification result tracking. None of that exists in a wiki.</p>
</div>
</details>

<details>
<summary>How is reqmd different from OpenFastTrace?</summary>
<div class="faq-body">
<p>Both are static-spec tools with a strong emphasis on bidirectional traceability. The differences:</p>
<ul>
  <li><strong>Data format.</strong> reqmd is Markdown-only with a YAML block; OFT uses its own <code>.xml</code> formats.</li>
  <li><strong>Runtime.</strong> reqmd is a single Go binary; OFT is a Java project.</li>
  <li><strong>Verification.</strong> reqmd has first-class CTRF ingestion and outcome-gated checks (<code>missing-verdict</code>, <code>failing-verdict</code>); OFT treats verification as separate artifacts.</li>
  <li><strong>Fit.</strong> OFT is a stronger fit for traditional spec-by-spec coverage analysis; reqmd is a stronger fit for git-native, text-first, CI-driven workflows.</li>
</ul>
<p>Both are good tools. Pick by your team's existing workflow.</p>
</div>
</details>

<details>
<summary>How is reqmd different from Jama / Polarion / DOORS?</summary>
<div class="faq-body">
<p>Those are database-backed requirements-management suites. They've been the default in regulated industries for 20 years, and they solve a real problem — managed workflows, audit trails, role-based access. But they come with structural problems that get worse as your team gets more agile:</p>
<ul>
  <li><strong>Slow.</strong> Every interaction is a web UI round-trip. A spec review that takes 10 minutes in a Git diff takes an hour in DOORS.</li>
  <li><strong>Disconnected from the code.</strong> The spec is in a database; the code is in Git. There is no native connection. When a developer renames a function, the spec doesn't know.</li>
  <li><strong>No real branching.</strong> "Baselines" are snapshots, not branches. You can't <code>git checkout -b</code>, try a change, and merge it. Two teams working on the spec simultaneously means lock-then-edit.</li>
  <li><strong>No submodules.</strong> A multi-team spec is one monolithic database. You can't split it into per-team repos and compose them.</li>
  <li><strong>Not CI-native.</strong> Running "does the spec validate?" in CI requires a REST API call, authentication, and a custom script. With reqmd it's <code>reqmd check spec/</code> — one binary, 16 ms, exit code 0/1/2.</li>
  <li><strong>Not developer-friendly.</strong> Developers don't open DOORS. They open their editor. If the spec is in a database, the developer never reads it, and the spec drifts from the code.</li>
  <li><strong>Vendor lock-in.</strong> Your spec is in a proprietary database schema. Exporting it is a project. With reqmd, your data is plain text — diffable, portable, yours.</li>
</ul>
<p><strong>Pick reqmd</strong> when you want the spec to live where the developers work — in Git, next to the code, with branching, submodules, CI, and pull-request review as the natural workflow. <strong>Pick an RM suite</strong> when you specifically need named approvers with electronic signatures and a regulated audit trail that an auditor reads from the tool. Some teams use both: reqmd for the engineering spec, the RM suite as the system of record. See the <a href="/compare/">compare page</a> for the full side-by-side.</p>
</div>
</details>

## How it works

<div class="faq">
<details>
<summary>How does the GitHub review workflow work?</summary>
<div class="faq-body">
<p>reqmd is designed for Git-based review. The spec lives in a Git repository (either next to the code or in a dedicated repo). Every change to a requirement, a trace link, a schema, or a verification result is a commit. Here's the typical workflow:</p>
<ol>
  <li><strong>Author.</strong> Open a branch. Edit the <code>.md</code> files — add a requirement, update a trace link, change a status. Commit and push.</li>
  <li><strong>Check locally.</strong> Run <code>reqmd check spec/</code> before pushing. If it exits non-zero, fix the issues.</li>
  <li><strong>Open a pull request.</strong> The diff shows exactly what changed: the requirement text, the trace link, the schema. Reviewers see the same diff they'd see for a code change.</li>
  <li><strong>CI runs the check.</strong> A GitHub Action (or GitLab CI job) runs <code>reqmd check --json spec/</code> on the PR. If it fails, the PR is blocked. The JSON output can be posted as a comment or an annotation.</li>
  <li><strong>Review.</strong> Reviewers read the diff, check the trace links, run <code>reqmd serve spec/</code> for a live preview if they want to see the rendered spec with links resolved.</li>
  <li><strong>Merge.</strong> Once the check passes and the review is approved, merge. The spec is now consistent — no broken links, no missing coverage, no stale version pins.</li>
</ol>
<p>The key advantage over a database-backed RM tool: the review happens in the same pull request interface the team already uses for code. The diff is the review artifact. The check is the gate. No separate tool, no separate login, no separate review state to track.</p>
</div>
</details>

<details>
<summary>How does the trace check work?</summary>
<div class="faq-body">
<p>reqmd builds a trace graph from every <code>trace:</code> link in every requirement, then checks it in one pass:</p>
<ul>
  <li><strong>Broken references</strong> — a <code>trace:</code> pointing to an ID that doesn't exist anywhere.</li>
  <li><strong>Missing coverage</strong> — a requirement with <code>requires-trace-from</code> that no downstream requirement traces to.</li>
  <li><strong>Circular dependencies</strong> — a chain of traces that loops back to itself.</li>
  <li><strong>Version pin staleness</strong> — a downstream pinned to <code>SYS-001~3</code> but the upstream is now version 5.</li>
  <li><strong>Missing verdicts</strong> — an approved measure with no verification result.</li>
  <li><strong>Failing verdicts</strong> — the latest result for a measure is <code>fail</code>.</li>
</ul>
<p>Broken references, cycles, and failing verdicts are errors (exit code 1). Missing coverage and missing verdicts are warnings (exit code 0, but shown in the output).</p>
</div>
</details>

<details>
<summary>How does the status lifecycle work?</summary>
<div class="faq-body">
<p>Every requirement has a <code>status</code>: <code>draft</code> (work in progress) or <code>approved</code> (signed off). Only <code>approved</code> requirements count as coverage providers — a <code>draft</code> requirement can be referenced by a trace but does <em>not</em> satisfy a <code>requires-trace-from</code> expectation.</p>
<p>Requirements can also carry a <code>disposition</code>: <code>implemented</code> (actively developed), <code>deferred</code> (parked, with a reason), or <code>rejected</code> (not doing it, with a reason). When a document directory sets <code>mandatory-disposition: true</code>, every requirement must have a disposition.</p>
</div>
</details>

<details>
<summary>How do I load verification results?</summary>
<div class="faq-body">
<p>Two formats, both loaded with the <code>--results</code> flag (repeatable):</p>
<ul>
  <li><strong>CTRF JSON</strong> — from automated test runs. The <code>x-reqmd.id</code> field in each test maps it to a requirement ID. reqmd reads the pass/fail/skipped verdict and the timestamp.</li>
  <li><strong>Manual results (Markdown)</strong> — for reviews, inspections, analyses. Each result is a Markdown file with an <code>attr</code> block carrying <code>outcome</code>, <code>verifier</code>, <code>verified-at</code>, and a <code>trace</code> back to the measure it verifies.</p>
</ul>
<p>reqmd synthesizes one pseudo-requirement per result and runs two new checks: <code>missing-verdict</code> (approved measure, no result) and <code>failing-verdict</code> (latest result is fail). The HTML export shows color-coded verdict badges on every measure card.</p>
</div>
</details>

<details>
<summary>Is there a web UI?</summary>
<div class="faq-body">
<p>There is no separate web UI; the HTML export <em>is</em> the web UI. Run <code>reqmd serve spec/</code> for a live-reloading preview, or <code>reqmd export html spec/ -o docs/</code> and host the resulting <code>docs/</code> directory as a static site.</p>
<p>The HTML export includes card-based layout, upstream/downstream trace links, a document chain tab strip for V-model navigation, status filters, theme toggle, and (with <code>--results</code>) color-coded verdict badges on measure cards.</p>
</div>
</details>

<details>
<summary>Does reqmd work without Git?</summary>
<div class="faq-body">
<p>Yes, for everything except <code>baseline diff</code>. The spec tree is plain files; no Git required to validate, list, or export. <code>baseline diff</code> is the one command that calls <code>git archive</code> to read a tagged snapshot — every other command operates on whatever directory you point it at.</p>
</div>
</details>

<details>
<summary>How do I use git submodules for configuration management?</summary>
<div class="faq-body">
<p>Configuration management &mdash; the discipline of controlling, versioning, and composing a multi-team spec across component boundaries &mdash; is a core systems engineering practice. In reqmd, it's built on Git submodules. Each team owns their own requirements repo as a git submodule, and a top-level repo assembles them into a single tree that <code>reqmd check</code> validates as a whole.</p>
<p>A typical layout:</p>
<pre><code>spec-monorepo/
  stakeholder/        &larr; submodule: git@github.com:team/stakeholder-spec.git
  system/             &larr; submodule: git@github.com:team/system-spec.git
  software/           &larr; submodule: git@github.com:team/software-spec.git
  tests/              &larr; submodule: git@github.com:team/test-spec.git</code></pre>
<p>Each submodule is a normal reqmd document directory with its own <code>schema.yaml</code>. The top-level repo just holds the submodule references &mdash; no spec content of its own. When a team updates their spec, they push to their repo and the top-level repo bumps the submodule pin.</p>
<p><strong>Cross-repo trace links</strong> work out of the box. A requirement in <code>system/</code> can trace to a requirement in <code>stakeholder/</code> by ID &mdash; reqmd resolves it across the submodule boundary because all submodules are on disk under the top-level directory. Use qualified references (<code>document-id/ID</code>) if two submodules share ID prefixes.</p>
<p><strong>The baseline is the submodule pins.</strong> When you need to compare two baselines, <code>reqmd baseline diff v1.0 v2.0</code> reads the submodule pins at each tag and reports what changed &mdash; added, removed, and modified requirements across all submodules, plus which submodule pins were updated. No separate configuration management tool needed; git is the tool.</p>
<p><strong>Review workflow</strong> works the same as a single repo. A PR that bumps a submodule pin shows the exact diff of the submodule's spec changes. Reviewers see what changed, CI runs <code>reqmd check</code> on the assembled tree, and the merge updates the baseline.</p>
</div>
</details>

<details>
<summary>How do I avoid hard-coded parent paths in submodule schemas?</summary>
<div class="faq-body">
<p>Use a stable <code>x-reqmd.document-id</code> for every independently maintained document and omit <code>upstream.sources</code> when the parent path is chosen by the integrating repository. The document ID stays stable even when a submodule is mounted at a different directory.</p>
<pre><code># system/schema.yaml
x-reqmd:
  document-id: system
  level: system-requirements
  upstream:
    level: stakeholder-needs
    # sources intentionally omitted</code></pre>
<p>Reference requirements by document ID in Markdown attributes:</p>
<pre><code>trace:
  - stakeholder/STK-001

requires-trace-from: [software]</code></pre>
<p>Run <code>reqmd check</code> from the assembled root containing all submodules. ReqMD discovers every schema below that root and resolves <code>document-id/requirement-id</code> references. A submodule validated by itself cannot resolve references to documents outside its checkout.</p>
<p><code>upstream.sources</code> remains available for HTML document-chain navigation and path-based boundary inference; add it only when the assembled repository has a stable layout. See the <a href="/quickstart/10-submodule-configuration/"><code>document-id</code> quickstart example</a>.</p>
</div>
</details>

<details>
<summary>How do I link requirements down to the actual code?</summary>
<div class="faq-body">
<p>Use the companion <a href="/import/"><code>reqmd-import</code></a> tool. It walks your Go or Python source code with a tree-sitter grammar and writes per-package <code>.md</code> requirements into your spec tree. Every function, type, and method gets a requirement ID in the same namespace as your spec.</p>
<p>The extracted code requirements are marked <code>external: true</code> and live in an <code>imported/</code> sub-tree. They don't need upstream traces (they're not authored by hand), but they <em>do</em> trace to the real implementation. The result: the same <code>reqmd check</code> that catches a broken spec reference also catches a broken implementation reference &mdash; if a function is renamed or deleted, the trace from the spec to that function becomes a broken reference.</p>
<p>The full traceability chain becomes: <strong>stakeholder need &rarr; system requirement &rarr; software requirement &rarr; test specification &rarr; test result &rarr; code implementation</strong>. Every link is a plain-text reference that reqmd validates mechanically.</p>
<p><code>reqmd-import</code> is a separate Go module with its own release cadence. It supports Go and Python today; Rust and Zig are reserved for future work.</p>
</div>
</details>
 </div>

## Schema & data

<div class="faq">
<details>
<summary>What does a schema.yaml file look like?</summary>
<div class="faq-body">
<p>Each document directory has a <code>schema.yaml</code> that defines the required attributes, their types, and reqmd-specific options. A minimal example:</p>
<pre><code class="language-yaml">$schema: "https://json-schema.org/draft/2020-12/schema"
$id: "my-requirements"
title: "My Requirements"
type: object
required:
  - priority
properties:
  priority:
    type: string
    enum: [Critical, High, Medium, Low]
additionalProperties: false

x-reqmd:
  level: system-requirements
  document-id: system
  id-prefix: SYS-
  upstream:
    level: stakeholder-needs
    sources:
      - ../01-stakeholder/</code></pre>
<p>See the <a href="/cheat-sheet/">cheat sheet</a> for the full list of <code>x-reqmd</code> options and all commands.</p>
</div>
</details>

<details>
<summary>Can I use JSON instead of YAML for the schema?</summary>
<div class="faq-body">
<p>No, the format is fixed to YAML. The <code>schema.yaml</code> filename and the YAML dialect are hard-coded. You can hand-author the file in JSON-compatible YAML and it will round-trip fine; extensions like <code>x-reqmd</code> work the same.</p>
<p>Why YAML? The data is small (a few hundred lines per directory at most), human-authored, and benefits from comments.</p>
</div>
</details>

<details>
<summary>What about Excel or Word import?</summary>
<div class="faq-body">
<p>Not built in. The recommended path is <code>reqmd init</code> with a custom preset and a one-time script that produces the Markdown files. For ongoing imports, write a small script that walks the Excel/Word file and emits one Markdown file per requirement.</p>
<p>The <code>reqmd-import</code> source-code extraction tool is the closest existing project — it extracts requirement IDs from Go and Python source via tree-sitter.</p>
</div>
</details>

<details>
<summary>Does reqmd work with my issue tracker (Jira, GitHub Issues, Linear)?</summary>
<div class="faq-body">
<p>No native integration. The common pattern is to put the issue-tracker key in a custom attribute (e.g. <code>jira: PROJ-123</code>) and search via <code>reqmd ls --json</code> or <code>reqmd stats --json</code>. The issue tracker is a downstream consumer of the spec, not a source.</p>
</div>
</details>

<details>
<summary>How does the trace check handle cycles?</summary>
<div class="faq-body">
<p><code>reqmd check</code> reports a cycle as an error (with the cycle path) and exits non-zero. There is no auto-resolution; cycles are a modeling problem to fix in the source.</p>
</div>
</details>

<details>
<summary>What about ReqIF?</summary>
<div class="faq-body">
<p>Not supported yet. The data model is compatible; the importer is what's missing. A future <code>reqmd-import reqif</code> subcommand is the most likely path.</p>
</div>
</details>
</div>

## Variants & configuration

<div class="faq">
<details>
<summary>Can I manage multiple product variants in one spec tree?</summary>
<div class="faq-body">
<p>Yes. reqmd deliberately has no built-in "variants" feature — you declare a <code>variant</code> array attribute in <code>schema.yaml</code> like any other custom attribute, tag each requirement, and scope every command with <code>--filter "&lt;expr&gt;"</code>. One spec tree serves any number of configurations (Base, Premium, Sport — or platforms, regions, owners) with no extra files, branches, or vocabulary.</p>
</div>
</details>

<details>
<summary>How do I include requirements that apply to all variants?</summary>
<div class="faq-body">
<p>A requirement with no <code>variant</code> attribute is treated as "common to all" — but the bare filter <code>"X" in variant</code> matches only requirements explicitly tagged X. Use the <code>or variant == nil</code> pattern to build the whole configuration view:</p>
<pre><code class="language-sh">reqmd check spec/ --filter '"Premium" in variant or variant == nil'</code></pre>
<p>This matches Premium-specific requirements plus the common-to-all ones. Use bare <code>"X" in variant</code> when you want only that configuration's specific requirements.</p>
</div>
</details>

<details>
<summary>Is coverage checking variant-aware?</summary>
<div class="faq-body">
<p>Yes — when <code>--filter</code> is active, coverage (<code>requires-trace-from</code>) is evaluated within the filtered subset only. A requirement excluded by the filter can neither require nor provide coverage. This prevents false failures: a Base-only requirement isn't reported under-covered just because its only provider is a Premium test that was never meant to cover it. Without <code>--filter</code>, coverage is computed across the whole tree exactly as before.</p>
</div>
</details>

<details>
<summary>How do I catch a trace link between incompatible variants?</summary>
<div class="faq-body">
<p>Use <code>--disjoint-check &lt;attr&gt;</code> — generalized, not variant-specific:</p>
<pre><code class="language-sh">reqmd check spec/ --disjoint-check variant</code></pre>
<p>For every trace link it verifies the two requirements share at least one value of the named array attribute. Zero intersection is an ERROR, e.g. a <code>variant: [Sport]</code> requirement tracing to a <code>variant: [Base]</code> target — no real configuration contains both. Requirements with an empty or absent value are exempt ("applies to all"). Declare it once in <code>schema.yaml</code> via <code>x-reqmd.disjoint-check: variant</code> to enforce it on every check.</p>
</div>
</details>

<details>
<summary>How do I produce a "what Premium adds over Base" report?</summary>
<div class="faq-body">
<p><code>baseline diff</code> has a same-commit two-view mode that compares two filtered views of the same tree — no divergent branches or tags needed:</p>
<pre><code class="language-sh">reqmd baseline diff \
  --filter-a 'variant == nil or "Base" in variant' \
  --filter-b '"Premium" in variant' \
  HEAD</code></pre>
<p>The output is a real semantic diff (added / removed / modified with attribute-level detail) — the configuration-management evidence artifact, generated from a single commit. You can also filter a normal two-tag diff: <code>reqmd baseline diff v1.0.0 v1.1.0 --filter '"Premium" in variant'</code>.</p>
</div>
</details>

<details>
<summary>Can I validate every variant in CI?</summary>
<div class="faq-body">
<p>Yes — run one <code>reqmd check</code> per configuration in a CI matrix. Each job validates its view with <code>--filter "${{ matrix.filter }}" --json</code>; the JSON summary records the exact filter in a <code>"filter"</code> field. Add a <code>--disjoint-check variant</code> job to catch cross-configuration trace mistakes. A complete example is in <a href="/quickstart/14-variant-management/">Step 14 — Variant management</a>.</p>
</div>
</details>

<details>
<summary>What does the filter expression language support?</summary>
<div class="faq-body">
<p><code>--filter</code> uses <code>expr-lang</code>, evaluated against each requirement's attribute map. Supported subset:</p>
<ul>
  <li><code>==</code>, <code>!=</code> — equality, e.g. <code>status == "approved"</code></li>
  <li><code>in</code> — array membership, e.g. <code>"Premium" in variant</code></li>
  <li><code>and</code>, <code>or</code>, <code>not</code> — boolean composition</li>
  <li><code>contains</code>, <code>startsWith</code>, <code>endsWith</code> — string helpers, e.g. <code>id startsWith "SYS-"</code></li>
  <li><code>variant == nil</code> — attribute absence / "common to all"</li>
</ul>
<p>Built-in variables are always available: <code>id</code>, <code>title</code>, <code>status</code>, <code>disposition</code>, <code>trace</code>, <code>version</code>. A filter referencing an attribute not declared in any <code>schema.yaml</code> is a compile-time ERROR — reqmd fails fast on typos instead of silently returning no matches.</p>
</div>
</details>

<details>
<summary>Do I need separate branches for each variant?</summary>
<div class="faq-body">
<p>No — that's the point. Variant differences live in the <code>variant</code> attribute of individual requirements, not in forked spec trees. The filter creates the views; <code>baseline diff --filter-a/--filter-b</code> creates the "variant A vs variant B" evidence; the CI matrix validates each view. One shared tree, one review process, no divergent branches to reconcile.</p>
</div>
</details>

<details>
<summary>How do I choose between variant attributes, branches, submodules, and forks?</summary>
<div class="faq-body">
<p>Three axes of divergence, one decision flow:</p>
<pre><code>Will the two things reconcile (merge or die)?
  yes → branch
  no → Do they release and review independently
       (separate cadence, no shared check)?
    yes → fork (separate repo)
    no  → variant attribute (one tree, one check, one review)</code></pre>
<p>The primary discriminator is the <strong>independence test</strong>: if the two things release and review independently (separate cadence, no shared <code>check</code>), they're a fork; if they must ship together from one commit and share one review, they're variants. The reqmd-native corollary is the <strong>validation test</strong>: if they must pass <code>check</code> against each other in one tree, they're variants. <strong>Submodules</strong> are orthogonal — an ownership/assembly mechanism that composes with all three, not a fourth axis.</p>
<p>For the full decision guide, a life-like topology showing all four mechanisms at once, and the fork-vs-submodule comparison, see the <a href="/structure/">Structure page</a>.</p>
</div>
</details>

<details>
<summary>When should I use a branch, and when is it the wrong tool?</summary>
<div class="faq-body">
<p>A branch is for change-in-progress that reconciles — it merges back or is discarded. It's the wrong tool when the difference is permanent: long-lived branches for variants lose single-tree validation and review (see <a href="/faq/#do-i-need-separate-branches-for-each-variant">Do I need separate branches for each variant?</a>), and long-lived branches that are really a separate product bit-rot and can't release independently (that's a fork). The one-line rule: <strong>if your branch will never merge and never die, it isn't a branch.</strong></p>
<p>See <a href="/structure/#branches">Branches</a> on the Structure page.</p>
</div>
</details>

<details>
<summary>When should I fork vs depend on a library at a pinned version?</summary>
<div class="faq-body">
<p>Discriminate by whether you intend to <em>modify</em> the library's requirements or <em>depend on</em> them at a version. <strong>Fork</strong> if you own and evolve a copy — you take the library's requirements as your starting point and change them, can track upstream (host fork or raw), and pay merge-conflict tax on local edits. <strong>Submodule</strong> if you consume the library at a pinned commit and don't modify its requirements — your system requirements <code>trace:</code> up with version pins (<code>LIB-001~2</code>), and <code>reqmd repin</code> updates the pins when you bump. No conflicts, because you never edited upstream.</p>
<p>For the full comparison table and the fork-vs-submodule decision, see <a href="/structure/#forks">Forks</a> on the Structure page.</p>
</div>
</details>

<details>
<summary>What's the difference between a raw-git fork and a GitHub/GitLab fork?</summary>
<div class="faq-body">
<p>Same axis — a separate repo that releases and reviews independently. A host fork (GitHub/GitLab "Fork" button) records the parent-link metadata and gives a sync UI ("Sync fork" / "Update now") plus the contribute-back-via-PR flow. A raw <code>git clone</code> + upstream remote gives the same git capabilities without the host record. Under the hood both are <code>git fetch upstream &amp;&amp; git merge</code>; reqmd sees both as a repo to validate independently.</p>
<p>See <a href="/structure/#forks">Forks</a> on the Structure page.</p>
</div>
</details>
</div>

## CI & performance

<div class="faq">
<details>
<summary>Can I run reqmd in CI without a database?</summary>
<div class="faq-body">
<p>Yes, no database, no daemon, no state. Each <code>check</code> is self-contained. Pin the version in CI. The recommended CI invocation:</p>

```yaml
- run: go install github.com/dVoo/reqmd/cmd/reqmd@<pinned-version>
- run: reqmd check --json requirements/ > report.json
- run: reqmd check requirements/ --results ci-artifacts/
```

<p>Use <code>--json</code> for programmatic parsing, plain output for human review.</p>
</div>
</details>

<details>
<summary>Does reqmd scale to my 50,000-requirement spec?</summary>
<div class="faq-body">
<p>Yes. Based on the <a href="/benchmarks/">measured linear scaling</a> (~50 req/ms on 12 cores), 50,000 requirements would take ~2.5 seconds for <code>check</code>. Even 720,000 requirements finish in under 10 seconds. The bottleneck is file I/O, not the graph. See the <a href="/benchmarks/">benchmarks page</a> for the full numbers.</p>
</div>
</details>

<details>
<summary>What do the check / warning / error symbols mean?</summary>
<div class="faq-body">
<p>Three severity levels:</p>
<ul>
  <li><code>✅</code> — passes schema validation and all required trace checks.</li>
  <li><code>⚠</code> — passes schema validation but has a warning (broken reference, untraced, missing verdict). Warnings do not affect the exit code.</li>
  <li><code>❌</code> — fails schema validation or has an error (cycle, missing required attribute, failing verdict). Errors exit non-zero.</li>
</ul>
<p>Only errors affect the exit code. Warnings are informational.</p>
</div>
</details>

<details>
<summary>What's the deal with the ladybug build tag?</summary>
<div class="faq-body">
<p><code>reqmd export graph</code> writes a LadybugDB database for Cypher queries. It pulls in a Cgo dependency, so it's gated behind a build tag:</p>

```sh
go build -tags ladybug -o reqmd ./cmd/reqmd
```

<p>Without the tag, <code>export graph</code> is not built. The rest of the CLI works exactly the same.</p>
</div>
</details>
</div>

## Still have questions?

- Read the [Quickstart](/quickstart/) — most "how do I…" questions are answered there.
- Read the [Cheat sheet](/cheat-sheet/) — full command overview and schema reference.
- Read the [Use cases](/use-cases/) page for what reqmd does and what it doesn't.
- See the [Compare](/compare/) page for how reqmd stacks up against other tools.
- Open an issue on [GitHub](https://github.com/dVoo/reqmd).