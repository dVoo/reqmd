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
