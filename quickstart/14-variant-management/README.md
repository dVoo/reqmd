# Step 14 — Variant Management

One spec tree, many configurations. This folder demonstrates variant
management with reqmd's generic `--filter` flag — no built-in "variants"
feature, just a custom `variant` attribute plus filtering.

## Layout

```
14-variant-management/
  system/            ← System requirements (top of V)
    schema.yaml       ← declares variant: [Base, Premium, Sport] + disjoint-check
    features.md       ← SYS-001 (Sport), SYS-002 (Base), SYS-003 (Premium)
  software/          ← Software requirements
    schema.yaml       ← upstream: ../system/, disjoint-check: variant
    components.md     ← SW-001..005 trace up to the system requirements
```

Three configurations (Base, Premium, Sport). Requirements with no `variant`
attribute are "common to all". `x-reqmd.disjoint-check: variant` is declared
in both schemas, so cross-configuration traces fail `check`.

## Try it

```sh
# Check everything (all variants together)
reqmd check 14-variant-management/

# Check only the Sport configuration — coverage is scoped to the view
reqmd check 14-variant-management/ --filter '"Sport" in variant'

# The full configuration view: Sport-specific + common-to-all requirements
reqmd check 14-variant-management/ --filter '"Sport" in variant or variant == nil'

# Per-configuration exports
reqmd export html 14-variant-management/ --filter '"Premium" in variant' -o html-out/premium/

# The disjoint-attribute check fires on a cross-configuration trace link
reqmd check 14-variant-management/ --disjoint-check variant
```

## Filter expressions

| Expression | Matches |
|---|---|
| `"Premium" in variant` | Premium-specific requirements |
| `variant == nil` | Common-to-all requirements |
| `"Premium" in variant or variant == nil` | The whole Premium configuration |
| `status == "approved"` | Approved requirements |
| `id startsWith "SYS-"` | Requirements whose ID starts with `SYS-` |
| `"Base" in variant and "Premium" in variant` | Multi-configuration requirements |

Built-ins available in every expression: `id`, `title`, `status`,
`disposition`, `trace`, `version`.

## Reference

- RFC: `discussions/variant-management-rfc.md`
- Site: [Step 14 — Variant management](https://reqmd.dev/quickstart/14-variant-management/)
