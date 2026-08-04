# ReqMD Quickstart

In the next 10 minutes you will build a **three-document V-model trace chain** — stakeholder needs → system requirements → software requirements — and validate it with ReqMD. You will see how `trace` links requirements upstream, how `requires-trace-from` declares expected downstream coverage, and how ReqMD catches broken links before they reach production.

**What you will accomplish:**
1. Create three linked requirement documents
2. Validate them against JSON Schema and trace correctness
3. Export a browsable HTML report with resolved trace links
4. Fix coverage warnings by adding `requires-trace-from` and downstream traces

---

## Prerequisites

- Go 1.26 or newer (only if building from source)
- Any text editor

---

## Step 1 — Build reqmd

```sh
go build -o reqmd ./cmd/reqmd
```

Or download a pre-built binary from the [GitHub Releases page](https://github.com/dVoo/reqmd/releases) (Linux, macOS, Windows).

Verify it works:

```sh
./reqmd --help
```

---

## Step 2 — Create a three-document project

Create this directory tree:

```
quickstart/
├── stakeholder/
│   ├── schema.yaml
│   └── goals.md
├── system/
│   ├── schema.yaml
│   └── features.md
└── software/
    ├── schema.yaml
    └── components.md
```

### stakeholder/schema.yaml

```yaml
x-reqmd:
  level: stakeholder-needs
  document-id: stakeholder

$schema: "https://json-schema.org/draft/2020-12/schema"
$id: "stakeholder-goals"
title: "Stakeholder Goals"
type: object
required: [priority]
properties:
  priority:
    type: string
    enum: [Low, Medium, High, Critical]
additionalProperties: false
```

### stakeholder/goals.md

```markdown
# Stakeholder Goals

## STK-GOAL-001: Fast Boot
```attr
priority: High
requires-trace-from: [system]
```
The system shall boot to the home screen within 5 seconds of ignition on.

*Rationale:* Fast boot is a key differentiator for user experience.
```

### system/schema.yaml

```yaml
x-reqmd:
  level: system-requirements
  document-id: system
  upstream:
    level: stakeholder-needs
    sources:
      - ../stakeholder/

$schema: "https://json-schema.org/draft/2020-12/schema"
$id: "system-requirements"
title: "System Requirements"
type: object
required: [status]
properties:
  status:
    type: string
    enum: [draft, approved]
additionalProperties: false
```

### system/features.md

```markdown
# System Features

## SYS-001: Boot Sequence
```attr
status: approved
requires-trace-from: [software, tests]
trace: [stakeholder/STK-GOAL-001]
```
The system shall execute the boot sequence and display the home screen within 5 seconds.

*Rationale:* Derived from stakeholder goal STK-GOAL-001.
```

### software/schema.yaml

```yaml
x-reqmd:
  level: software-requirements
  document-id: software
  upstream:
    level: system-requirements
    sources:
      - ../system/

$schema: "https://json-schema.org/draft/2020-12/schema"
$id: "software-requirements"
title: "Software Requirements"
type: object
required: [status]
properties:
  status:
    type: string
    enum: [draft, approved]
additionalProperties: false
```

### software/components.md

```markdown
# Software Components

## SW-001: Boot Manager
```attr
status: draft
requires-trace-from: [tests]
trace: [system/SYS-001]
```
The boot manager shall initialize the display driver and launch the UI shell within 4.5 seconds.

*Rationale:* Leaves 0.5 s margin for hardware variation.
```

---

## Step 3 — Validate

Run the check command on the `quickstart/` root:

```sh
./reqmd check quickstart/
```

**Expected output:**

```text
=
Schema : stakeholder-goals — Stakeholder Goals
File   : quickstart/stakeholder  (1 requirements)
=
  ✅  STK-GOAL-001  all attributes valid

=
Schema : system-requirements — System Requirements
File   : quickstart/system  (1 requirements)
=
  ⚠  SYS-001      all attributes valid (trace warnings)
  ⚠  SYS-001      unknown coverage expectation: "tests"

=
Schema : software-requirements — Software Requirements
File   : quickstart/software  (1 requirements)
=
  ⚠  SW-001       all attributes valid (trace warnings)
  ⚠  SW-001       unknown coverage expectation: "tests"

Summary: 3 total, 3 valid, 0 invalid, 0 parse errors, 2 warnings
```

**What just happened:**

- `STK-GOAL-001` is fully valid — `SYS-001` traces to it via `stakeholder/STK-GOAL-001`, satisfying its `requires-trace-from: [system]` coverage expectation.
- `SYS-001` warns because it declares `requires-trace-from: [software, tests]`. `SW-001` traces back to it, so the `software` need is satisfied, but no `tests/` document exists yet, so the `tests` need shows a warning.
- `SW-001` warns because it declares `requires-trace-from: [tests]` but no `tests/` document traces to it yet.

**Key insight:** `requires-trace-from` tells ReqMD *which downstream documents you expect coverage from*. If you do not declare `requires-trace-from`, ReqMD falls back to generic boundary inference (orphan/missing-downstream warnings based on V-model topology). `requires-trace-from` gives you **per-requirement** precision.

---

## Step 4 — List all requirements

```sh
./reqmd ls quickstart/
```

**Expected output:**

```text
=== quickstart/stakeholder ===
ID             | priority | requires-trace-from | trace | version |
-------------------------------------------------------------------
STK-GOAL-001   | High     | ["system"]         |       |         |

=== quickstart/system ===
ID             | status   | requires-trace-from      | trace                       | version |
-----------------------------------------------------------------------------------------------------
SYS-001        | approved | ["software","tests"]   | ["stakeholder/STK-GOAL-001"] |         |

=== quickstart/software ===
ID             | status | requires-trace-from | trace              | version |
-------------------------------------------------------------------------------
SW-001         | draft  | ["tests"]            | ["system/SYS-001"] |         |
```

User-defined attributes appear first (`priority`, `status`), then built-in attributes (`requires-trace-from`, `trace`, `version`).

---

## Step 5 — Export HTML

```sh
./reqmd export html quickstart/ -o /tmp/quickstart-html/
```

Open `/tmp/quickstart-html/stakeholder-requirements.html` in a browser. You will see:

- A **left sidebar TOC** with all requirements
- **Cards** for each requirement showing attributes, body, rationale
- **Trace links** — upstream and downstream neighbours with clickable anchors
- A **document trace chain** tab strip at the top showing the V-model flow

The HTML is standalone — no server required.

---

## Step 6 — Fix the warnings

Create a `tests/` document to satisfy the remaining `requires-trace-from` coverage.

### tests/schema.yaml

```yaml
x-reqmd:
  level: test-spec-unit
  document-id: tests

$schema: "https://json-schema.org/draft/2020-12/schema"
$id: "test-specs"
title: "Test Specifications"
type: object
required: [test-type]
properties:
  test-type:
    type: string
    enum: [unit, integration, end-to-end]
additionalProperties: false
```

### tests/boot-test.md

```markdown
# Boot Tests

## TEST-001: Boot Time Measurement
```attr
test-type: unit
trace: [software/SW-001, system/SYS-001]
```
The test shall measure boot time on reference hardware and assert it is ≤ 5 s.
```

Run check again:

```sh
./reqmd check quickstart/
```

**Expected output:**

```text
=
Schema : stakeholder-goals — Stakeholder Goals
File   : quickstart/stakeholder  (1 requirements)
=
  ✅  STK-GOAL-001  all attributes valid

=
Schema : system-requirements — System Requirements
File   : quickstart/system  (1 requirements)
=
  ✅  SYS-001      all attributes valid

=
Schema : software-requirements — Software Requirements
File   : quickstart/software  (1 requirements)
=
  ✅  SW-001       all attributes valid

=
Schema : test-specs — Test Specifications
File   : quickstart/tests  (1 requirements)
=
  ✅  TEST-001     all attributes valid

Summary: 4 total, 4 valid, 0 invalid, 0 parse errors, 0 warnings
```

All green. Exit code `0`.

---

## Step 7 — Try `requires-trace-from: []` to opt out

Sometimes a requirement genuinely needs no downstream coverage. Set `requires-trace-from: []` to suppress the generic fallback warnings explicitly.

Edit `stakeholder/goals.md`:

```markdown
## STK-GOAL-001: Fast Boot
```attr
priority: High
requires-trace-from: []
```
```

Run check:

```sh
./reqmd check quickstart/
```

`requires-trace-from: []` tells ReqMD "I expect no downstream traces." No orphan warning appears.

---

## Step 8 — Compare baselines

Tag the current state, then change a requirement to see how `reqmd baseline diff`
reports the delta between git tags.

```sh
git tag -a quickstart-v1 -m "baseline v1"
```

Edit `system/features.md` to add a new requirement:

```markdown
## SYS-002: Resume Audio
```attr
status: draft
trace: [stakeholder/STK-GOAL-001]
```
The system shall resume the last active audio source after an ignition cycle.
```

Stage and commit the change, then tag it:

```sh
git add system/features.md
git commit -m "add resume audio requirement"
git tag -a quickstart-v2 -m "baseline v2"
```

Compare the two baselines:

```sh
./reqmd baseline diff quickstart-v1 quickstart-v2
```

**Expected output:**

```text
Requirements
  + SYS-002  added in quickstart-v2

Schemas
  no schema changes
```

Use `--json` for a machine-readable report:

```sh
./reqmd baseline diff quickstart-v1 quickstart-v2 --json
```

The command always exits `0` because diffing is informational, not validation.

---

## What you learned

| Concept | What it does |
|---------|-------------|
| `schema.yaml` | JSON Schema 2020-12 in YAML per document directory |
| `x-reqmd` | Directory metadata: `level`, `document-id`, `upstream`, `id-prefix` |
| `trace` | Upstream requirement references; supports `document-id/ID` qualified form |
| `requires-trace-from` | Coverage expectations — which downstream documents must trace to this requirement |
| `requires-trace-from: []` | Explicit opt-out: no downstream coverage expected |
| `version` | Integer version for trace pinning (`SYS-001~2`) |
| `disposition` + `disposition-reason` | Record why a requirement is deferred or rejected |
| `check` | Validate schema + trace graph; exit code 0 = all valid |
| `ls` | Tabular listing of all requirements with all attributes |
| `export html` | Standalone HTML with trace links and TOC |

---

## Next steps

- Read the [full README](README.md) for command reference and advanced features
- Explore the [dogfooding spec](spec/) — 233 real requirements validated by reqmd itself
- Compare with [OpenFastTrace](external_docs/openfasttrace_user_guide.md) to understand ReqMD's design differences
- Wire `reqmd check` into CI (see README CI example)

---

## Common pitfalls

| Symptom | Cause | Fix |
|---------|-------|-----|
| `missing properties: ["asil"]` | Required attribute missing | Add the attribute to the `attr` block |
| `broken reference: target "FOO-001" not found` | Trace references an ID that does not exist | Create the target requirement or fix the ID |
| `ambiguous reference` | Same ID exists in multiple documents | Use `document-id/ID` qualified form |
| `no downstream coverage from "tests"` | `requires-trace-from: [tests]` but no requirement in `tests/` traces to this one | Add a downstream trace or change `requires-trace-from` |
| `untraced: no downstream reference` | No `trace` attribute and not a root document | Add `trace: [parent/ID]` or `requires-trace-from: []` if root |
| `id-prefix mismatch` | ID does not start with directory's `id-prefix` | Fix the ID or the `id-prefix` in `schema.yaml` |

---

*This quickstart uses exact file paths and copy-pasteable content. Every command shown produces the output displayed when run against the files as written.*
