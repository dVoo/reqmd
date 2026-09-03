---
title: "Step 1 — Get started"
description: "Scaffold a new requirements project and run your first validation."
weight: 1
---

Scaffold a new requirements project and run your first validation.

## Scaffold

```sh
reqmd init my-project/
```

This creates:

```
my-project/
  schema.yaml       # JSON Schema with x-reqmd config
  requirements.md   # example file with two sample requirements
```

## The requirement format

A requirement is just a Markdown document — a heading, a YAML `attr` block, and free-form prose. Open `my-project/requirements.md` to see the format the scaffold generated:

````markdown
# Requirements

This is a sample requirements document. The `status`, `id`, `title`, and `trace`
attributes are built-in; `owner` is declared in the schema.

## REQ-001
```attr
status: draft
owner: Team Alpha
```
The system shall provide a login mechanism.

*Rationale: Authentication is required for secure access.*

## REQ-002
```attr
status: draft
owner: Team Beta
trace:
  - REQ-001
```
The system shall enforce role-based access control.
````

Three parts make up each requirement:

| Part | How it's written | Purpose |
|------|------------------|---------|
| **Heading** | `## REQ-001` — the ID, optionally followed by `: Title` | Gives the requirement its identity; reqmd reads the ID from the heading |
| **Attributes** | A fenced code block tagged ```attr``` containing YAML | Structured data: `status`, `owner`, `trace`, and anything else your schema declares |
| **Body** | Free-form Markdown prose after the `attr` block | The human-readable requirement statement |

The *Rationale:* paragraph is optional but encouraged — it shows up in the HTML export and gives reviewers the *why*.

The ID can also carry a title, e.g. `## REQ-001: Login`. Anything after the first `:` on the heading line is the requirement title. Every requirement ID must be unique across the whole spec tree — reqmd checks that for you.

## Heading structure

The scaffolded file also shows how reqmd treats headings *without* an
`attr` block: it doesn't drop them. Every heading becomes a node in a
document tree — with one exception: `#` (level 1) is the **document
title**, not content.

- **`# Requirements`** is the document title. It renders as the page
  title in exports and is never a requirement, container, or info node.
- **`## REQ-001`** and **`## REQ-002`** have `attr` blocks — **requirements**,
  the only nodes that participate in validation, trace checks, coverage,
  and `--filter`. Only level-2+ headings can be requirements.
- A heading such as **`## Notes`** (no `attr` block) with nested nodes is a
  **container**, a folder-like grouping; without children it's an **info**
  item.
- Prose attaches to the nearest heading's body; prose before the first
  heading becomes a headingless **info** item.

Containers and info items carry no attributes and take no part in
validation, trace checks, coverage, or `--filter` — they exist so exported
documents stay faithful to the source. `reqmd check` remains
requirement-scoped.

Heading levels are flexible from level 2 up: any level works, and a deeper
heading after a requirement nests under it as a sub-requirement:

```markdown
## SYS-001                        ← requirement
### SYS-001.1                     ← sub-requirement (parent = SYS-001)
### Notes                         ← info item (no attr block)
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

`reqmd ls` shows the whole document — the container and its requirements
side by side:

```text
=== my-project ===
Type       | ID                       | owner        | trace        | disposition  | disposition-reason | requires-trace-from | version      | status       |
-------------------------------------------------------------------------------------------------------------------------------------------------
container  | Requirements             |              |              |              |              |              |              |              |
req        | REQ-001                  | Team Alpha   |              |              |              |              |              | draft        |
req        | REQ-002                  | Team Beta    | ["REQ-001"]  |              |              |              |              | draft        |
```

The leading `Type` column marks every node: `req` for requirements,
`container` and `info` for headings without attribute blocks.

## Try the other presets

`reqmd init` ships three built-in presets:

```sh
# ASPICE-oriented template with safety attributes
reqmd init my-aspice/ --preset aspice

# Generic results template (review/inspection/analysis)
reqmd init my-results/ --preset results --id-prefix VR
```

## Where to look things up

- The full requirement-format spec, with a worked example: [The requirement format in the cheat sheet](/cheat-sheet/#the-requirement-format)
- Every `reqmd init` flag: [`reqmd init` in the cheat sheet](/cheat-sheet/#reqmd-init--scaffold-a-new-project)
- Every `reqmd check` flag and exit code: [`reqmd check` in the cheat sheet](/cheat-sheet/#reqmd-check--validate-requirements-and-trace-links)

## What's next

You now have a single-document project. The real power of reqmd comes from multi-level traceability — go to [Step 2](/quickstart/02-trace-your-spec/), where this example becomes the four-level V-model boot-sequence spec used through the rest of the quickstart.
