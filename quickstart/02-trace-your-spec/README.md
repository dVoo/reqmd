# Step 2 — Trace Your Spec

Build a multi-level spec tree with trace links between levels.

This folder contains a complete four-level V-model spec tree:

```
02-trace-your-spec/
  stakeholder/          ← Stakeholder needs (top of V)
    schema.yaml
    goals.md             ← STK-GOAL-001: Fast Boot
  system/               ← System requirements
    schema.yaml          ← upstream: ../stakeholder/
    features.md          ← SYS-001 traces to STK-GOAL-001
  software/             ← Software requirements
    schema.yaml          ← upstream: ../system/
    components.md        ← SW-001 traces to SYS-001
  tests/                 ← Test specifications (bottom of V)
    schema.yaml          ← upstream: ../software/
    boot-test.md         ← TEST-001/002 trace to SW-001 and SYS-001
```

## How trace links work

Each requirement has a `trace` attribute listing its upstream IDs. reqmd
validates that every trace target exists:

```yaml
## SYS-001: Boot Sequence
```attr
status: approved
requires-trace-from: [software, tests]
trace: [stakeholder/STK-GOAL-001]
```
```

The `requires-trace-from` attribute declares which downstream levels must trace
back to this requirement. reqmd checks coverage both ways.

## Validate the full tree

```sh
reqmd check 02-trace-your-spec/
```

Expected output:

```
Summary: 5 total, 5 valid, 0 invalid, 0 parse errors
```

## List all requirements

```sh
reqmd ls 02-trace-your-spec/
```

## Stats per document

```sh
reqmd stats 02-trace-your-spec/
```

## What's next

The spec tree validates. Now export it for stakeholders — go to
[Step 3](../03-export/).