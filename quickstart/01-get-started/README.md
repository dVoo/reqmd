# Step 1 — Get Started

Scaffold a new requirements project and run your first validation.

## Scaffold

```sh
reqmd init my-project/
```

This creates:

```
my-project/
  schema.yaml       ← JSON Schema with x-reqmd config
  requirements.md   ← example file with two sample requirements
```

## Validate

```sh
reqmd check my-project/
```

You should see both sample requirements pass validation:

```
Schema : my-project — my-project Requirements
File   : my-project  (2 requirements)

  ✅  REQ-001  all attributes valid
  ✅  REQ-002  all attributes valid

Summary: 2 total, 2 valid, 0 invalid, 0 parse errors
```

## Try the other presets

`reqmd init` ships three built-in presets:

```sh
# Automotive SPICE template with ASIL/safety attributes
reqmd init my-aspice/ --preset aspice

# Manual verification results template (review/inspection/analysis)
reqmd init my-results/ --preset results --id-prefix VR
```

## What's next

You now have a single-document project. The real power of reqmd comes from
multi-level traceability — go to [Step 2](../02-trace-your-spec/).