---
description: Stakeholder goals define the high-level business and technical objectives that the reqmd tool must satisfy. As the top V-model layer, they have no upstream traces.
---

# reqmd Stakeholder Goals

## STK-GOAL-001: Text-first Git-native authoring
```attr
priority: Critical
source: Product
category: Authoring
```
The requirement management tool shall enable text-first, Git-native authoring of requirement specifications without requiring a GUI. New projects shall be able to start from a working scaffold (schema plus example document) without hand-writing schema boilerplate.

*Rationale:* Engineers work in IDEs and version control daily. A plain-text format that diffs cleanly and reviews in standard pull requests minimizes tool-switching friction. A scaffolded starting point removes the blank-page cost of adopting a schema-driven format.

## STK-GOAL-002: Human-readable machine-parsable specs
```attr
priority: Critical
source: Engineering
category: Authoring
```
Requirement documents shall be human-readable in any Markdown renderer and simultaneously machine-parsable into structured attributes.

*Rationale:* Non-technical stakeholders review requirements in rendered form (GitHub, GitLab, docs sites). Automated tooling needs structured data for validation and export. Both modes must coexist in the same file.

## STK-GOAL-003: Schema-driven validation
```attr
priority: Critical
source: Quality
category: Validation
```
The tool shall validate requirements against a project-defined schema, rejecting malformed or incomplete entries before they reach production reviews. Validation shall be scopable to a filtered subset of the tree without losing trace resolution.

*Rationale:* Manual review alone does not scale. Schema-driven validation catches missing fields, invalid values, and broken trace chains automatically in CI. Scoping keeps incremental CI affordable on large trees.

## STK-GOAL-004: Multi-format export for review and analytics
```attr
priority: High
source: Product
category: Export
```
The tool shall export requirement specifications to standard interchange formats: CSV for suppliers, standalone HTML for human review, flat NDJSON tables for reporting platforms and data warehouses, and a graph database for structural queries.

*Rationale:* Tier-1 suppliers and homologation authorities require structured export formats. CSV covers basic exchange; HTML supports human trace navigation; NDJSON feeds analytics and dashboards without a bespoke integration; the graph export serves cross-cutting dependency and impact queries.

## STK-GOAL-005: V-model traceability
```attr
priority: High
source: Safety
category: Traceability
```
The tool shall support V-model traceability patterns conformant with ASPICE and ISO 26262, including upstream-only trace declarations, disposition tracking, version pins, and automatic parent-child hierarchy linking from heading levels.

*Rationale:* Functional safety standards mandate bidirectional traceability from stakeholder needs through system requirements, software requirements, and test specifications. Heading-level hierarchy lets authors decompose requirements naturally while the tool auto-links child-to-parent traces.

## STK-GOAL-006: Verification results close the V-model
```attr
priority: High
source: Quality
category: Verification
```
The tool shall consume verification and validation results produced by automated tests and manual reviews, and shall enforce that every approved verification measure carries a current, passing result.

*Rationale:* A trace to a test specification proves that verification was *planned*, not that it *passed*. ASPICE requires recording and evaluating verification results. Loading results per invocation (CTRF reports and review documents) keeps run-to-run churn out of git while still failing CI on a failing or missing verdict.

## STK-GOAL-007: Interoperable requirement ingestion
```attr
priority: High
source: Engineering
category: Interoperability
```
The tool shall ingest requirements that originate outside authored Markdown — source-code symbols with documented trace markers, and records exported from ALM systems such as DOORS or codeBeamer — into the same ID space and document format as hand-authored requirements.

*Rationale:* Real projects do not start with an empty repository. Imported items must validate, trace, export, and diff exactly like authored ones, or the tool only covers half the specification. A single versioned exchange format keeps every importer small and keeps the Markdown/YAML syntax rules in one implementation.
