# Safety Requirements

This is a sample ASPICE-flavored requirements document. The `status`, `id`,
`title`, and `trace` attributes are built-in; `asil`, `safety-goal`, and
`owner` are declared in the schema.

## SAFE-001
```attr
status: draft
asil: C
safety-goal: SG-001
owner: Safety Team
```
The system shall detect a lane departure within 200 ms.

*Rationale: Required for functional safety compliance per ISO 26262.*

## SAFE-002
```attr
status: draft
asil: B
safety-goal: SG-002
owner: Safety Team
trace:
  - SAFE-001
```
The system shall issue a visual warning to the driver upon lane departure.
