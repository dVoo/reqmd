---
description: >
  This document defines stakeholder requirements for reqmd derived from
  Automotive SPICE v4.0 base practices. Each requirement identifies which
  ASPICE BPs it fulfills and assesses implementation status within reqmd.
  These requirements serve as the bridge between the ASPICE v4.0 standard
  (00-aspice/) and reqmd's system features (02-system/).
---

# ASPICE v4.0 Stakeholder Requirements for reqmd

## ASP-SR-001: Structured requirement organization
```attr
process: SYS
aspice-bp: ["SYS2-BP2", "SWE1-BP2", "HWE1-BP2", "MLE1-BP2"]
status: approved
disposition: implemented
coverage: Full
trace:
  - ASP-SYS2-BP2
  - ASP-SWE1-BP2
  - ASP-HWE1-BP2
  - ASP-MLE1-BP2
```
The tool shall support structured organization of requirements using configurable heading levels, ID prefixes, and document-level grouping.

*Rationale:* ASPICE BP2 of each requirements analysis process requires structuring requirements. reqmd supports this via directory-based document grouping, configurable ID prefixes in `x-reqmd.id-prefix`, and dynamic requirement level discovery from the first heading with an attr block. Requirements can be organized hierarchically into documents (directories) and sub-requirements via heading level nesting.

*Already fulfilled by:* Dynamic reqLevel discovery, `x-reqmd.id-prefix` in schema, document directory grouping.


## ASP-SR-002: Schema-based requirement validation
```attr
process: SYS
aspice-bp: ["SYS2-BP1", "SWE1-BP1", "HWE1-BP1", "MLE1-BP1"]
status: approved
disposition: implemented
coverage: Full
trace:
  - ASP-SYS2-BP1
  - ASP-SWE1-BP1
  - ASP-HWE1-BP1
  - ASP-MLE1-BP1
```
The tool shall validate requirements against a project-defined schema, enforcing required attributes, value ranges, and type constraints for functional and non-functional requirements.

*Rationale:* ASPICE BP1 of each requirements analysis process requires specifying requirements according to defined characteristics. reqmd's JSON Schema 2020-12 validation enforces required fields, allowed values (enums), types, and additional property constraints per document directory.

*Already fulfilled by:* Schema compilation and validation per document directory (internal/schema/).


## ASP-SR-003: Heading-level hierarchy decomposition
```attr
process: SYS
aspice-bp: ["SYS2-BP2", "SWE1-BP2"]
status: approved
disposition: implemented
coverage: Full
trace:
  - ASP-SYS2-BP2
  - ASP-SWE1-BP2
```
The tool shall support requirement decomposition through heading-level hierarchy, automatically linking child requirements to their parent without manual trace declarations.

*Rationale:* ASPICE requires structuring requirements in a hierarchy. reqmd's dynamic `reqLevel` discovery lets authors use any heading level (`#`, `##`, `###`) for requirements. A heading at `reqLevel + 1` with an attr block becomes a child of the preceding `reqLevel` requirement, with automatic `ParentID` assignment and graph edge creation.

*Already fulfilled by:* Dynamic reqLevel parser, `###` sub-requirement detection via `isNextAttrBlock()`, Pass 3 auto-edges in graph.


## ASP-SR-004: Cross-document traceability
```attr
process: SYS
aspice-bp: ["SYS2-BP5", "SYS3-BP4", "SWE1-BP5", "SWE2-BP4", "SWE3-BP4", "HWE1-BP5", "HWE2-BP5"]
status: approved
disposition: implemented
coverage: Full
trace:
  - ASP-SYS2-BP5
  - ASP-SYS3-BP4
  - ASP-SWE1-BP5
  - ASP-SWE2-BP4
  - ASP-SWE3-BP4
  - ASP-HWE1-BP5
  - ASP-HWE2-BP5
```
The tool shall establish bidirectional traceability between requirements at adjacent V-model levels across document boundaries, and validate the integrity of trace chains.

*Rationale:* Multiple ASPICE base practices across SYS, SWE, and HWE require bidirectional traceability between successive refinement levels. reqmd's `trace` attribute supports upstream-only trace declarations, with the graph builder automatically computing bidirectional (inbound/outbound) relationships. Cross-document trace resolution uses absolute path lookups through `x-reqmd.upstream.sources`.

*Already fulfilled by:* `trace` attribute, graph `CachedNode` with Inbound/Outbound, `upstream` document chaining, check results for broken references.


## ASP-SR-005: Trace chain validation
```attr
process: SYS
aspice-bp: ["SYS2-BP5", "SYS3-BP4", "SWE4-BP4", "SWE5-BP6", "SWE6-BP4", "HWE3-BP5", "HWE4-BP5"]
status: approved
disposition: implemented
coverage: Full
trace:
  - ASP-SYS2-BP5
  - ASP-SYS3-BP4
  - ASP-SWE4-BP4
  - ASP-SWE5-BP6
  - ASP-SWE6-BP4
  - ASP-HWE3-BP5
  - ASP-HWE4-BP5
```
The tool shall verify completeness and integrity of trace chains, detecting broken references, untraced requirements, circular dependencies, and no-downstream linkages.

*Rationale:* ASPICE traceability BPs require consistency checking of trace relationships. reqmd's graph checker performs 10 distinct checks: broken reference (WARNING), circular dependencies (ERROR), untraced requirements (WARNING — suppressed for V-model top boundary, `external:true`, `disposition:set`, and sub-requirements), no-downstream (WARNING — suppressed for V-model bottom boundary and `disposition:set`), disposition without reason (WARNING), mandatory disposition (ERROR), ID prefix mismatch (ERROR), ID prefix collision (ERROR), duplicate requirement ID (ERROR), and ambiguous trace reference (ERROR).

*Already fulfilled by:* Graph Pass 2 (10 checks), `checkUntraced()`, `checkCircular()`, `checkBrokenReference()`, etc.


## ASP-SR-006: Requirement-to-verification traceability
```attr
process: SYS
aspice-bp: ["SYS4-BP4", "SYS5-BP4", "SWE4-BP4", "SWE5-BP6", "SWE6-BP4", "VAL1-BP3"]
status: approved
disposition: implemented
coverage: Full
trace:
  - ASP-SYS4-BP4
  - ASP-SYS5-BP4
  - ASP-SWE4-BP4
  - ASP-SWE5-BP6
  - ASP-SWE6-BP4
  - ASP-VAL1-BP3
```
The tool shall support traceability between requirements and their verification measures, enabling validation that each requirement is addressed by at least one test or verification specification.

*Rationale:* ASPICE verification processes require bidirectional traceability between requirements and verification measures. reqmd's V-model document chain (stakeholder → system → software → tests) supports requirements tracing to test specifications. The graph's no-downstream check identifies requirements without outgoing traces (unless in a V-model bottom boundary or with `disposition:set`).

*Already fulfilled by:* Multi-level doc chain tab strip, no-downstream check, boundary inference suppression.


## ASP-SR-007: Change tracking for requirements
```attr
process: SUP
aspice-bp: ["SUP10-BP4"]
status: approved
disposition: implemented
coverage: Full
trace:
  - ASP-SUP10-BP4
```
The tool shall support recording of change information for individual requirements, enabling bidirectional traceability between change requests and affected requirements.

*Rationale:* ASPICE SUP.10 requires traceability between change requests and affected work products. reqmd's built-in `version` attribute tracks version changes at the individual requirement level. The `disposition` and `disposition-reason` attributes document the resolution state. Change request IDs can be referenced via the `trace` attribute to establish explicit links.

*Already fulfilled by:* `version` built-in, `disposition`/`disposition-reason` attributes, `trace` attribute for CR linking.


## ASP-SR-008: Automated quality validation
```attr
process: SUP
aspice-bp: ["SUP1-BP3", "SUP1-BP6"]
status: approved
disposition: implemented
coverage: Full
trace:
  - ASP-SUP1-BP3
  - ASP-SUP1-BP6
```
The tool shall automatically validate requirement work products against defined quality criteria, report non-conformances with severity levels, and provide machine-readable output for CI pipeline integration.

*Rationale:* ASPICE SUP.1 requires quality assurance of work products and resolution of non-conformances. reqmd's validation pipeline (parse → schema validate → graph build → trace check) produces structured results with severity levels (ERROR, WARNING, INFO). The `--json` flag outputs machine-readable reports suitable for CI integration. Exit codes distinguish success (0), validation errors (1), and parse errors (2).

*Already fulfilled by:* 4-pass validation pipeline, severity levels, `--json` flag, exit codes (0/1/2).


## ASP-SR-009: Configuration management support
```attr
process: SUP
aspice-bp: ["SUP8-BP1", "SUP8-BP2", "SUP8-BP3", "SUP8-BP5", "SUP8-BP7"]
status: draft
disposition: deferred
disposition-reason: No built-in baseline capture or configuration management tool integration beyond what the VCS provides.
coverage: Partial
trace:
  - ASP-SUP8-BP1
  - ASP-SUP8-BP2
  - ASP-SUP8-BP3
  - ASP-SUP8-BP5
  - ASP-SUP8-BP7
```
The tool shall support configuration management of requirement work products through unique identification, version tracking, and consistent baseline representation.

*Rationale:* ASPICE SUP.8 requires identification, property definition, and control of configuration items. reqmd provides unique requirement identification via ID prefixes and the `version` built-in attribute. However, baseline management (SUP.8.BP5) and configuration management mechanisms (SUP.8.BP3) are delegated to the version control system (git). reqmd supports this by being Git-native: requirements are plain-text files that diff cleanly, and git tags/snapshots serve as baselines.

*Already fulfilled by:* Unique IDs per requirement, `version` attribute, Git-native file format. *Gap:* No built-in baseline capture or configuration management tool integration beyond what the VCS provides.


## ASP-SR-010: Readable requirement documentation
```attr
process: SYS
aspice-bp: ["SYS1-BP4", "SYS5-BP5", "SWE6-BP5", "VAL1-BP5", "SUP8-BP6"]
status: approved
disposition: implemented
coverage: Full
trace:
  - ASP-SYS1-BP4
  - ASP-SYS5-BP5
  - ASP-SWE6-BP5
  - ASP-VAL1-BP5
  - ASP-SUP8-BP6
```
The tool shall generate human-readable requirement documentation that communicates requirement status, trace relationships, and document structure to all affected parties.

*Rationale:* ASPICE requires communicating requirements status (SYS.1.BP4), verification results (SYS.5.BP5, SWE.6.BP5, VAL.1.BP5), and configuration status (SUP.8.BP6) to affected parties. reqmd's HTML export generates self-contained documentation with: card-based requirement layout showing all attributes, a tree TOC sidebar with scroll-spy navigation, trace chain tab strip showing the V-model hierarchy, dark mode toggle, and search/filter toolbar.

*Already fulfilled by:* HTML export with cards, tree TOC, trace tabs, dark mode, search/filter.


## ASP-SR-011: Interdependency analysis
```attr
process: SYS
aspice-bp: ["SYS2-BP3", "SWE1-BP3", "SWE2-BP3", "MLE1-BP3"]
status: draft
disposition: deferred
disposition-reason: Semantic correctness analysis is a human review concern.
coverage: Partial
trace:
  - ASP-SYS2-BP3
  - ASP-SWE1-BP3
  - ASP-SWE2-BP3
  - ASP-MLE1-BP3
```
The tool shall support analysis of requirement interdependencies to detect inconsistencies and support impact assessment.

*Rationale:* ASPICE requires analyzing requirement dependencies for correctness and feasibility (SYS.2.BP3, SWE.1.BP3). reqmd's graph provides dependency analysis through trace chain queries, `UpstreamNeighbors()` and `DownstreamNeighbors()` methods, and Pass 2 checks that detect circular dependencies and missing linkages. However, semantic analysis (correctness, feasibility) is beyond the tool's scope — these require domain expertise that reqmd supports by making dependencies visible.

*Already fulfilled by:* Graph query methods, circular dependency detection, no-downstream check. *Gap:* Semantic correctness analysis is a human review concern.

## ASP-SR-013: Issue and change tracking
```attr
process: SUP
aspice-bp: ["SUP9-BP1", "SUP9-BP6", "SUP10-BP1", "SUP10-BP5", "SUP10-BP6"]
status: draft
disposition: deferred
disposition-reason: No built-in issue/CR workflow.
coverage: Partial
trace:
  - ASP-SUP9-BP1
  - ASP-SUP9-BP6
  - ASP-SUP10-BP1
  - ASP-SUP10-BP5
  - ASP-SUP10-BP6
```
The tool shall support unique identification and tracking of issues and change requests that affect requirements.

*Rationale:* ASPICE SUP.9 and SUP.10 require unique identification, tracking, and closure confirmation for problems and change requests. reqmd supports this through unique requirement IDs that can reference external issue/CR IDs via the `trace` attribute. The `disposition` and `disposition-reason` attributes document resolution state. The tool's validation ensures trace integrity when CRs are linked. However, dedicated issue/CR management (workflow, approval, lifecycle) is delegated to external systems.

*Already fulfilled by:* Unique IDs, `trace` for CR linking, `disposition`/`disposition-reason`. *Gap:* No built-in issue/CR workflow.

## ASP-SR-015: External standard referencing
```attr
process: REU
aspice-bp: ["REU2-BP1", "REU2-BP5"]
status: approved
disposition: implemented
coverage: Full
trace:
  - ASP-REU2-BP1
  - ASP-REU2-BP5
```
The tool shall support referencing and reusing requirements from external standards, including those not authored within the tool itself.

*Rationale:* ASPICE REU.2 requires selection and provision of reusable products. reqmd supports this through the `external: true` flag on document directories, which marks requirements as external references. External requirements participate in the graph for trace validation (they can be traced to/from) but are exempt from certain checks (untraced suppression). The ASPICE v4.0 base practices (00-aspice/) are themselves expressed as external reqmd requirements, demonstrating this pattern.

*Already fulfilled by:* `external: true` flag, external requirements in graph, cross-document trace resolution.


## ASP-SR-016: Portable export distribution
```attr
process: SYS
aspice-bp: ["SYS2-BP6", "SYS3-BP5", "SWE1-BP6", "SWE2-BP5", "HWE1-BP6", "HWE2-BP6", "MLE1-BP6", "MLE2-BP7"]
status: approved
disposition: implemented
coverage: Full
trace:
  - ASP-SYS2-BP6
  - ASP-SYS3-BP5
  - ASP-SWE1-BP6
  - ASP-SWE2-BP5
  - ASP-HWE1-BP6
  - ASP-HWE2-BP6
  - ASP-MLE1-BP6
  - ASP-MLE2-BP7
```
The tool shall enable distribution of requirement documentation and impact analysis results through portable, self-contained export formats.

*Rationale:* ASPICE requires communicating agreed requirements and impact analysis to affected parties across all engineering domains. reqmd's HTML export generates self-contained `.html` files with inline CSS and JS — no server or network dependencies. The doc-level trace chain tab strip visually communicates the V-model context. The tree TOC sidebar with scroll-spy provides navigation. CSV export enables spreadsheet-based distribution for non-technical stakeholders.

*Already fulfilled by:* Standalone HTML export, CSV export, trace chain tab strip, tree TOC sidebar.


## ASP-SR-017: Impact analysis support
```attr
process: SYS
aspice-bp: ["SYS2-BP3", "SYS2-BP4", "SWE1-BP3", "SWE1-BP4", "HWE1-BP3", "HWE1-BP4"]
status: draft
disposition: deferred
disposition-reason: No dedicated `reqmd impact <id>` CLI command yet — current impact analysis requires using the graph API programmatically.
coverage: Partial
trace:
  - ASP-SYS2-BP3
  - ASP-SYS2-BP4
  - ASP-SWE1-BP3
  - ASP-SWE1-BP4
  - ASP-HWE1-BP3
  - ASP-HWE1-BP4
```
The tool shall support impact analysis by identifying all requirements affected by a proposed change through trace chain traversal.

*Rationale:* ASPICE requires analyzing the impact of requirements on system context and operating environment (SYS.2.BP4, SWE.1.BP4, HWE.1.BP4). reqmd's graph supports traversal via `UpstreamNeighbors()` and `DownstreamNeighbors()`, which provides the infrastructure for impact analysis queries.

*Already fulfilled by:* Graph traversal methods (`UpstreamNeighbors`, `DownstreamNeighbors`), bidirectional Inbound/Outbound edges. *Gap:* No dedicated `reqmd impact <id>` CLI command yet — current impact analysis requires using the graph API programmatically.


## ASP-SR-018: Requirement metrics and measurement
```attr
process: MAN
aspice-bp: ["MAN6-BP2", "MAN6-BP3", "MAN6-BP4"]
status: draft
disposition: deferred
disposition-reason: No historical trend tracking or metric dashboard.
coverage: Partial
trace:
  - ASP-MAN6-BP2
  - ASP-MAN6-BP3
  - ASP-MAN6-BP4
```
The tool shall provide metrics and measurements on requirement artifacts, including completeness, trace coverage, and attribute distribution.

*Rationale:* ASPICE MAN.6 requires specifying, collecting, and analyzing metrics. reqmd's `stats` command provides attribute-value breakdowns per document directory, showing distributions of all requirement attributes. The `ls` command lists all requirements with their attributes. The `--json` flag on all commands supports integration with external analytics tools. However, trend analysis over time and historical metric collection are out of scope — reqmd provides point-in-time snapshots.

*Already fulfilled by:* `stats` command with attribute breakdowns, `ls` command, `--json` output on all commands. *Gap:* No historical trend tracking or metric dashboard.

## ASP-SR-020: V-model refinement levels
```attr
process: SYS
aspice-bp: ["SYS5-BP2", "SWE4-BP2", "SWE5-BP3", "SWE6-BP2", "HWE3-BP3", "HWE4-BP3", "VAL1-BP2"]
status: approved
disposition: implemented
coverage: Full
trace:
  - ASP-SYS5-BP2
  - ASP-SWE4-BP2
  - ASP-SWE5-BP3
  - ASP-SWE6-BP2
  - ASP-HWE3-BP3
  - ASP-HWE4-BP3
  - ASP-VAL1-BP2
```
The tool shall support V-model work product progression through multiple levels of requirements refinement, with configurable trace links between adjacent levels.

*Rationale:* ASPICE requires selection of verification measures with sufficient coverage at each V-model level. reqmd's multi-level document structure (stakeholder → system → software → tests) mirrors the V-model directly. Each document directory has its own schema, ID prefix, and upstream trace target. The HTML export's trace chain tab strip renders the V-model chain (e.g., Stakeholder → System → Software → Tests) with the active document highlighted as an inactive span, providing visual context for the current level.

*Already fulfilled by:* Multi-level doc chain, `upstream` config, trace chain tab strip in HTML export, configurable per-dir schema.


## ASP-SR-021: Safety attribute support
```attr
process: SYS
aspice-bp: ["SYS2-BP1", "SWE1-BP1"]
status: draft
disposition: deferred
disposition-reason: ASIL consistency across the V-model trace chain is not validated — deferred to external tooling via graph export.
coverage: Partial
trace:
  - ASP-SYS2-BP1
  - ASP-SWE1-BP1
```
The tool shall support safety attribute assignment to requirements (ASIL QM/A/B/C/D per ISO 26262) as a user-defined schema attribute.

*Rationale:* ISO 26262-6 7.4.2 requires ASIL designation for software requirements. Users define `asil` as a normal attribute in their `schema.yaml` with the desired enum values. The tool's standard validation (required fields, enum values, type checking) applies to user-defined safety attributes through JSON Schema 2020-12 validation. ASIL consistency checking across the trace chain is out of scope — orgs requiring ASIL propagation checks should use the ladybugdb graph export and external query tools.

*Already fulfilled by:* User-defined schema attributes, JSON Schema validation (required, enum, type). *Gap:* ASIL consistency across the V-model trace chain is not validated — deferred to external tooling via graph export.
