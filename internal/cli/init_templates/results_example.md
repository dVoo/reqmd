# Manual Verification Results

This document holds ephemeral verification results — review, inspection,
analysis, or demonstration outcomes. It is loaded via `reqmd check --results`
and is not part of the persistent spec tree.

The built-in `status`, `id`, `title`, and `trace` attributes are injected
automatically; `outcome`, `verifier`, `evidence`, and `verified-at` are
declared in the schema.

## {{ .IDPrefix }}-001: Example Review Result
```attr
status: approved
outcome: pass
verifier: "A. Reviewer"
evidence: minutes/example-review.md
verified-at: "2026-07-21"
trace:
  - TEST-001
```
The review confirmed the requirement is satisfied.

*Rationale: Manual verification methods (review, inspection, analysis, demonstration) record their results as markdown with a user-supplied schema. The `evidence` field links to the artifact carrying the full verification record.*