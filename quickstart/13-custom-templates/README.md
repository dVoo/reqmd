# Step 13 — Custom Templates

Define your own requirement schema with a custom `reqmd init` preset.

## What's in this folder

```
13-custom-templates/
  preset/                   ← the custom preset directory
    schema.yaml.tmpl        ← Go template for schema.yaml
    example.md.tmpl         ← Go template for the example .md file
```

This preset defines a **safety-critical requirements template** with:

- `safety-level` (required) — ASIL classification: QM, ASIL-A, ASIL-B, ASIL-C, ASIL-D
- `verification-method` (required) — test, review, analysis, inspection, demonstration
- `owner` (optional) — responsible team

## Scaffold from the custom preset

```sh
reqmd init my-safety/ --preset 13-custom-templates/preset/ --id-prefix SC
```

This produces:

```
my-safety/
  schema.yaml        ← rendered from schema.yaml.tmpl
  requirements.md   ← rendered from example.md.tmpl
```

## Validate the scaffolded output

```sh
reqmd check my-safety/
```

Expected: two requirements (`SC-001`, `SC-002`) pass validation.

## Template variables

Both `.tmpl` files use Go `text/template` syntax with:

| Variable | Source | Example |
|----------|--------|---------|
| `{{ .ID }}` | `--id` flag or dir basename | `my-safety` |
| `{{ .Title }}` | `--title` flag or `<dir> Requirements` | `my-safety Requirements` |
| `{{ .Level }}` | `--level` flag (default `requirements`) | `requirements` |
| `{{ .IDPrefix }}` | `--id-prefix` flag | `SC` |

## Create your own preset

1. Create a directory with `schema.yaml.tmpl` + `example.md.tmpl`
   (or plain `schema.yaml` + any `.md` file — no template syntax needed)
2. Use `{{ .ID }}`, `{{ .Title }}`, `{{ .Level }}`, `{{ .IDPrefix }}` variables
3. Scaffold with `reqmd init <dir> --preset <your-preset-dir>/`

## What's next

You can scaffold custom document types. Finally, scope one spec tree into
many configurations — go to [Step 14](../14-variant-management/).