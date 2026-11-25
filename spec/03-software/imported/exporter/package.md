# Package: exporter

<!-- reqmd: generated -->

## internal/exporter/analysis.go
### IMP-exporter-LinkState-3d8ba67: LinkState
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/analysis.go
x-reqmd.source-line: 14
x-reqmd.symbol-kind: type
trace: []
```

LinkState is one trace link as the views present it. Title is the target's
human-readable title; Pin/Direction describe the authored version pin and
its health against the target's current Version.

### IMP-exporter-NodeState-d58911b: NodeState
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/analysis.go
x-reqmd.source-line: 28
x-reqmd.symbol-kind: type
trace: []
```

NodeState is the presentation state of one requirement, adapted from graph
analysis. Views render these states instead of inferring coverage or
verification from link presence.

### IMP-exporter-Analysis-8356668: Analysis
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/analysis.go
x-reqmd.source-line: 44
x-reqmd.symbol-kind: type
trace: []
```

Analysis is the semantic layer shared by every HTML view: coverage,
verification, and trace link health adapted from the graph so the export
cannot disagree with reqmd check.

#### IMP-exporter-mirrorDownstream-813dfdb: mirrorDownstream
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/analysis.go
x-reqmd.source-line: 157
x-reqmd.symbol-kind: method
trace: []
```

mirrorDownstream copies every resolved outbound link into its target's
downstream list, keeping the authored pin and its health, then sorts each
downstream list for deterministic output.

#### IMP-exporter-CoverageOf-79c00d7: CoverageOf
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/analysis.go
x-reqmd.source-line: 199
x-reqmd.symbol-kind: method
trace: []
```

CoverageOf summarizes the given requirement states.

#### IMP-exporter-VerdictsOf-41a2273: VerdictsOf
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/analysis.go
x-reqmd.source-line: 238
x-reqmd.symbol-kind: method
trace: []
```

VerdictsOf summarizes the verification states of the given requirements.

#### IMP-exporter-LinksOf-ec77cd8: LinksOf
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/analysis.go
x-reqmd.source-line: 272
x-reqmd.symbol-kind: method
trace: []
```

LinksOf summarizes the declared outbound trace links of the given
requirements.

### IMP-exporter-NewAnalysis-a7d8896: NewAnalysis
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/analysis.go
x-reqmd.source-line: 51
x-reqmd.symbol-kind: function
trace: []
```

NewAnalysis adapts graph results into presentation state. titleMap supplies
link titles; missing entries fall back to the bare ID.

### IMP-exporter-realRequirementIDs-85a8a31: realRequirementIDs
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/analysis.go
x-reqmd.source-line: 78
x-reqmd.symbol-kind: function
trace: []
```

realRequirementIDs returns the IDs of authored requirements. Synthetic nodes
(RESULT: pseudo-requirements and synthesized test cases) are evidence, not
requirements: they have no page and must not appear as trace links.

### IMP-exporter-nodeState-abdcb3b: nodeState
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/analysis.go
x-reqmd.source-line: 95
x-reqmd.symbol-kind: function
trace: []
```

nodeState assembles one requirement's presentation state.

### IMP-exporter-evidenceItems-1adbdcd: evidenceItems
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/analysis.go
x-reqmd.source-line: 116
x-reqmd.symbol-kind: function
trace: []
```

evidenceItems converts graph evidence into the exporter's render type.

### IMP-exporter-upLinks-e625598: upLinks
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/analysis.go
x-reqmd.source-line: 134
x-reqmd.symbol-kind: function
trace: []
```

upLinks keeps the authored trace refs that point at a real requirement (or
are unresolved, so broken links stay visible).

### IMP-exporter-CoverageSummary-b108d40: CoverageSummary
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/analysis.go
x-reqmd.source-line: 188
x-reqmd.symbol-kind: type
trace: []
```

CoverageSummary counts the coverage states of the given requirements.

### IMP-exporter-VerdictSummary-3236c9d: VerdictSummary
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/analysis.go
x-reqmd.source-line: 228
x-reqmd.symbol-kind: type
trace: []
```

VerdictSummary counts verification states over the eligible measures of the
given requirements. Unverified eligible measures are counted as NoVerdict,
so percentages always divide by the eligible total.

### IMP-exporter-LinkSummary-6938ace: LinkSummary
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/analysis.go
x-reqmd.source-line: 263
x-reqmd.symbol-kind: type
trace: []
```

LinkSummary counts trace link health across the given requirements.

### IMP-exporter-isSyntheticID-c152164: isSyntheticID
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/analysis.go
x-reqmd.source-line: 293
x-reqmd.symbol-kind: function
trace: []
```

isSyntheticID reports whether an ID belongs to a synthesized node (result
pseudo-requirement or test case) rather than an authored requirement.

### IMP-exporter-idsOf-ae4c59d: idsOf
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/analysis.go
x-reqmd.source-line: 298
x-reqmd.symbol-kind: function
trace: []
```

idsOf returns the given document's requirement IDs in document order.

## internal/exporter/boundary.go
### IMP-exporter-DocBoundary-38d6d1a: DocBoundary
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/boundary.go
x-reqmd.source-line: 15
x-reqmd.symbol-kind: type
trace: []
```

DocBoundary describes a document directory's position in the V-model chain.
IsRoot is true when the document has no upstream sources (or is
declared external). IsLeaf is true when no other document references
it as an upstream source.

### IMP-exporter-ComputeDocBoundaries-625fcd1: ComputeDocBoundaries
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/boundary.go
x-reqmd.source-line: 24
x-reqmd.symbol-kind: function
trace: []
```

ComputeDocBoundaries returns the boundary for each document directory path.
A document is a root boundary when it has no upstream sources, or is
declared external. A document is a leaf boundary when it is not
referenced as an upstream source by any other document.

## internal/exporter/boundary_test.go
### IMP-exporter-TestComputeBoundaries_MixedTopology-c690610: TestComputeBoundaries_MixedTopology
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/boundary_test.go
x-reqmd.source-line: 9
x-reqmd.symbol-kind: function
trace: []
```

### IMP-exporter-TestComputeBoundaries_ExternalOverridesSources-d8fb15c: TestComputeBoundaries_ExternalOverridesSources
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/boundary_test.go
x-reqmd.source-line: 55
x-reqmd.symbol-kind: function
trace: []
```

### IMP-exporter-TestComputeBoundaries_NilXReqmd-f1e8644: TestComputeBoundaries_NilXReqmd
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/boundary_test.go
x-reqmd.source-line: 69
x-reqmd.symbol-kind: function
trace: []
```

### IMP-exporter-TestComputeBoundaries_EmptySources-eebf857: TestComputeBoundaries_EmptySources
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/boundary_test.go
x-reqmd.source-line: 81
x-reqmd.symbol-kind: function
trace: []
```

## internal/exporter/browser_test.go
### IMP-exporter-browserFixture-a979123: browserFixture
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/browser_test.go
x-reqmd.source-line: 19
x-reqmd.symbol-kind: function
trace: []
```

browserFixture builds a document whose analysis exercises every state:
covered, missing coverage, a failing verdict, and a draft requirement that
the default status filter hides.

### IMP-exporter-exportBrowserFixture-7f72731: exportBrowserFixture
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/browser_test.go
x-reqmd.source-line: 62
x-reqmd.symbol-kind: function
trace: []
```

exportBrowserFixture writes the fixture page and returns its path.

### IMP-exporter-browserHarness-8506996: browserHarness
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/browser_test.go
x-reqmd.source-line: 83
x-reqmd.symbol-kind: const
trace: []
```

browserHarness drives the exported page and reports results through the
document title, which --dump-dom prints.

### IMP-exporter-deepLinkHarness-092fda4: deepLinkHarness
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/browser_test.go
x-reqmd.source-line: 164
x-reqmd.symbol-kind: const
trace: []
```

### IMP-exporter-restoredStateHarness-8e71cdf: restoredStateHarness
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/browser_test.go
x-reqmd.source-line: 179
x-reqmd.symbol-kind: const
trace: []
```

### IMP-exporter-findChrome-ba3a8cb: findChrome
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/browser_test.go
x-reqmd.source-line: 197
x-reqmd.symbol-kind: function
trace: []
```

findChrome returns the first usable headless-capable browser.

### IMP-exporter-runHeadless-9b7fc94: runHeadless
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/browser_test.go
x-reqmd.source-line: 210
x-reqmd.symbol-kind: function
trace: []
```

runHeadless loads one exported page through file:// with the given harness
appended and returns the decoded RESULT payload.

### IMP-exporter-stringSet-f736881: stringSet
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/browser_test.go
x-reqmd.source-line: 252
x-reqmd.symbol-kind: function
trace: []
```

stringSet converts a decoded JSON array of strings into a set.

### IMP-exporter-assertNoJSErrors-521bcf3: assertNoJSErrors
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/browser_test.go
x-reqmd.source-line: 267
x-reqmd.symbol-kind: function
trace: []
```

### IMP-exporter-TestHTMLExportBrowserSmoke-39a9a50: TestHTMLExportBrowserSmoke
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/browser_test.go
x-reqmd.source-line: 277
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

TestHTMLExportBrowserSmoke is the end-to-end guard for workflows that only
exist in the client: offline load, filtering, sorting, the on-demand views,
the bounded matrix, and keyboard operation.

### IMP-exporter-TestHTMLExportBrowserDeepLink-0c47a85: TestHTMLExportBrowserDeepLink
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/browser_test.go
x-reqmd.source-line: 337
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

TestHTMLExportBrowserDeepLink verifies a #REQ-ID deep link opens Document
mode and reveals the target.

### IMP-exporter-TestHTMLExportBrowserRestoresURLState-6a90c76: TestHTMLExportBrowserRestoresURLState
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/browser_test.go
x-reqmd.source-line: 356
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

TestHTMLExportBrowserRestoresURLState verifies a shared URL reproduces the
exact filtered view, including values with commas and colons.

### IMP-exporter-TestHTMLExportBrowserExplicitEmptyFilters-7a5ba17: TestHTMLExportBrowserExplicitEmptyFilters
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/browser_test.go
x-reqmd.source-line: 376
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

TestHTMLExportBrowserExplicitEmptyFilters verifies `?f=` keeps the document
default status filter cleared after reload.

### IMP-exporter-TestHTMLExportBrowserOutlineResize-a3bc093: TestHTMLExportBrowserOutlineResize
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/browser_test.go
x-reqmd.source-line: 393
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

TestHTMLExportBrowserOutlineResize verifies the outline width handle: arrow
keys change the one width variable the grid consumes, and double-click
restores the default.

### IMP-exporter-TestHTMLExportBrowserMobileOutline-e7e2291: TestHTMLExportBrowserMobileOutline
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/browser_test.go
x-reqmd.source-line: 442
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

TestHTMLExportBrowserMobileOutline verifies the outline becomes a drawer at
mobile width and the toolbar stops sticking when vertical space is tight.

### IMP-exporter-exportNestedFilterFixture-bf5116a: exportNestedFilterFixture
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/browser_test.go
x-reqmd.source-line: 473
x-reqmd.symbol-kind: function
trace: []
```

exportNestedFilterFixture exports a section containing a draft parent
requirement with an approved child, so the default status filter hides the
parent but not the child.

### IMP-exporter-TestHTMLExportBrowserFilteredParentKeepsVisibleChild-072f90e: TestHTMLExportBrowserFilteredParentKeepsVisibleChild
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/browser_test.go
x-reqmd.source-line: 509
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

TestHTMLExportBrowserFilteredParentKeepsVisibleChild verifies that hiding a
parent card does not hide its visible children in the outline, and that a
jump reopens a collapsed section folder.

### IMP-exporter-TestHTMLExportBrowserPopoverAndDetail-e6e0e18: TestHTMLExportBrowserPopoverAndDetail
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/browser_test.go
x-reqmd.source-line: 547
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

TestHTMLExportBrowserPopoverAndDetail covers the two interaction fixes that
only exist in the client: filter menus close like popovers (outside click,
Escape), and an expanded table row shows its detail block already open.

### IMP-exporter-TestHTMLExportBrowserMatrixBound-dc99342: TestHTMLExportBrowserMatrixBound
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/browser_test.go
x-reqmd.source-line: 582
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

TestHTMLExportBrowserMatrixBound verifies the full matrix refuses to draw
beyond its cell limit and points the reviewer at the trace neighbourhood.

### IMP-exporter-TestHTMLExportBrowserLinkSemantics-e6d4d20: TestHTMLExportBrowserLinkSemantics
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/browser_test.go
x-reqmd.source-line: 657
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

TestHTMLExportBrowserLinkSemantics covers the navigation fixes: external
trace nodes are links, requirement references inside analytical views open
the Document view, and the shortcuts dialog receives focus.

### IMP-exporter-TestHTMLExportBrowserContextJumps-61d6857: TestHTMLExportBrowserContextJumps
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/browser_test.go
x-reqmd.source-line: 719
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

TestHTMLExportBrowserContextJumps verifies that jumping from a coverage
queue keeps the reviewer's filters and that the result count names the
reason requirements are hidden.

### IMP-exporter-TestHTMLExportBrowserOutlineSelectsRequirement-9f7b2f3: TestHTMLExportBrowserOutlineSelectsRequirement
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/browser_test.go
x-reqmd.source-line: 760
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

TestHTMLExportBrowserOutlineSelectsRequirement verifies the trace explorer is
driven by the outline (the requirement list the reader already has) instead
of a separate picker, and that an outline click keeps filters and state.

### IMP-exporter-TestHTMLExportBrowserSelectionPersistence-4d61dfe: TestHTMLExportBrowserSelectionPersistence
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/browser_test.go
x-reqmd.source-line: 820
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

TestHTMLExportBrowserSelectionPersistence covers the outline selection rules:
the same requirement stays selected across views, the Trace view is the
selection (not the scroll position), and a Table outline click stays in the
table instead of reopening the reading view.

### IMP-exporter-TestHTMLExportBrowserDocumentJumpScrollsToSelection-24f22f9: TestHTMLExportBrowserDocumentJumpScrollsToSelection
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/browser_test.go
x-reqmd.source-line: 914
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

TestHTMLExportBrowserDocumentJumpScrollsToSelection verifies that entering the
reading view lands on the selected requirement (not the top of the
document), and that the outline's position highlight agrees with it.

### IMP-exporter-TestHTMLExportBrowserJumpClearsStickyChrome-46305f0: TestHTMLExportBrowserJumpClearsStickyChrome
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/browser_test.go
x-reqmd.source-line: 973
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

TestHTMLExportBrowserJumpClearsStickyChrome verifies that a jump (palette,
deep link, or view switch) leaves the requirement heading below the sticky
app bar and toolbar instead of hidden underneath them.

### IMP-exporter-TestHTMLExportBrowserSparseColumnsHidden-4576236: TestHTMLExportBrowserSparseColumnsHidden
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/browser_test.go
x-reqmd.source-line: 1023
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

TestHTMLExportBrowserSparseColumnsHidden verifies that table columns which
are blank or nearly blank across the rows are dropped (an empty column reads
as a bug) while the column chooser still offers them.

### IMP-exporter-TestHTMLExportModHintIsPlatformCorrect-e618d41: TestHTMLExportModHintIsPlatformCorrect
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/browser_test.go
x-reqmd.source-line: 1084
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

TestHTMLExportModHintIsPlatformCorrect verifies the palette shortcut is
spelled for the reader's platform instead of always claiming ⌘K.

### IMP-exporter-TestHTMLExportBrowserPaletteNavigation-b07a5cc: TestHTMLExportBrowserPaletteNavigation
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/browser_test.go
x-reqmd.source-line: 1130
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

TestHTMLExportBrowserPaletteNavigation covers the command palette with a
requirement query: arrow keys step one row at a time, hover does not fight
the keyboard, typing keeps the focus in the input, and Enter runs the
highlighted row.

### IMP-exporter-TestHTMLExportBrowserPaletteFuzzy-c785d42: TestHTMLExportBrowserPaletteFuzzy
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/browser_test.go
x-reqmd.source-line: 1201
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

TestHTMLExportBrowserPaletteFuzzy covers abbreviation matching in the
command palette: "sysfm1" must find SYS-001-style IDs whose segments the
query skips part of, and ids are ranked ahead of incidental title hits.

## internal/exporter/csv.go
### IMP-exporter-CSV-4877f3a: CSV
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/csv.go
x-reqmd.source-line: 16
x-reqmd.symbol-kind: type
trace: []
```

CSV exports document content as comma-separated values. When verdicts
are set via SetVerdicts, extra columns (Verdict, Verdict Source, Verdict
Cases) are appended per row.

#### IMP-exporter-SetVerdicts-5a7abe8: SetVerdicts
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/csv.go
x-reqmd.source-line: 23
x-reqmd.symbol-kind: method
trace: []
```

SetVerdicts stores verification verdicts for CSV export. When non-nil,
extra columns (Verdict, Verdict Source, Verdict Cases) are appended to
the output.

#### IMP-exporter-Export-e9edde0: Export
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/csv.go
x-reqmd.source-line: 27
x-reqmd.symbol-kind: method
trace: []
```

## internal/exporter/docchain.go
### IMP-exporter-BuildTitleMap-9c92b3a: BuildTitleMap
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/docchain.go
x-reqmd.source-line: 12
x-reqmd.symbol-kind: function
trace: []
```

BuildTitleMap builds a reqID → title map across all documents, for use as
trace-link labels. Requirements with an empty title are omitted.

### IMP-exporter-RenderContext-58b5069: RenderContext
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/docchain.go
x-reqmd.source-line: 26
x-reqmd.symbol-kind: type
trace: []
```

RenderContext holds precomputed document chain and link resolution data
for cross-document HTML traceability. It is shared by the export and serve commands.

#### IMP-exporter-ResolveLink-6286435: ResolveLink
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/docchain.go
x-reqmd.source-line: 103
x-reqmd.symbol-kind: method
trace: []
```

ResolveLink returns a function that maps a requirement ID to an HTML
anchor link relative to the given output path. Since all HTML files
are written as flat siblings in the output directory, the relative
path between any two files is just the target filename. The current
file is detected by comparing basenames so absolute vs relative path
mismatches don't cause false negatives.

#### IMP-exporter-BuildChainGraph-f460196: BuildChainGraph
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/docchain.go
x-reqmd.source-line: 156
x-reqmd.symbol-kind: method
trace: []
```

BuildChainGraph constructs the tiered document trace graph for the given doc.

Upstream tiers (negative levels) are computed by BFS outward from the current
doc following upstream.sources references. Docs at the same shortest
distance are grouped into a single horizontal tier so parallel branches at
the same level render side-by-side. External docs (x-reqmd.external: true)
are flagged so the renderer can mark them.

Downstream tiers are computed symmetrically by BFS following inverted
references (docs whose upstream.sources point at the current doc, or any
other already-reached downstream doc).

Self-references and visited nodes are tracked globally to terminate cycles.

### IMP-exporter-docLinkInfo-538ca38: docLinkInfo
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/docchain.go
x-reqmd.source-line: 36
x-reqmd.symbol-kind: type
trace: []
```

### IMP-exporter-absDocInfo-50f98bc: absDocInfo
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/docchain.go
x-reqmd.source-line: 42
x-reqmd.symbol-kind: type
trace: []
```

### IMP-exporter-NewRenderContext-cf49397: NewRenderContext
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/docchain.go
x-reqmd.source-line: 48
x-reqmd.symbol-kind: function
trace: []
```

NewRenderContext precomputes the cross-document link map and trace chain data.

### IMP-exporter-ChainCard-7c51781: ChainCard
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/docchain.go
x-reqmd.source-line: 115
x-reqmd.symbol-kind: type
trace: []
```

ChainCard represents a single document in the chain graph visualization.

### IMP-exporter-ChainTier-41d8237: ChainTier
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/docchain.go
x-reqmd.source-line: 124
x-reqmd.symbol-kind: type
trace: []
```

ChainTier is a horizontal row of docs at the same trace distance.

### IMP-exporter-DocChainGraph-229e95f: DocChainGraph
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/docchain.go
x-reqmd.source-line: 132
x-reqmd.symbol-kind: type
trace: []
```

DocChainGraph is the full tiered document graph for a single current doc.
It is the source of truth for the Confluence-Flow visualization.

#### IMP-exporter-HasContent-031093e: HasContent
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/docchain.go
x-reqmd.source-line: 139
x-reqmd.symbol-kind: method
trace: []
```

HasContent reports whether the graph has any tier that should be rendered.

### IMP-exporter-upstreamTierLabel-257c99f: upstreamTierLabel
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/docchain.go
x-reqmd.source-line: 269
x-reqmd.symbol-kind: function
trace: []
```

### IMP-exporter-downstreamTierLabel-46dba48: downstreamTierLabel
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/docchain.go
x-reqmd.source-line: 273
x-reqmd.symbol-kind: function
trace: []
```

### IMP-exporter-tierLabel-1fa0b78: tierLabel
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/docchain.go
x-reqmd.source-line: 279
x-reqmd.symbol-kind: function
trace: []
```

tierLabel formats a chain tier header: the bare noun at one level of
distance, otherwise "<noun> · N levels <direction>".

### IMP-exporter-eachChainCard-9f44252: eachChainCard
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/docchain.go
x-reqmd.source-line: 289
x-reqmd.symbol-kind: function
trace: []
```

eachChainCard calls fn for every card in the graph's upstream and
downstream tiers (never the current card). Shared by the three card
path-rewriting functions below.

### IMP-exporter-ResolveChainOutputs-b4bf6ab: ResolveChainOutputs
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/docchain.go
x-reqmd.source-line: 309
x-reqmd.symbol-kind: function
trace: []
```

ResolveChainOutputs replaces every non-current card's Path — which the
render context seeds with the flat output file name (e.g.
"01-stakeholder-requirements.html") — with the real output path for that
document, as reported by outPath(dirName). Callers know where each page
is actually written (flat into an output dir, or into each document's own
dir), so they supply the mapping. Relative card paths are then produced
by RelativizeChainGraph.

### IMP-exporter-RelativizeChainGraph-fd140dc: RelativizeChainGraph
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/docchain.go
x-reqmd.source-line: 325
x-reqmd.symbol-kind: function
trace: []
```

RelativizeChainGraph rewrites every non-current card's Path to be
relative to baseDir. This is a free function (not a method) so callers
in other packages can use it without exposing the unexported fields
of DocChainGraph.

### IMP-exporter-BasenameChainGraph-c3014af: BasenameChainGraph
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/docchain.go
x-reqmd.source-line: 339
x-reqmd.symbol-kind: function
trace: []
```

BasenameChainGraph collapses every non-current card's Path to its
basename. Useful when all HTML files are served at a single URL root
(the dev server).

## internal/exporter/docchain_test.go
### IMP-exporter-makeDoc-967426e: makeDoc
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/docchain_test.go
x-reqmd.source-line: 12
x-reqmd.symbol-kind: function
trace: []
```

makeDoc constructs a minimal model.Document with the given path, title and
optional x-reqmd block. Requirements are empty (the chain graph doesn't need
any) and schema is the bare map containing the title.

### IMP-exporter-makeXReqmd-f261749: makeXReqmd
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/docchain_test.go
x-reqmd.source-line: 21
x-reqmd.symbol-kind: function
trace: []
```

makeXReqmd is a tiny constructor for XReqmd pointer values.

### IMP-exporter-TestBuildChainGraphSimpleLinear-fdb374a: TestBuildChainGraphSimpleLinear
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/docchain_test.go
x-reqmd.source-line: 29
x-reqmd.symbol-kind: function
trace: []
```

### IMP-exporter-TestBuildChainGraphParallelBranches-fec9442: TestBuildChainGraphParallelBranches
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/docchain_test.go
x-reqmd.source-line: 64
x-reqmd.symbol-kind: function
trace: []
```

### IMP-exporter-TestBuildChainGraphExternalFlag-b8a454c: TestBuildChainGraphExternalFlag
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/docchain_test.go
x-reqmd.source-line: 93
x-reqmd.symbol-kind: function
trace: []
```

### IMP-exporter-TestBuildChainGraphCycle-0f31159: TestBuildChainGraphCycle
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/docchain_test.go
x-reqmd.source-line: 112
x-reqmd.symbol-kind: function
trace: []
```

### IMP-exporter-TestBuildChainGraphDeeperLevels-ca5d1c3: TestBuildChainGraphDeeperLevels
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/docchain_test.go
x-reqmd.source-line: 134
x-reqmd.symbol-kind: function
trace: []
```

### IMP-exporter-TestBuildChainGraphNoUpstream-8342a2f: TestBuildChainGraphNoUpstream
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/docchain_test.go
x-reqmd.source-line: 165
x-reqmd.symbol-kind: function
trace: []
```

### IMP-exporter-TestResolveAndRelativizeChainOutputs-012d207: TestResolveAndRelativizeChainOutputs
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/docchain_test.go
x-reqmd.source-line: 181
x-reqmd.symbol-kind: function
trace: []
```

## internal/exporter/exporter_test.go
### IMP-exporter-schemaWithTitle-be55a35: schemaWithTitle
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/exporter_test.go
x-reqmd.source-line: 11
x-reqmd.symbol-kind: function
trace: []
```

### IMP-exporter-TestCSVExport-2cda038: TestCSVExport
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/exporter_test.go
x-reqmd.source-line: 25
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-CSV-4877f3a
```

### IMP-exporter-TestCSVExportSpecialTypes-8b79af8: TestCSVExportSpecialTypes
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/exporter_test.go
x-reqmd.source-line: 87
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-CSV-4877f3a
```

### IMP-exporter-TestHTMLExportDocChain-b22f386: TestHTMLExportDocChain
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/exporter_test.go
x-reqmd.source-line: 136
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

### IMP-exporter-TestHTMLExportDocChainExternal-6ed6705: TestHTMLExportDocChainExternal
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/exporter_test.go
x-reqmd.source-line: 249
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

### IMP-exporter-TestHTMLExportDocChainDeep-aebe84c: TestHTMLExportDocChainDeep
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/exporter_test.go
x-reqmd.source-line: 302
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

### IMP-exporter-TestHTMLExportDocChainCrowded-470a28a: TestHTMLExportDocChainCrowded
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/exporter_test.go
x-reqmd.source-line: 366
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

### IMP-exporter-TestHTMLExportNoDocLinks-dbc8d25: TestHTMLExportNoDocLinks
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/exporter_test.go
x-reqmd.source-line: 405
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

### IMP-exporter-TestCSVExportItems-db45cdc: TestCSVExportItems
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/exporter_test.go
x-reqmd.source-line: 433
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-CSV-4877f3a
```

### IMP-exporter-TestCSVExportVerdicts-e321c76: TestCSVExportVerdicts
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/exporter_test.go
x-reqmd.source-line: 466
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-CSV-4877f3a
```

## internal/exporter/graph.go
### IMP-exporter-ExportGraph-394530c: ExportGraph
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/graph.go
x-reqmd.source-line: 25
x-reqmd.symbol-kind: function
trace: []
```

ExportGraph writes all parsed requirements as a ladybugdb database at outDir.
Each requirement becomes a node with id and file properties. Every trace
link becomes a TracesTo edge between Requirement nodes.

The resulting database is at <outDir>/reqmd-graph.lbug/ and can be queried
with the ladybug CLI or SQL ODBC driver.

### IMP-exporter-escapeCypherString-a245d4b: escapeCypherString
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/graph.go
x-reqmd.source-line: 122
x-reqmd.symbol-kind: function
trace: []
```

escapeCypherString escapes single quotes by doubling them for safe inline
Cypher string literals.

## internal/exporter/html.go
### IMP-exporter-staticFS-5a7bdab: staticFS
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html.go
x-reqmd.source-line: 27
x-reqmd.symbol-kind: var
trace: []
```

go:embed static

### IMP-exporter-mustReadStatic-35352d1: mustReadStatic
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html.go
x-reqmd.source-line: 38
x-reqmd.symbol-kind: function
trace: []
```

### IMP-exporter-EvidenceItem-24b28ec: EvidenceItem
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html.go
x-reqmd.source-line: 48
x-reqmd.symbol-kind: type
trace: []
```

EvidenceItem is one verification result contributing to a measure's
rolled-up evidence set, ready for rendering in exports.

### IMP-exporter-VerdictInfo-3d839d7: VerdictInfo
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html.go
x-reqmd.source-line: 58
x-reqmd.symbol-kind: type
trace: []
```

VerdictInfo carries the rolled-up verification verdict for a single
measure: the outcome, a representative source file, the case keys that
contributed (for CSV), and the full evidence list (for HTML).

### IMP-exporter-HTML-1ac7457: HTML
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html.go
x-reqmd.source-line: 67
x-reqmd.symbol-kind: type
trace: []
```

HTML exports requirements as standalone HTML: a server-rendered Document
view plus on-demand analytical views driven by one shared view model.

#### IMP-exporter-SetAnalysis-5f0cad9: SetAnalysis
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html.go
x-reqmd.source-line: 77
x-reqmd.symbol-kind: method
trace: []
```

SetAnalysis supplies the graph-derived coverage, verification, and link
state every view renders. A nil analysis renders requirements without
semantic status (still valid output, just less information).

#### IMP-exporter-SetResolveLink-abfe895: SetResolveLink
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html.go
x-reqmd.source-line: 83
x-reqmd.symbol-kind: method
trace: []
```

SetResolveLink supplies the cross-document href resolver. A nil resolver
produces same-page fragments.

#### IMP-exporter-SetDocChainGraph-b72256c: SetDocChainGraph
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html.go
x-reqmd.source-line: 206
x-reqmd.symbol-kind: method
trace: []
```

SetDocChainGraph configures the tiered document graph used by the
document-chain navigation. Pass a zero-value graph to disable the chain.

#### IMP-exporter-SetBoundary-be634e9: SetBoundary
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html.go
x-reqmd.source-line: 212
x-reqmd.symbol-kind: method
trace: []
```

SetBoundary configures the document-level boundary flags (root/leaf
position in the V-model chain) used by the header badges.

#### IMP-exporter-Export-9012d07: Export
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html.go
x-reqmd.source-line: 222
x-reqmd.symbol-kind: method
trace: []
```

Export writes requirements as a standalone HTML page. The Document view is
server-rendered and fully visible without JavaScript; Table, Trace, and
Coverage render on demand from the embedded view model.

### IMP-exporter-schemaTitle-9cb6a59: schemaTitle
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html.go
x-reqmd.source-line: 89
x-reqmd.symbol-kind: function
trace: []
```

schemaTitle safely extracts the title field from a document's parsed
schema. Returns "" when Schema is nil, not a map, or has no string title.

### IMP-exporter-docTitle-4498ba2: docTitle
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html.go
x-reqmd.source-line: 100
x-reqmd.symbol-kind: function
trace: []
```

docTitle returns the title used for rendering a document page or card:
the parsed file h1 (document title), falling back to the schema title.

### IMP-exporter-faviconSVG-fbc5520: faviconSVG
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html.go
x-reqmd.source-line: 110
x-reqmd.symbol-kind: const
trace: []
```

faviconSVG is the site's brand mark inlined as a data URI, so an exported
page carries an icon without a second file or a network request. The CSS
media query keeps it legible on dark tab bars.

### IMP-exporter-favicon-31df4ba: favicon
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html.go
x-reqmd.source-line: 113
x-reqmd.symbol-kind: var
trace: []
```

favicon is the data URI written into every exported page's <head>.

### IMP-exporter-mermaidCDN-02784fc: mermaidCDN
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html.go
x-reqmd.source-line: 120
x-reqmd.symbol-kind: const
trace: []
```

mermaidCDN is the pinned Mermaid UMD runtime that a page containing a diagram
loads, and the only external request an exported page ever makes (SW-EXP-004).
The UMD build is used, not the ESM module: ES modules are blocked by CORS on
file://, and an exported tree must open offline. Keep the version in step
with the documentation site's pin.

### IMP-exporter-bodyParser-cc1ffb9: bodyParser
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html.go
x-reqmd.source-line: 126
x-reqmd.symbol-kind: var
trace: []
```

bodyParser and bodyRenderer convert stored body Markdown to HTML. goldmark v2
separates parsing from rendering, so each document body is parsed once and
rendered with the presentation extensions: GitHub alerts, emoji shortcodes,
chroma highlighting, and client-side Mermaid diagrams.

### IMP-exporter-bodyRenderer-d0402f5: bodyRenderer
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html.go
x-reqmd.source-line: 133
x-reqmd.symbol-kind: var
trace: []
```

### IMP-exporter-mermaidInitThemed-86e4c20: mermaidInitThemed
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html.go
x-reqmd.source-line: 158
x-reqmd.symbol-kind: const
trace: []
```

goldmark-diagram emits a fixed mermaid.initialize call with no theme, so a
diagram would always render in Mermaid's light palette. It also appends the
runtime script to every rendered body that holds a diagram, so a page with
several diagrams would load the pinned runtime once per diagram.

renderMarkdown therefore strips the emitted runtime block and the page
emits it once, themed, when the document contains a diagram. The coupling to
the upstream strings is deliberate: if they stop matching, every diagram
loads its own runtime again in the light palette, which is what
TestHTMLExportMermaidDiagram guards against.

### IMP-exporter-mermaidInitUpstream-40da68e: mermaidInitUpstream
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html.go
x-reqmd.source-line: 158
x-reqmd.symbol-kind: const
trace: []
```

goldmark-diagram emits a fixed mermaid.initialize call with no theme, so a
diagram would always render in Mermaid's light palette. It also appends the
runtime script to every rendered body that holds a diagram, so a page with
several diagrams would load the pinned runtime once per diagram.

renderMarkdown therefore strips the emitted runtime block and the page
emits it once, themed, when the document contains a diagram. The coupling to
the upstream strings is deliberate: if they stop matching, every diagram
loads its own runtime again in the light palette, which is what
TestHTMLExportMermaidDiagram guards against.

### IMP-exporter-mermaidRuntimeOnce-a0a50a6: mermaidRuntimeOnce
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html.go
x-reqmd.source-line: 158
x-reqmd.symbol-kind: const
trace: []
```

goldmark-diagram emits a fixed mermaid.initialize call with no theme, so a
diagram would always render in Mermaid's light palette. It also appends the
runtime script to every rendered body that holds a diagram, so a page with
several diagrams would load the pinned runtime once per diagram.

renderMarkdown therefore strips the emitted runtime block and the page
emits it once, themed, when the document contains a diagram. The coupling to
the upstream strings is deliberate: if they stop matching, every diagram
loads its own runtime again in the light palette, which is what
TestHTMLExportMermaidDiagram guards against.

### IMP-exporter-mermaidRuntimeUpstream-e9195a6: mermaidRuntimeUpstream
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html.go
x-reqmd.source-line: 158
x-reqmd.symbol-kind: const
trace: []
```

goldmark-diagram emits a fixed mermaid.initialize call with no theme, so a
diagram would always render in Mermaid's light palette. It also appends the
runtime script to every rendered body that holds a diagram, so a page with
several diagrams would load the pinned runtime once per diagram.

renderMarkdown therefore strips the emitted runtime block and the page
emits it once, themed, when the document contains a diagram. The coupling to
the upstream strings is deliberate: if they stop matching, every diagram
loads its own runtime again in the light palette, which is what
TestHTMLExportMermaidDiagram guards against.

### IMP-exporter-renderMarkdown-c4860fb: renderMarkdown
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html.go
x-reqmd.source-line: 167
x-reqmd.symbol-kind: function
trace: []
```

renderMarkdown renders markdown content to safe HTML.

### IMP-exporter-hasDiagram-68b1023: hasDiagram
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html.go
x-reqmd.source-line: 182
x-reqmd.symbol-kind: function
trace: []
```

hasDiagram reports whether a document body carries a mermaid fence, which is
what makes the page load the pinned runtime.

### IMP-exporter-nodesHaveDiagram-1832692: nodesHaveDiagram
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html.go
x-reqmd.source-line: 191
x-reqmd.symbol-kind: function
trace: []
```

### IMP-exporter-pageBufferSize-7c8ec5e: pageBufferSize
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html.go
x-reqmd.source-line: 217
x-reqmd.symbol-kind: const
trace: []
```

pageBufferSize is the buffered-writer size for one exported page.

### IMP-exporter-renderModelJSON-0b90809: renderModelJSON
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html.go
x-reqmd.source-line: 265
x-reqmd.symbol-kind: function
trace: []
```

renderModelJSON writes the shared view model the client renders from.

### IMP-exporter-renderScript-f1c838d: renderScript
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html.go
x-reqmd.source-line: 270
x-reqmd.symbol-kind: function
trace: []
```

### IMP-exporter-formatAttrLabel-8be58cc: formatAttrLabel
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html.go
x-reqmd.source-line: 278
x-reqmd.symbol-kind: function
trace: []
```

formatAttrLabel converts a snake_case or kebab-case attribute name to
a human-readable title-case label (e.g. "disposition-reason" → "Disposition Reason").

### IMP-exporter-formatAttrHTML-1cde9fb: formatAttrHTML
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html.go
x-reqmd.source-line: 291
x-reqmd.symbol-kind: function
trace: []
```

formatAttrHTML renders an attribute value as rich HTML for the card grid.
Arrays become inline pill tags. Nested maps become a compact <pre> JSON block.
All output is already HTML-escaped where needed by the caller.

### IMP-exporter-renderChainGraph-ec41759: renderChainGraph
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html.go
x-reqmd.source-line: 346
x-reqmd.symbol-kind: function
trace: []
```

renderChainGraph writes the Document Stack chain as HTML to b.
The output is pure HTML/CSS — no JavaScript.

### IMP-exporter-renderChainZone-347b221: renderChainZone
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html.go
x-reqmd.source-line: 368
x-reqmd.symbol-kind: function
trace: []
```

renderChainZone writes one horizontal zone (upstream or downstream)
with its header and stacked tiers.

### IMP-exporter-renderChainSeparator-dfd1416: renderChainSeparator
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html.go
x-reqmd.source-line: 386
x-reqmd.symbol-kind: function
trace: []
```

renderChainSeparator writes the chevron rule between chain zones.

### IMP-exporter-renderChainCard-92666bd: renderChainCard
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html.go
x-reqmd.source-line: 395
x-reqmd.symbol-kind: function
trace: []
```

renderChainCard writes one <a> (or <span> for the current doc) card.

## internal/exporter/html_test.go
### IMP-exporter-analysisFor-931d377: analysisFor
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html_test.go
x-reqmd.source-line: 15
x-reqmd.symbol-kind: function
trace: []
```

analysisFor builds the presentation analysis from a document set.

### IMP-exporter-exportPage-753db53: exportPage
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html_test.go
x-reqmd.source-line: 25
x-reqmd.symbol-kind: function
trace: []
```

exportPage renders one document and returns the HTML.

### IMP-exporter-pageModelOf-5578d77: pageModelOf
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html_test.go
x-reqmd.source-line: 35
x-reqmd.symbol-kind: function
trace: []
```

pageModelOf extracts and decodes the embedded page model.

### IMP-exporter-basicDoc-20af48e: basicDoc
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html_test.go
x-reqmd.source-line: 49
x-reqmd.symbol-kind: function
trace: []
```

### IMP-exporter-TestHTMLExport-ea186e5: TestHTMLExport
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html_test.go
x-reqmd.source-line: 75
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

### IMP-exporter-TestHTMLExportSectionFolderWrapsChildren-7b2c3da: TestHTMLExportSectionFolderWrapsChildren
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html_test.go
x-reqmd.source-line: 159
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

TestHTMLExportSectionFolderWrapsChildren pins the folder contract: a section
with children is a collapsible details block that actually contains them, so
clicking the summary folds the content.

### IMP-exporter-TestHTMLExportRequirementChildrenGrouped-cdfe3af: TestHTMLExportRequirementChildrenGrouped
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html_test.go
x-reqmd.source-line: 190
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

TestHTMLExportRequirementChildrenGrouped pins that a requirement's children
(methods, sub-requirements, doc sections) sit in one container after it, so
the content frame shows the same hierarchy as the outline.

### IMP-exporter-TestHTMLExportHeaderEyebrow-fadd28b: TestHTMLExportHeaderEyebrow
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html_test.go
x-reqmd.source-line: 228
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

TestHTMLExportHeaderEyebrow pins the document identity line above the title.

### IMP-exporter-TestHTMLExportOfflineContent-8b4836d: TestHTMLExportOfflineContent
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html_test.go
x-reqmd.source-line: 241
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

TestHTMLExportOfflineContent is the Phase 0/1 regression guard: primary
content is visible by default, and nothing loads from a CDN.

### IMP-exporter-TestHTMLExportHTMLEscaping-3a09054: TestHTMLExportHTMLEscaping
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html_test.go
x-reqmd.source-line: 271
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

### IMP-exporter-TestHTMLExportModelIsHTMLSafe-a5d99de: TestHTMLExportModelIsHTMLSafe
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html_test.go
x-reqmd.source-line: 302
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

TestHTMLExportModelIsHTMLSafe pins that author-controlled values in the
embedded JSON cannot break out of the script element.

### IMP-exporter-TestHTMLExportModelOmitsBodies-42a84a0: TestHTMLExportModelOmitsBodies
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html_test.go
x-reqmd.source-line: 328
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

TestHTMLExportModelOmitsBodies verifies the duplicate-data fix: the reading
view is the only representation of requirement prose.

### IMP-exporter-TestHTMLExportTraceAttributeNotInGrid-76c925c: TestHTMLExportTraceAttributeNotInGrid
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html_test.go
x-reqmd.source-line: 345
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

### IMP-exporter-TestHTMLExportArrayPillsAndNestedObjects-145c708: TestHTMLExportArrayPillsAndNestedObjects
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html_test.go
x-reqmd.source-line: 358
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

### IMP-exporter-TestHTMLExportItems-820ff8f: TestHTMLExportItems
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html_test.go
x-reqmd.source-line: 399
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

### IMP-exporter-TestHTMLExport_UndeclaredAttributeRendered-7cd4750: TestHTMLExport_UndeclaredAttributeRendered
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html_test.go
x-reqmd.source-line: 447
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

TestHTMLExport_UndeclaredAttributeRendered verifies undeclared attributes
still render (flagged) and feed the filter menu.

### IMP-exporter-TestHTMLExportTraceLinkState-864d49b: TestHTMLExportTraceLinkState
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html_test.go
x-reqmd.source-line: 486
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

TestHTMLExportTraceLinkState checks pins, versions, and link health in the
shared model and the server-rendered links.

### IMP-exporter-TestHTMLExportDownstreamLinkDirection-1c8effe: TestHTMLExportDownstreamLinkDirection
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html_test.go
x-reqmd.source-line: 545
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

TestHTMLExportDownstreamLinkDirection pins that a requirement's inbound
links appear as Down (who traces to it), never as Up.

### IMP-exporter-TestMatrixTooltipDirection-ff8c0eb: TestMatrixTooltipDirection
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html_test.go
x-reqmd.source-line: 588
x-reqmd.symbol-kind: function
trace: []
```

TestMatrixTooltipDirection pins the client's direction-specific matrix
tooltip, the regression from the old downstream matrix.

### IMP-exporter-TestHTMLExportCoverageStates-20de61d: TestHTMLExportCoverageStates
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html_test.go
x-reqmd.source-line: 601
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

TestHTMLExportCoverageStates checks that coverage comes from graph
semantics, not link presence: a draft downstream does not cover.

### IMP-exporter-TestHTMLExportNoVerdictIsFirstClass-5de0635: TestHTMLExportNoVerdictIsFirstClass
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html_test.go
x-reqmd.source-line: 638
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

TestHTMLExportNoVerdictIsFirstClass verifies unverified eligible measures
show a no-verdict chip and are marked eligible in the model, so the client
denominator includes them.

### IMP-exporter-TestHTMLExportEvidenceList-248536c: TestHTMLExportEvidenceList
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html_test.go
x-reqmd.source-line: 682
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

TestHTMLExportEvidenceList keeps the rolled-up evidence rendering.

### IMP-exporter-TestHTMLExportLargeDocumentBound-184d97b: TestHTMLExportLargeDocumentBound
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html_test.go
x-reqmd.source-line: 715
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

TestHTMLExportLargeDocumentBound guards against quadratic hidden markup:
a large document's output grows linearly and emits no matrix/table DOM.

### IMP-exporter-TestProjectIndex-5f32171: TestProjectIndex
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html_test.go
x-reqmd.source-line: 759
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-ProjectIndex-cbf0bb1
```

TestProjectIndex checks the project home page content and queues.

### IMP-exporter-TestHTMLExportMermaidDiagram-1448465: TestHTMLExportMermaidDiagram
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html_test.go
x-reqmd.source-line: 810
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

TestHTMLExportMermaidDiagram covers diagram rendering: the fence becomes a
client-rendered container carrying the pinned runtime, and a page without a
diagram stays free of that external request. It also pins the theme handoff:
goldmark-diagram's hard-coded light-only initialise call must be rewritten to
pick the palette from the `dark` class, and app.js re-renders on toggle. The
client bundle carries the click-to-enlarge wiring, which clones the rendered
svg into an overlay instead of re-rendering or moving it.

### IMP-exporter-TestHTMLExportAlert-1fcb599: TestHTMLExportAlert
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/html_test.go
x-reqmd.source-line: 852
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-HTML-1ac7457
```

TestHTMLExportAlert covers the extension that replaced fenced divs: a GitHub
alert block in a body renders as a styled callout with its kind-specific
class, not as a plain blockquote.

## internal/exporter/json.go
### IMP-exporter-JSON-210eb34: JSON
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/json.go
x-reqmd.source-line: 25
x-reqmd.symbol-kind: type
trace: []
```

JSON exports the parsed tree as NDJSON tables. Verdicts set via
SetVerdicts are emitted by ExportResults; the graph passed to the
Export* methods is the single source of truth for trace edges,
verdicts, and findings.

#### IMP-exporter-SetVerdicts-7642e8a: SetVerdicts
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/json.go
x-reqmd.source-line: 30
x-reqmd.symbol-kind: method
trace: []
```

SetVerdicts stores verification verdicts for JSON export, mirroring CSV.

#### IMP-exporter-ExportRequirements-0cdaae2: ExportRequirements
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/json.go
x-reqmd.source-line: 86
x-reqmd.symbol-kind: method
trace:
  - SW-EXP-003
```

reqmd:trace SW-EXP-003
ExportRequirements writes one NDJSON row per requirement across all authored
documents. Sections are structure, not facts, so they are not rows: their
titles reach the fact table as the `section_path` column on the requirements
beneath them. Synthetic documents (verification results and test cases) are
skipped: the requirement table is authored truth.

ordinals maps every requirement ID to its 1-based position in the
unfiltered document order, so a --filtered export keeps the authored
numbering instead of renumbering the surviving rows.

#### IMP-exporter-ExportTraces-87ecca4: ExportTraces
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/json.go
x-reqmd.source-line: 158
x-reqmd.symbol-kind: method
trace: []
```

ExportTraces writes one NDJSON row per declared trace reference, read
from the raw `trace` attribute values and resolved against the graph.
Both are needed: the attribute carries the declared intent (including
dangling refs), the graph carries resolution.

#### IMP-exporter-ExportResults-d7ba418: ExportResults
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/json.go
x-reqmd.source-line: 235
x-reqmd.symbol-kind: method
trace: []
```

ExportResults writes one NDJSON row per measure with a non-empty
evidence set. Verdicts come from the graph's strict roll-up over a
measure's own results and its approved downstream test cases, so the
value matches what HTML badges and CSV columns show.

#### IMP-exporter-ExportFindings-1e3ee58: ExportFindings
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/json.go
x-reqmd.source-line: 288
x-reqmd.symbol-kind: method
trace: []
```

ExportFindings writes one NDJSON row per graph check finding (the same
results the text and JSON reporters print). Rows are ordered by
requirement, then code, then message so repeated exports are identical.

### IMP-exporter-promotedAttrs-6e2ebea: promotedAttrs
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/json.go
x-reqmd.source-line: 37
x-reqmd.symbol-kind: var
trace: []
```

promotedAttrs maps a built-in requirement attribute to the top-level
NDJSON key it is promoted to, so warehouse SQL can read it directly
instead of digging into the attrs envelope.

### IMP-exporter-jsonRequirement-5e44e8f: jsonRequirement
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/json.go
x-reqmd.source-line: 55
x-reqmd.symbol-kind: type
trace: []
```

jsonRequirement is one row of requirements.ndjson. Built-in attributes
are promoted to top-level keys; the remaining declared attributes stay
in Attrs so nothing is lost.

Every fixed key is present in every row so warehouse schema inference —
DuckDB, Spark, Snowflake's infer_schema — sees the full column set
regardless of which row it samples. String fields are "" when unset;
collections, pin, version, and Attrs are null when unset.

### IMP-exporter-jsonTrace-7a728a5: jsonTrace
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/json.go
x-reqmd.source-line: 144
x-reqmd.symbol-kind: type
trace: []
```

jsonTrace is one row of traces.ndjson: one declared trace reference.
To is the target requirement ID and never carries a `doc-id/` qualifier:
DocID holds the qualifier on its own, and Ref always holds the raw form
as written. Resolved is false for a dangling reference; such a ref also
appears in findings.ndjson, so consumers can count or filter it themselves.

### IMP-exporter-jsonCase-0ceebc8: jsonCase
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/json.go
x-reqmd.source-line: 215
x-reqmd.symbol-kind: type
trace: []
```

jsonCase is one contributing verification result in a measure's
evidence set.

### IMP-exporter-jsonResult-21d5a0d: jsonResult
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/json.go
x-reqmd.source-line: 223
x-reqmd.symbol-kind: type
trace: []
```

jsonResult is one row of results.ndjson: one rolled-up measure verdict.

### IMP-exporter-jsonFinding-13fc831: jsonFinding
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/json.go
x-reqmd.source-line: 275
x-reqmd.symbol-kind: type
trace: []
```

jsonFinding is one row of findings.ndjson, mirroring the reporter's
CheckResult fields. ERROR and WARNING both appear; consumers filter.

### IMP-exporter-remainingAttrs-dbbe94e: remainingAttrs
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/json.go
x-reqmd.source-line: 325
x-reqmd.symbol-kind: function
trace: []
```

remainingAttrs returns a copy of attrs without the promoted built-in
keys, so a value never appears both top-level and in the envelope.

### IMP-exporter-jsonVersion-ca75da8: jsonVersion
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/json.go
x-reqmd.source-line: 344
x-reqmd.symbol-kind: function
trace: []
```

jsonVersion renders the `version` attribute as a decimal string so the
column has one type regardless of the YAML scalar, or null when absent.

## internal/exporter/json_test.go
### IMP-exporter-jsonDocs-2b7e0bb: jsonDocs
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/json_test.go
x-reqmd.source-line: 17
x-reqmd.symbol-kind: function
trace: []
```

jsonDocs builds a small two-document tree: one upstream with a version,
one downstream that traces to it with a pin and a dangling ref. Enough to
exercise every table.

### IMP-exporter-jsonOrdinals-4072e66: jsonOrdinals
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/json_test.go
x-reqmd.source-line: 103
x-reqmd.symbol-kind: function
trace: []
```

jsonOrdinals mirrors the unfiltered ordinal map the CLI builds: every
requirement ID maps to its 1-based position in authored document order.

### IMP-exporter-encodeRows-d3b8475: encodeRows
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/json_test.go
x-reqmd.source-line: 113
x-reqmd.symbol-kind: function
trace: []
```

### IMP-exporter-TestJSONExportRequirements-57bc282: TestJSONExportRequirements
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/json_test.go
x-reqmd.source-line: 138
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-JSON-210eb34
```

### IMP-exporter-TestJSONExportTraces-5ab6144: TestJSONExportTraces
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/json_test.go
x-reqmd.source-line: 217
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-JSON-210eb34
```

### IMP-exporter-TestJSONExportFindings-e5864f8: TestJSONExportFindings
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/json_test.go
x-reqmd.source-line: 281
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-JSON-210eb34
```

### IMP-exporter-TestJSONExportOrdinalStableUnderFilter-5d17068: TestJSONExportOrdinalStableUnderFilter
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/json_test.go
x-reqmd.source-line: 324
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-JSON-210eb34
```

### IMP-exporter-TestJSONExportDeterministic-4e5c108: TestJSONExportDeterministic
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/json_test.go
x-reqmd.source-line: 352
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-JSON-210eb34
```

### IMP-exporter-TestJSONExportStableColumns-7e6119c: TestJSONExportStableColumns
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/json_test.go
x-reqmd.source-line: 385
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-JSON-210eb34
```

### IMP-exporter-TestJSONExportResults-8dd3038: TestJSONExportResults
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/json_test.go
x-reqmd.source-line: 461
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-JSON-210eb34
```

### IMP-exporter-TestJSONExportResultsWithoutResults-dde095a: TestJSONExportResultsWithoutResults
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/json_test.go
x-reqmd.source-line: 507
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-JSON-210eb34
```

## internal/exporter/model.go
### IMP-exporter-filterSet-116d20c: filterSet
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/model.go
x-reqmd.source-line: 13
x-reqmd.symbol-kind: type
trace: []
```

filterSet is the per-attribute sorted unique values collected across
all requirements, used to populate the search toolbar filter menus.

### IMP-exporter-pageLink-3b6e905: pageLink
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/model.go
x-reqmd.source-line: 17
x-reqmd.symbol-kind: type
trace: []
```

pageLink is one trace link in the client model. Href is resolved for the
page that embeds the model; Unresolved links carry the authored Ref only.

### IMP-exporter-pageNode-1d781d9: pageNode
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/model.go
x-reqmd.source-line: 32
x-reqmd.symbol-kind: type
trace: []
```

pageNode is one content node in the shared view model. Body and rationale
are deliberately absent: the server-rendered Document view is their single
representation and analytical views read them from the DOM.

### IMP-exporter-pageDocMeta-f12c074: pageDocMeta
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/model.go
x-reqmd.source-line: 46
x-reqmd.symbol-kind: type
trace: []
```

pageDocMeta carries the document facts the client toolbar and views show.

### IMP-exporter-pageModel-86c2446: pageModel
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/model.go
x-reqmd.source-line: 56
x-reqmd.symbol-kind: type
trace: []
```

pageModel is the single view model for one exported document page. Every
analytical view, count, aggregate, and outline is derived from it, and the
filters are the only filtering input.

### IMP-exporter-renderData-77ae77e: renderData
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/model.go
x-reqmd.source-line: 68
x-reqmd.symbol-kind: type
trace: []
```

renderData is the single O(n) intermediate produced by buildRenderData: it
powers the server-rendered TOC, the filter menus, and the view model.

### IMP-exporter-buildRenderData-e6e8856: buildRenderData
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/model.go
x-reqmd.source-line: 83
x-reqmd.symbol-kind: function
trace: []
```

buildRenderData runs a single O(n) pass over the document content tree to
produce the per-document data the rendering helpers need: the filter set,
the per-status breakdown, and the stable DOM anchors.

Every node gets a stable anchor: requirements anchor on their ID,
containers and info items on a generated "info-N" anchor (N = document
order sequence). The same anchors are used by the TOC and by the rendered
DOM so scroll-spy stays consistent.

### IMP-exporter-buildPageModel-c285b4e: buildPageModel
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/model.go
x-reqmd.source-line: 140
x-reqmd.symbol-kind: function
trace: []
```

buildPageModel turns the document tree and graph analysis into the shared
view model. resolveLink maps a requirement ID to a cross-document href; a
nil resolver yields same-page fragments. propOrder is the schema property
order used to order the table's attribute columns.

### IMP-exporter-sectionSlug-a859f7e: sectionSlug
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/model.go
x-reqmd.source-line: 157
x-reqmd.symbol-kind: function
trace: []
```

sectionSlug turns a section title into a stable DOM anchor: lowercase ASCII
letters and digits, runs of anything else collapsed to a single hyphen. An
untitled section (the document preamble) becomes "sec-preamble".

### IMP-exporter-pageMeta-17405bc: pageMeta
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/model.go
x-reqmd.source-line: 186
x-reqmd.symbol-kind: function
trace: []
```

pageMeta fills the document-level model fields.

### IMP-exporter-modelHref-9521b8a: modelHref
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/model.go
x-reqmd.source-line: 210
x-reqmd.symbol-kind: function
trace: []
```

modelHref resolves cross-document links for the model, falling back to an
empty href (the client uses the same-page anchor).

### IMP-exporter-pageNodeFor-04547bc: pageNodeFor
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/model.go
x-reqmd.source-line: 221
x-reqmd.symbol-kind: function
trace: []
```

pageNodeFor converts one content node, attaching graph-derived state for
requirement nodes.

### IMP-exporter-attributeColumns-0320a2d: attributeColumns
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/model.go
x-reqmd.source-line: 247
x-reqmd.symbol-kind: function
trace: []
```

attributeColumns returns every attribute present on at least one
requirement, in schema property order followed by undeclared extras A→Z.
`trace` and `status` are excluded: both have dedicated presentation
(trace link columns and the status chip/column).

### IMP-exporter-attrsForModel-78c252a: attrsForModel
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/model.go
x-reqmd.source-line: 281
x-reqmd.symbol-kind: function
trace: []
```

attrsForModel copies a requirement's attributes for the client model,
dropping `trace` (represented by the Up links).

### IMP-exporter-pageLinks-fceb382: pageLinks
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/model.go
x-reqmd.source-line: 298
x-reqmd.symbol-kind: function
trace: []
```

pageLinks converts analysis links for JSON. Same-page links resolve to
"#id", which needs no href in the model; links to another document keep
their page-relative target so the client can navigate to them.

### IMP-exporter-sortedKeys-081ec53: sortedKeys
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/model.go
x-reqmd.source-line: 327
x-reqmd.symbol-kind: function
trace: []
```

sortedKeys returns the keys of a string-keyed map in sorted order.

## internal/exporter/project.go
### IMP-exporter-ProjectDoc-975307f: ProjectDoc
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/project.go
x-reqmd.source-line: 17
x-reqmd.symbol-kind: type
trace: []
```

ProjectDoc is one document row on the project home page.

### IMP-exporter-ProjectIndex-cbf0bb1: ProjectIndex
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/project.go
x-reqmd.source-line: 32
x-reqmd.symbol-kind: type
trace: []
```

ProjectIndex is the project-level home page: document discovery,
project-wide health, actionable queues, and cross-document search.

#### IMP-exporter-Export-8c71d4d: Export
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/project.go
x-reqmd.source-line: 70
x-reqmd.symbol-kind: method
trace: []
```

Export writes the project home page.

#### IMP-exporter-renderTiles-72bbbbc: renderTiles
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/project.go
x-reqmd.source-line: 109
x-reqmd.symbol-kind: method
trace: []
```

renderTiles writes the project-wide summary tiles.

#### IMP-exporter-renderSearch-057694b: renderSearch
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/project.go
x-reqmd.source-line: 187
x-reqmd.symbol-kind: method
trace: []
```

renderSearch writes the cross-document search input and result list.

#### IMP-exporter-queue-bbc3f76: queue
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/project.go
x-reqmd.source-line: 207
x-reqmd.symbol-kind: method
trace: []
```

queue builds one queue across every document, capped to keep the page
bounded. The full count is returned so the section can say what was cut.

#### IMP-exporter-renderQueues-c87e993: renderQueues
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/project.go
x-reqmd.source-line: 250
x-reqmd.symbol-kind: method
trace: []
```

renderQueues writes the actionable exception lists.

#### IMP-exporter-renderQueue-081c8d2: renderQueue
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/project.go
x-reqmd.source-line: 361
x-reqmd.symbol-kind: method
trace: []
```

renderQueue writes one queue section, capped so the page stays bounded.

#### IMP-exporter-renderCatalog-ada131e: renderCatalog
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/project.go
x-reqmd.source-line: 385
x-reqmd.symbol-kind: method
trace: []
```

renderCatalog writes the document catalog table.

### IMP-exporter-NewProjectIndex-0ea572f: NewProjectIndex
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/project.go
x-reqmd.source-line: 40
x-reqmd.symbol-kind: function
trace: []
```

NewProjectIndex builds the project home model. hrefFor maps a document
directory path to the exported page href.

### IMP-exporter-queueItem-d6f06de: queueItem
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/project.go
x-reqmd.source-line: 197
x-reqmd.symbol-kind: type
trace: []
```

queueItem is one actionable finding on the project page.

### IMP-exporter-filepathBase-39a8116: filepathBase
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/project.go
x-reqmd.source-line: 242
x-reqmd.symbol-kind: function
trace: []
```

filepathBase returns the last path element, tolerating URL separators.

### IMP-exporter-queueSpec-2b56656: queueSpec
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/project.go
x-reqmd.source-line: 262
x-reqmd.symbol-kind: type
trace: []
```

queueSpec is one actionable queue: a title, an explanation, and the matcher
that decides which requirement states belong to it.

### IMP-exporter-projectQueueSpecs-bd0367a: projectQueueSpecs
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/project.go
x-reqmd.source-line: 272
x-reqmd.symbol-kind: var
trace: []
```

projectQueueSpecs is the fixed set of exception queues on the project home.

### IMP-exporter-missingCoverageNote-db9d580: missingCoverageNote
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/project.go
x-reqmd.source-line: 307
x-reqmd.symbol-kind: function
trace: []
```

missingCoverageNote matches requirements whose coverage expectation is not
met by an approved downstream requirement.

### IMP-exporter-unresolvedCoverageNote-d270f27: unresolvedCoverageNote
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/project.go
x-reqmd.source-line: 320
x-reqmd.symbol-kind: function
trace: []
```

unresolvedCoverageNote matches requirements expecting coverage from an
unknown document-id or level.

### IMP-exporter-missingVerdictNote-8a68a5f: missingVerdictNote
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/project.go
x-reqmd.source-line: 328
x-reqmd.symbol-kind: function
trace: []
```

missingVerdictNote matches eligible measures with no result.

### IMP-exporter-failingVerdictNote-c6db047: failingVerdictNote
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/project.go
x-reqmd.source-line: 336
x-reqmd.symbol-kind: function
trace: []
```

failingVerdictNote matches measures whose rolled-up verdict is fail.

### IMP-exporter-brokenTraceNote-bbf7d9c: brokenTraceNote
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/project.go
x-reqmd.source-line: 345
x-reqmd.symbol-kind: function
trace: []
```

brokenTraceNote matches requirements with unresolved, outdated, or predated
outbound trace links.

### IMP-exporter-attentionClass-2201240: attentionClass
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/project.go
x-reqmd.source-line: 428
x-reqmd.symbol-kind: function
trace: []
```

attentionClass returns a modifier class when a count should stand out.

### IMP-exporter-projectIndexCapacity-973368d: projectIndexCapacity
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/project.go
x-reqmd.source-line: 436
x-reqmd.symbol-kind: const
trace: []
```

projectIndexCapacity pre-sizes the cross-document search index.

### IMP-exporter-renderProjectJSON-4c8539d: renderProjectJSON
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/project.go
x-reqmd.source-line: 440
x-reqmd.symbol-kind: function
trace: []
```

renderProjectJSON writes the cross-document requirement index used by the
project search.

### IMP-exporter-entry-4a2c802: entry
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/project.go
x-reqmd.source-line: 441
x-reqmd.symbol-kind: type
trace: []
```

## internal/exporter/render.go
### IMP-exporter-normalizeAttrValue-cf3603f: normalizeAttrValue
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 28
x-reqmd.symbol-kind: function
trace: []
```

The HTML export has one server-rendered reading view and three analytical
views rendered on demand from the embedded page model:

	document — quiet card-per-requirement reading view (works without JS)
	table    — dense sortable grid with a column chooser and row expansion
	trace    — selected-requirement trace explorer with a bounded matrix
	coverage — actionable coverage, verification, and link-health lists

Every view, count, aggregate, and outline reads the same filtered
requirement set, so the toolbar means one thing.
normalizeAttrValue collapses []any / []string / scalar attribute
values into a uniform []string, dropping nil entries and nested maps.

### IMP-exporter-buildFilters-67a8eed: buildFilters
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 66
x-reqmd.symbol-kind: function
trace: []
```

buildFilters collects per-attribute unique values across all requirements
into a filterSet. Attributes with fewer than two distinct values are omitted
(single-value filters are not useful).

### IMP-exporter-collectFilterValues-5487f74: collectFilterValues
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 81
x-reqmd.symbol-kind: function
trace: []
```

collectFilterValues gathers the distinct values of every filterable
attribute. The trace attribute never becomes a filter; status values are
skipped when the document opts out of the status lifecycle.

### IMP-exporter-orderFilterValues-0a26312: orderFilterValues
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 106
x-reqmd.symbol-kind: function
trace: []
```

orderFilterValues returns one attribute's values in display order: status is
canonical (draft, approved, declared additions, then A→Z); everything else
is sorted A→Z.

### IMP-exporter-sortedStatusValues-bd62ea3: sortedStatusValues
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 121
x-reqmd.symbol-kind: function
trace: []
```

sortedStatusValues returns the given status values in canonical order:
built-ins [draft, approved] first, then declared additions A→Z, then any
unrecognized values A→Z.

### IMP-exporter-orderStatusFilterValues-449bd14: orderStatusFilterValues
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 153
x-reqmd.symbol-kind: function
trace: []
```

orderStatusFilterValues pulls the declared additions from doc.XReqmd
(nil-safe) and returns the values in canonical order.

### IMP-exporter-statusBreakdownItem-a3f0268: statusBreakdownItem
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 162
x-reqmd.symbol-kind: type
trace: []
```

statusBreakdownItem describes one status value and its count.

### IMP-exporter-countStatusBreakdown-fe5005e: countStatusBreakdown
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 170
x-reqmd.symbol-kind: function
trace: []
```

countStatusBreakdown tallies status values across all requirements in a
document, in display order: built-in (draft, approved) first, then
additions A→Z. Values not declared are treated as StatusDefault.

### IMP-exporter-statusColor-1a3f7db: statusColor
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 195
x-reqmd.symbol-kind: function
trace: []
```

statusColor returns the CSS color expression for a status value. Built-ins
return a CSS variable so the active theme selects the right value.

### IMP-exporter-effectiveStatus-bbda605: effectiveStatus
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 223
x-reqmd.symbol-kind: function
trace: []
```

effectiveStatus returns the requirement's status, defaulting to approved.

### IMP-exporter-countItems-fd559eb: countItems
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 231
x-reqmd.symbol-kind: function
trace: []
```

countItems tallies non-requirement nodes (containers + info items).

### IMP-exporter-maxHeadingLevel-cb1ef80: maxHeadingLevel
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 243
x-reqmd.symbol-kind: const
trace: []
```

Content headings start at h2 (h1 is the document title) and stop at h6.

### IMP-exporter-minHeadingLevel-55c9a36: minHeadingLevel
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 243
x-reqmd.symbol-kind: const
trace: []
```

Content headings start at h2 (h1 is the document title) and stop at h6.

### IMP-exporter-itemHeadingTag-bed1fcf: itemHeadingTag
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 251
x-reqmd.symbol-kind: function
trace: []
```

itemHeadingTag returns the HTML heading tag for an item's native markdown
level. Top-level items start at h2 (h1 is the document title); nested items
keep their native depth, floored one level below their parent.

### IMP-exporter-linkHref-8d73c6f: linkHref
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 266
x-reqmd.symbol-kind: function
trace: []
```

linkHref resolves a requirement ID to an href for this page: the
cross-document resolver when available, else a same-page fragment.

### IMP-exporter-renderDocHead-c2c6172: renderDocHead
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 284
x-reqmd.symbol-kind: function
trace: []
```

renderDocHead writes the <!DOCTYPE> + <html> + <head> block. The
stylesheet and client script are inlined: there is no runtime CDN request,
so an exported page is fully offline. The theme is applied by a small head
script before first paint to avoid a light/dark flash. The one exception is
a diagram-bearing body, which the Mermaid extender annotates with its own
pinned runtime script (see SW-EXP-004).

### IMP-exporter-brandMark-a3bdb12: brandMark
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 302
x-reqmd.symbol-kind: const
trace: []
```

brandMark is the same inline mark the documentation site uses: four blocks
joined by traces. currentColor lets it inherit the export accent in both
themes, so no second brand colour is introduced.

### IMP-exporter-renderBodyOpen-5d442da: renderBodyOpen
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 307
x-reqmd.symbol-kind: function
trace: []
```

renderBodyOpen writes the <body>, the app bar, the shell open, and the
server-rendered outline. The view switcher lives in the app bar so the
sticky toolbar is exactly one search-and-filter row.

### IMP-exporter-renderTOC-5664c60: renderTOC
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 334
x-reqmd.symbol-kind: function
trace: []
```

renderTOC writes the server-rendered outline. It works without JavaScript;
the client hides items that fall outside the filtered requirement set.

### IMP-exporter-renderDocHeader-f617d05: renderDocHeader
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 380
x-reqmd.symbol-kind: function
trace: []
```

renderDocHeader writes the document eyebrow, title, count line, status
distribution, boundary badges, and the frontmatter description. The trace
chain lives at the top of the reading view, below the toolbar.

### IMP-exporter-renderStatusBar-47f95f5: renderStatusBar
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 433
x-reqmd.symbol-kind: function
trace: []
```

renderStatusBar writes the proportional status distribution and its legend,
so the mix is readable at a glance without a chart library.

### IMP-exporter-docEyebrow-9600d47: docEyebrow
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 454
x-reqmd.symbol-kind: function
trace: []
```

docEyebrow identifies the document's position in the spec tree above the
title: the declared level and document-id when present.

### IMP-exporter-renderDocChain-0bf4359: renderDocChain
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 469
x-reqmd.symbol-kind: function
trace: []
```

renderDocChain writes the optional document chain navigation block.

### IMP-exporter-renderToolbar-c1616bd: renderToolbar
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 504
x-reqmd.symbol-kind: function
trace: []
```

renderToolbar writes the search field, the live count, the active-filter
chips placeholder, and the per-attribute filter menus. The view switcher
lives in the app bar, so this is a single sticky row.

### IMP-exporter-renderDocumentView-5d5c146: renderDocumentView
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 556
x-reqmd.symbol-kind: function
trace:
  - SW-EXP-002
```

reqmd:trace SW-EXP-002
renderDocumentView writes the server-rendered reading view: one card per
requirement, containers as collapsible sections, and info items as plain
blocks. Contents are visible by default (no x-cloak), so the reading view
survives a JavaScript failure or an offline file:// open.

### IMP-exporter-renderEmptyState-6c308e4: renderEmptyState
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 591
x-reqmd.symbol-kind: function
trace: []
```

renderEmptyState writes a consistent empty-state block.

### IMP-exporter-renderSectionOpen-fda944d: renderSectionOpen
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 601
x-reqmd.symbol-kind: function
trace: []
```

renderSectionOpen writes a non-requirement node and, when it is a titled
section with children, opens the collapsible folder that must wrap those
children. The caller renders the children and then calls renderSectionClose;
a section without children is a plain prose block, heading optional — an
untitled section is the document preamble.

### IMP-exporter-renderSectionClose-0f0772f: renderSectionClose
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 624
x-reqmd.symbol-kind: function
trace: []
```

renderSectionClose closes what renderSectionOpen started.

### IMP-exporter-sectionIsFolder-da1fa6d: sectionIsFolder
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 634
x-reqmd.symbol-kind: function
trace: []
```

sectionIsFolder reports whether a section is rendered as a collapsible
folder. An untitled section has no summary to click, so its children render
directly instead of inside a details element nobody can open.

### IMP-exporter-renderCard-4292b6c: renderCard
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 641
x-reqmd.symbol-kind: function
trace: []
```

renderCard writes one requirement card. Semantic state is carried on data
attributes (used by CSS and the client) plus at most a few state chips, so
the reading view stays quiet.

### IMP-exporter-detailMeta-b4d21c6: detailMeta
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 694
x-reqmd.symbol-kind: function
trace: []
```

detailMeta summarises what the collapsed detail block holds, so a reader
can tell whether opening it is worth the click.

### IMP-exporter-plural-66ff5c4: plural
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 711
x-reqmd.symbol-kind: function
trace: []
```

### IMP-exporter-renderCardHeader-8577c39: renderCardHeader
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 719
x-reqmd.symbol-kind: function
trace: []
```

renderCardHeader writes the ID/title heading and the anchor link.

### IMP-exporter-renderCardFlags-979ca3f: renderCardFlags
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 736
x-reqmd.symbol-kind: function
trace: []
```

renderCardFlags writes the state chips. Only exceptional or decision-
relevant states are shown: draft/custom status, coverage problems, and any
verdict. A covered requirement with an approved status adds no chrome.

### IMP-exporter-statusChip-b95da2e: statusChip
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 759
x-reqmd.symbol-kind: function
trace: []
```

statusChip renders a chip only for non-approved statuses: an approved
requirement adds no chrome.

### IMP-exporter-coverageChip-f82b232: coverageChip
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 770
x-reqmd.symbol-kind: function
trace: []
```

coverageChip renders coverage problems only (missing or unresolved); a
covered or not-required requirement adds no chrome.

### IMP-exporter-verdictChip-0cf63a9: verdictChip
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 797
x-reqmd.symbol-kind: function
trace: []
```

verdictChip renders the rolled-up verdict, or an explicit no-verdict chip
for an eligible measure when results were loaded.

### IMP-exporter-attrRow-18a99f7: attrRow
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 813
x-reqmd.symbol-kind: type
trace: []
```

attrRow is one rendered attribute cell in a requirement card.

### IMP-exporter-attrRows-e82db4c: attrRows
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 822
x-reqmd.symbol-kind: function
trace: []
```

attrRows returns the attribute cells to render, in display order: the
schema property order first, then any remaining attributes actually present
on the requirement sorted A→Z. The trace attribute is skipped.

### IMP-exporter-renderAttrGrid-23c9b19: renderAttrGrid
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 851
x-reqmd.symbol-kind: function
trace: []
```

renderAttrGrid writes the attribute definition list.

### IMP-exporter-renderRationale-e48ec03: renderRationale
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 868
x-reqmd.symbol-kind: function
trace: []
```

renderRationale writes the rationale block when present.

### IMP-exporter-renderEvidenceList-38ef797: renderEvidenceList
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 876
x-reqmd.symbol-kind: function
trace: []
```

renderEvidenceList writes the rolled-up verification evidence.

### IMP-exporter-renderTraceLinks-c0549dc: renderTraceLinks
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 902
x-reqmd.symbol-kind: function
trace: []
```

renderTraceLinks writes the upstream/downstream trace columns with the
authored reference, title, pin, and resolution state.

### IMP-exporter-renderTraceColumn-7b51849: renderTraceColumn
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 913
x-reqmd.symbol-kind: function
trace: []
```

renderTraceColumn writes one direction column.

### IMP-exporter-renderTraceLink-7ef88c7: renderTraceLink
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 925
x-reqmd.symbol-kind: function
trace: []
```

renderTraceLink writes one link with its pin/version state.

### IMP-exporter-renderPalette-8a06c99: renderPalette
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 955
x-reqmd.symbol-kind: function
trace: []
```

renderPalette writes the command-palette overlay. Overlays are the only
content hidden until the client boots.

### IMP-exporter-renderShortcuts-c47e1b1: renderShortcuts
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 971
x-reqmd.symbol-kind: function
trace: []
```

renderShortcuts writes the keyboard-shortcut dialog.

### IMP-exporter-renderBodyClose-a5a946e: renderBodyClose
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render.go
x-reqmd.source-line: 999
x-reqmd.symbol-kind: function
trace: []
```

renderBodyClose writes the end of the shell and document.

## internal/exporter/render_test.go
### IMP-exporter-testDoc-f9ee094: testDoc
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render_test.go
x-reqmd.source-line: 14
x-reqmd.symbol-kind: function
trace: []
```

testDoc builds a document with a container, requirement children, and an
info item so the render data covers every node kind.

### IMP-exporter-TestBuildRenderData_SlugAnchors-58f3692: TestBuildRenderData_SlugAnchors
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render_test.go
x-reqmd.source-line: 53
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-buildRenderData-e6e8856
```

TestBuildRenderData_SlugAnchors covers the section anchor rules: a slug from
the title, a deterministic suffix on collisions, and a fixed anchor for the
untitled preamble.

### IMP-exporter-TestBuildRenderData_SlugNeverShadowsRequirementID-431feb7: TestBuildRenderData_SlugNeverShadowsRequirementID
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render_test.go
x-reqmd.source-line: 85
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-buildRenderData-e6e8856
```

TestBuildRenderData_SlugNeverShadowsRequirementID guards the collision rule:
a requirement whose ID happens to look like a section slug keeps its own
anchor, and the section takes the next free suffix.

### IMP-exporter-TestBuildRenderData-2b68936: TestBuildRenderData
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render_test.go
x-reqmd.source-line: 99
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-buildRenderData-e6e8856
```

### IMP-exporter-TestBuildRenderDataIgnoreStatus-d03e1fa: TestBuildRenderDataIgnoreStatus
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render_test.go
x-reqmd.source-line: 145
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-buildRenderData-e6e8856
```

### IMP-exporter-TestBuildFiltersOmitsSingleValueAttrs-16c43f2: TestBuildFiltersOmitsSingleValueAttrs
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render_test.go
x-reqmd.source-line: 157
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-buildFilters-67a8eed
```

### IMP-exporter-TestBuildFiltersStatusOrdering-cdfe9f7: TestBuildFiltersStatusOrdering
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render_test.go
x-reqmd.source-line: 168
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-buildFilters-67a8eed
```

### IMP-exporter-TestNormalizeAttrValue-232c71f: TestNormalizeAttrValue
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render_test.go
x-reqmd.source-line: 183
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-normalizeAttrValue-cf3603f
```

### IMP-exporter-TestSortedStatusValues-363fa83: TestSortedStatusValues
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render_test.go
x-reqmd.source-line: 208
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-sortedStatusValues-bd62ea3
```

### IMP-exporter-TestItemHeadingTagNativeLevels-6d0fc83: TestItemHeadingTagNativeLevels
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render_test.go
x-reqmd.source-line: 219
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-itemHeadingTag-bed1fcf
```

### IMP-exporter-TestDocTitleFallback-43206a0: TestDocTitleFallback
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render_test.go
x-reqmd.source-line: 247
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-docTitle-4498ba2
```

### IMP-exporter-TestAttributeColumns-7fd3a4d: TestAttributeColumns
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render_test.go
x-reqmd.source-line: 259
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-attributeColumns-0320a2d
```

### IMP-exporter-TestBuildPageModel-d7d6eae: TestBuildPageModel
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render_test.go
x-reqmd.source-line: 269
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-buildPageModel-c285b4e
```

### IMP-exporter-TestBuildPageModelResolveLink-35121e6: TestBuildPageModelResolveLink
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render_test.go
x-reqmd.source-line: 314
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-buildPageModel-c285b4e
```

### IMP-exporter-TestDocTitleFallsBackToPathInExport-fc54ff0: TestDocTitleFallsBackToPathInExport
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/render_test.go
x-reqmd.source-line: 332
x-reqmd.symbol-kind: function
trace:
  - IMP-exporter-docTitle-4498ba2
```

## internal/exporter/verdict.go
### IMP-exporter-VerdictInfoFromGraph-4e7e699: VerdictInfoFromGraph
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/exporter/verdict.go
x-reqmd.source-line: 8
x-reqmd.symbol-kind: function
trace: []
```

VerdictInfoFromGraph converts a graph-side rolled-up NodeVerdict into the
exporter's render type, carrying the outcome, representative source file,
the contributing case keys (for CSV), and the evidence items (for HTML).

