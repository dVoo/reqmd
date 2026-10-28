# reqmd quickstart

Get from zero to a validated, linked requirements tree in about five minutes.
For the full tutorial — 15 numbered steps covering every feature — see the
[`quickstart/`](quickstart/) tree or the rendered guide at
<https://reqmd.dev/quickstart/>.

## 1. Install

Download a binary from the [releases page](https://github.com/dVoo/reqmd/releases),
or build from source:

```sh
git clone https://github.com/dVoo/reqmd
cd reqmd
go build -o reqmd ./cmd/reqmd
```

Requires Go 1.26 or newer for the source build.

## 2. Scaffold and validate a project

```sh
reqmd init my-project/
reqmd check my-project/
```

```
=
Schema : my-project — my-project Requirements
File   : my-project  (2 requirements)
=
  ✅  REQ-001  all attributes valid
  ✅  REQ-002  all attributes valid

Summary: 2 total, 2 valid, 0 invalid, 0 parse errors
```

`reqmd init` writes a `schema.yaml` and a `requirements.md` with two sample
requirements. Edit the Markdown, re-run `reqmd check`, and the schema and trace
rules are enforced on every save and commit.

## 3. Validate a linked V-model tree

The repo ships a complete four-level trace chain — stakeholder needs → system
requirements → software components → tests:

```sh
reqmd check quickstart/02-trace-your-spec/
```

```
=
Schema : software-requirements — Software Requirements
File   : quickstart/02-trace-your-spec/software  (1 requirements)
=
  ✅  SW-001  all attributes valid

=
Schema : stakeholder-goals — Stakeholder Goals
File   : quickstart/02-trace-your-spec/stakeholder  (1 requirements)
=
  ✅  STK-GOAL-001  all attributes valid

=
Schema : system-requirements — System Requirements
File   : quickstart/02-trace-your-spec/system  (1 requirements)
=
  ⚠  SYS-001  attributes valid (see trace checks below)
  ⚠  SYS-001  no upstream trace: no approved requirement from "software" traces to this item (1 draft downstreams ignored)

=
Schema : test-specs — Test Specifications
File   : quickstart/02-trace-your-spec/tests  (2 requirements)
=
  ✅  TEST-001  all attributes valid
  ✅  TEST-002  all attributes valid

Summary: 5 total, 5 valid, 0 invalid, 0 parse errors, 1 warnings
```

The one warning is intentional: `SYS-001` expects downstream coverage from the
`software` layer, but `SW-001` is still `draft` — a draft requirement does not
count as an approved coverage provider. Promote it to `approved` to make the
tree green; [Step 5](quickstart/05-status-disposition/) explains the status
lifecycle.

Trace links work through a `trace` attribute on the requirement:

````yaml
## SYS-001: Boot Sequence
```attr
status: approved
trace: [stakeholder/STK-GOAL-001]
```
````

The `document-id/ID` form is preferred — it stays stable when directories move
or documents are assembled from git submodules.

## 4. Browse and iterate

```sh
reqmd ls quickstart/02-trace-your-spec/                          # tabular listing
reqmd export html quickstart/02-trace-your-spec/ -o /tmp/spec-html/  # browsable report
reqmd serve quickstart/02-trace-your-spec/                       # live preview, auto-reload
reqmd check --json quickstart/02-trace-your-spec/                # machine-readable report
```

`reqmd export html` writes a standalone page per document with a sidebar TOC,
resolved cross-file trace links, and a document-chain view — no server needed.
`reqmd serve` watches the tree and reloads the browser on every edit.

## Where to go next

| If you want to… | Read |
|---|---|
| Walk every feature, step by step | [`quickstart/`](quickstart/) (14 steps) |
| Look up a flag or built-in attribute | [README.md](README.md) |
| Wire `reqmd check` into CI | [Step 6 — CI integration](quickstart/06-ci-integration/) |
| Compare two baselines | [Step 9 — Baseline diff](quickstart/09-baseline-diff/) |
| Load test results and gate on verdicts | [Step 11 — Verification results (CTRF)](quickstart/11-verification-results-ctrf/) |
| See the rendered tutorial | <https://reqmd.dev/quickstart/> |
