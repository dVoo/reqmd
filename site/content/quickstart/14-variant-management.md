---
title: "Step 14 — Variant management"
description: "One spec tree, many configurations: tag requirements with a variant attribute and scope every reqmd command with --filter."
weight: 14
---

One spec tree, many configurations. Tag requirements with a `variant` attribute and scope every reqmd command with `--filter`.

reqmd doesn't have a built-in `variants` feature — it doesn't need one. You declare a `variant` attribute in your schema exactly like any other custom attribute, tag each requirement with the configurations it belongs to, and then use the generic `--filter` flag to turn one spec tree into any number of ad-hoc views. The same mechanism works for any classification attribute: platform, region, priority, owner.

This step shows the full workflow on the `14-variant-management/` tree: three configurations (Base, Premium, Sport) across a two-level V-model.

## The pattern in 60 seconds

1. **Declare** the attribute in `schema.yaml` — a plain JSON Schema array with an `enum`:
2. **Tag** each requirement in its `attr` block.
3. **Scope** any command with `--filter "<expr>"`.

## 1. Declare the variant attribute

Each document directory declares `variant` as a regular schema property — no `x-reqmd` magic:

```yaml
# system/schema.yaml
properties:
  variant:
    type: array
    items:
      type: string
      enum: [Base, Premium, Sport]
additionalProperties: false
```

The `enum` gives you free validation: a typo like `variant: [Premuim]` fails `reqmd check` automatically.

## 2. Tag your requirements

````markdown
## SYS-001: Launch Control
```attr
status: approved
variant: [Sport]
```
````

A requirement can belong to several configurations (`variant: [Base, Premium]`). A requirement with no `variant` attribute is "common to all" — it appears in every view.

Downstream coverage is declared once for the whole layer in `x-reqmd.requires-trace-from` in `system/schema.yaml` (`[software-requirements]`); every system requirement inherits it. See [Step 2](/quickstart/02-trace-your-spec/).

## 3. Scope commands with --filter

```sh
# List only the Premium requirements
reqmd ls 14-variant-management/ --filter '"Premium" in variant'

# Check only the Base configuration — the check is scoped, not just the report
reqmd check 14-variant-management/ --filter '"Base" in variant'

# Per-configuration stats
reqmd stats 14-variant-management/ --filter '"Sport" in variant'

# Export only one configuration's HTML
reqmd export html 14-variant-management/ --filter '"Premium" in variant' -o html-out/premium/
```

The expression language is `expr-lang`, evaluated against each requirement's attributes plus the built-ins `id`, `title`, `status`, `disposition`, `trace`, and `version`:

| Expression | Matches |
|---|---|
| `"Premium" in variant` | Requirements tagged Premium |
| `variant == nil` | Requirements common to all configurations |
| `"Premium" in variant or variant == nil` | The whole Premium configuration (Premium-only **plus** common) |
| `status == "approved"` | Approved requirements |
| `id startsWith "SYS-"` | Requirements whose ID starts with `SYS-` |
| `"Base" in variant and "Premium" in variant` | Multi-configuration requirements |

<div class="callout">
<div class="callout-icon">
<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/></svg>
</div>
<div class="callout-body">
<p><strong>Use <code>or variant == nil</code> for the "common to all" rule.</strong> The usual configuration view is <code>"X" in variant or variant == nil</code> — the configuration's own requirements plus the shared ones. Use bare <code>"X" in variant</code> when you want only that configuration's specific requirements.</p>
</div>
</div>

## Filter-aware coverage checking

Coverage is evaluated *within* the filtered subset, not across the whole tree. A requirement excluded by the filter can neither require nor provide coverage — so a Base requirement doesn't falsely appear under-covered because its only provider is a Premium test that was never meant to cover it.

```sh
# Base view: SYS-002 (Base) is covered by SW-002 (Base) — passes
reqmd check 14-variant-management/ --filter '"Base" in variant'

# Premium view: SYS-003 is covered by SW-003 — passes, without the Base
# requirements dragging in their own coverage expectations
reqmd check 14-variant-management/ --filter '"Premium" in variant'
```

## Catch cross-configuration trace mistakes

A trace link between two requirements that share no common configuration is a modeling error — no real configuration contains both. reqmd can check this for any array attribute:

```sh
reqmd check 14-variant-management/ --disjoint-check variant
```

Or declare it once in the schema so every `check` enforces it:

```yaml
x-reqmd:
  disjoint-check: variant
```

A cross-configuration trace then fails:

```
❌ SW-005  disjoint-attribute: traces to SYS-002 (variant: [Base]) but SW-005's
           variant [Sport] has no overlap — no valid configuration includes both
```

## Variant CI matrix

The `--filter` flag turns one spec into a full variant test matrix. Each configuration is checked as its own job; every pull request must pass all of them:

```yaml
# .github/workflows/variants.yml
name: Validate variants
on: [pull_request]
jobs:
  check-all-variants:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        include:
          - { name: Base,    filter: '"Base" in variant or variant == nil' }
          - { name: Premium, filter: '"Premium" in variant or variant == nil' }
          - { name: Sport,   filter: '"Sport" in variant or variant == nil' }
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.26' }
      - run: go build -o reqmd ./cmd/reqmd
      - run: ./reqmd check 14-variant-management/ --filter "${{ matrix.filter }}" --json > report-${{ matrix.name }}.json
      - run: ./reqmd check 14-variant-management/ --disjoint-check variant
      - uses: actions/upload-artifact@v4
        if: always()
        with:
          name: reqmd-report-${{ matrix.name }}
          path: report-${{ matrix.name }}.json
```

The `--json` summary includes a `"filter"` field so each report records exactly which view it validated.

## Per-configuration baselines

`--filter` works with `baseline diff` too — compare two tags restricted to one configuration:

```sh
reqmd baseline diff v1.0.0 v1.1.0 --filter '"Premium" in variant or variant == nil'
```

And the same-commit two-view mode answers "what does Premium add over Base" without needing two branches:

```sh
reqmd baseline diff \
  --filter-a 'variant == nil or "Base" in variant' \
  --filter-b '"Premium" in variant' \
  HEAD
```

The output is a real diff report (added / removed / modified with attribute-level detail) — the configuration-management evidence artifact for the ASPICE record, generated from a single commit.

## Preview a configuration while editing

```sh
reqmd serve 14-variant-management/ --filter '"Sport" in variant or variant == nil'
```

Reviewers preview exactly the Sport configuration locally — trace links, status, and verdicts all scoped to that view — before approving.

## Where to look things up

- The filter expression language and built-ins: [Filter expressions in the cheat sheet](/cheat-sheet/#filter-expressions)
- `--filter` on every command: [`reqmd check`](/cheat-sheet/#reqmd-check--validate-requirements-and-trace-links), [`reqmd ls`](/cheat-sheet/#reqmd-ls--list-all-requirements), [`reqmd stats`](/cheat-sheet/#reqmd-stats--attribute-value-breakdown-per-document), [`reqmd export html`](/cheat-sheet/#reqmd-export-html--export-to-html)

## What's next

You've seen the generic filtering primitive that makes variant management work. It's the same flag used for any attribute slice — owner, priority, platform. For verification per configuration, combine `--filter` with `--results` (see [Step 11](/quickstart/11-verification-results-ctrf/)); for the full command surface, see the [Cheat sheet](/cheat-sheet/).
