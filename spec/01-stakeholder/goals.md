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
The requirement management tool shall enable text-first, Git-native authoring of requirement specifications without requiring a GUI.

*Rationale:* Engineers work in IDEs and version control daily. A plain-text format that diffs cleanly and reviews in standard pull requests minimizes tool-switching friction.

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
The tool shall validate requirements against a project-defined schema, rejecting malformed or incomplete entries before they reach production reviews.

*Rationale:* Manual review alone does not scale. Schema-driven validation catches missing fields, invalid values, and broken trace chains automatically in CI.

## STK-GOAL-004: Multi-format export
```attr
priority: High
source: Product
category: Export
```
The tool shall export requirement specifications to standard interchange formats including CSV, HTML, and graph database. The HTML export shall present requirements in a card-based layout with a persistent tree TOC sidebar for navigation.

*Rationale:* Tier-1 suppliers and homologation authorities require structured export formats. CSV covers basic exchange; HTML supports human review with trace navigation and a tree sidebar for browsing requirement hierarchies.

## STK-GOAL-005: V-model traceability
```attr
priority: High
source: Safety
category: Traceability
```
The tool shall support V-model traceability patterns conformant with ASPICE and ISO 26262, including upstream-only trace declarations, disposition tracking, and automatic parent-child hierarchy linking from heading levels.

*Rationale:* Functional safety standards mandate bidirectional traceability from stakeholder needs through system requirements, software requirements, and test specifications. Heading-level hierarchy lets authors decompose requirements naturally while the tool auto-links child-to-parent traces.
