# Step 10 — Submodule-Friendly Configuration

Keep independently maintained requirement documents portable when their final
integration paths are not known yet. This example uses stable
`x-reqmd.document-id` values instead of hard-coded parent paths.

## Example layout

The documents may live in separate git repositories or submodules and be
mounted anywhere by the integrating repository:

```text
product-spec/
├── stakeholder-spec/    # stakeholder submodule
├── system-spec/         # system submodule
└── software-spec/       # software submodule
```

The three schemas in this folder deliberately omit `upstream.sources`. Their
`document-id` values remain stable regardless of where the submodules are
checked out.

## How references work

The system requirement traces to `stakeholder/STK-001`, and the software
requirement traces to `system/SYS-001`. These are document IDs, not directory
paths:

```yaml
trace:
  - system/SYS-001
```

Coverage expectations use the same IDs, declared once per document in
`x-reqmd.requires-trace-from` so every requirement in the layer inherits them
regardless of where the submodule is mounted:

```yaml
# system/schema.yaml
x-reqmd:
  document-id: system
  level: system-requirements
  requires-trace-from: [software]
```

A requirement may override the document default with its own
`requires-trace-from` attribute, or opt out of downstream coverage with
`requires-trace-from: []`.

## Validate the assembled example

From the repository root, run:

```sh
reqmd check quickstart/10-submodule-configuration/
```

The command discovers all three schemas below one root and resolves their
qualified trace references. If a submodule is validated by itself, references
to documents outside that checkout cannot resolve; validate the assembled tree
for cross-document checks.

## Boundary and HTML navigation

`upstream.sources` is optional, but it is still used for path-based boundary
inference and HTML document-chain navigation. Add it in an integration-specific
schema overlay only when the assembled repository has a stable directory
layout. `document-id` references continue to work without it.
