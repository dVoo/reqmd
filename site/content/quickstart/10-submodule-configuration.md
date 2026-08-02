---
title: "Step 10 — Submodule-friendly configuration"
description: "Keep independently maintained requirement documents portable with stable document-id references instead of hard-coded parent paths."
weight: 10
---

Keep independently maintained requirement documents portable when their final integration paths are not known yet. Use stable `x-reqmd.document-id` values instead of hard-coded parent paths.

This is an optional step. It applies whenever your spec is assembled from git submodules or other configuration-managed sources where the mounted path isn't fixed.

## Example layout

The documents may live in separate git repositories or submodules and be mounted anywhere by the integrating repository:

```text
product-spec/
├── stakeholder-spec/    # stakeholder submodule
├── system-spec/         # system submodule
└── software-spec/       # software submodule
```

The three schemas in `10-submodule-configuration/` deliberately omit `upstream.sources`. Their `document-id` values remain stable regardless of where the submodules are checked out.

## How references work

The system requirement traces to `stakeholder/STK-001`, and the software requirement traces to `system/SYS-001`. These are document IDs, not directory paths:

```yaml
trace:
  - system/SYS-001
```

Coverage expectations use the same IDs:

```yaml
requires-trace-from: [software]
```

## Validate the assembled example

From the repository root, run:

```sh
reqmd check 10-submodule-configuration/
```

The command discovers all three schemas below one root and resolves their qualified trace references. If a submodule is validated by itself, references to documents outside that checkout cannot resolve; validate the assembled tree for cross-document checks.

## Boundary and HTML navigation

`upstream.sources` is optional, but it is still used for path-based boundary inference and HTML document-chain navigation. Add it in an integration-specific schema overlay only when the assembled repository has a stable directory layout. `document-id` references continue to work without it.

## Where to look things up

- The `document-id` and `upstream.sources` schema fields: [x-reqmd options in the cheat sheet](/cheat-sheet/#x-reqmd-options)

## What's next

Continue the V-model tour — load verification results for the boot tests. Go to [Step 11](/quickstart/11-verification-results-ctrf/).
