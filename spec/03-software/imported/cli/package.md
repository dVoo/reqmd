# Package: cli

<!-- reqmd: generated -->

## internal/cli/baseline.go
### IMP-cli-newBaselineCmd-9148414: newBaselineCmd
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/baseline.go
x-reqmd.source-line: 18
x-reqmd.symbol-kind: function
trace: []
```

### IMP-cli-newBaselineDiffCmd-65bc3e9: newBaselineDiffCmd
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/baseline.go
x-reqmd.source-line: 27
x-reqmd.symbol-kind: function
trace: []
```

### IMP-cli-runBaselineDiff-79607c8: runBaselineDiff
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/baseline.go
x-reqmd.source-line: 74
x-reqmd.symbol-kind: function
trace: []
```

### IMP-cli-runBaselineDiffFilteredViews-a6edaf7: runBaselineDiffFilteredViews
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/baseline.go
x-reqmd.source-line: 154
x-reqmd.symbol-kind: function
trace: []
```

runBaselineDiffFilteredViews loads a single snapshot and compares two
filtered views of it (RFC §3.4 — "what does Premium add over Base" from
a single commit).

### IMP-cli-applyFilter-b409837: applyFilter
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/baseline.go
x-reqmd.source-line: 194
x-reqmd.symbol-kind: function
trace: []
```

applyFilter compiles and applies a filter expression to a doc slice,
returning the filtered docs. Shared by both baseline diff modes.

## internal/cli/baseline_test.go
### IMP-cli-TestBaselineDiff_FlagGroupValidation-57599da: TestBaselineDiff_FlagGroupValidation
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/baseline_test.go
x-reqmd.source-line: 12
x-reqmd.symbol-kind: function
trace: []
```

TestBaselineDiff_FlagGroupValidation verifies cobra flag-group rules:
--filter-a/--filter-b are required together (RFC §3.4) and --filter is
mutually exclusive with the two-view mode. These are validated before
RunE, so no git repo is required.

### IMP-cli-TestBaselineDiff_TwoTagFilter-333109c: TestBaselineDiff_TwoTagFilter
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/baseline_test.go
x-reqmd.source-line: 38
x-reqmd.symbol-kind: function
trace: []
```

TestBaselineDiff_TwoTagFilter compares two tags and verifies --filter
scopes both snapshots (RFC §3.3).

### IMP-cli-TestBaselineDiff_FilteredViews-f36004f: TestBaselineDiff_FilteredViews
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/baseline_test.go
x-reqmd.source-line: 82
x-reqmd.symbol-kind: function
trace: []
```

TestBaselineDiff_FilteredViews verifies the same-commit two-view mode
(RFC §3.4): "what Premium adds over Base" from a single commit.

## internal/cli/check.go
### IMP-cli-newCheckCmd-22608ab: newCheckCmd
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/check.go
x-reqmd.source-line: 24
x-reqmd.symbol-kind: function
trace: []
```

### IMP-cli-runValidationPipeline-6e88051: runValidationPipeline
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/check.go
x-reqmd.source-line: 100
x-reqmd.symbol-kind: function
trace: []
```

runValidationPipeline runs the common validation pipeline for a set of documents.
It builds doc headers, runs Pass 1 JSON Schema validation, and (when docs are from
a discovery walk) runs Pass 2 graph trace checks.

relaxedVersions, when true, demotes "outdated" version-pin findings from
ERROR to WARNING. Predated findings (pin ahead of upstream) stay ERROR
because they are a data integrity issue, not a process issue.

relaxedFilter, when true, lets the --filter expression reference
attributes that no schema.yaml declares. Attributes actually present on a
requirement are always allowed; anything else outside the declared set is
reported as a warning instead of aborting the run.

### IMP-cli-graphResult-18ba9d0: graphResult
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/check.go
x-reqmd.source-line: 172
x-reqmd.symbol-kind: type
trace: []
```

Pass 2: the trace graph only reads docs, so it is built concurrently
with Pass 1 below — its cost hides behind schema validation. The result
is consumed after Pass 1 so report entries keep their order (compile
errors before graph errors). root is non-empty only for dir mode;
single-file mode has no graph.

### IMP-cli-compileResult-866a429: compileResult
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/check.go
x-reqmd.source-line: 195
x-reqmd.symbol-kind: type
trace: []
```

Pass 1: JSON Schema validation (spec docs only; result doc has no schema)
Schema compilation and per-requirement validation are both parallelized.
Compile is per-doc-dir (independent schemas), validate is per-requirement
(the compiled schema is read-only after construction).
Phase A: compile all schemas in parallel.

### IMP-cli-valResult-35dc3fb: valResult
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/check.go
x-reqmd.source-line: 232
x-reqmd.symbol-kind: type
trace: []
```

Phase B: validate all requirements in parallel across all docs
that compiled successfully. Each Validate call is independent —
the Compiled struct is read-only and each req's attrs map is separate.

### IMP-cli-valJob-fd6fb35: valJob
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/check.go
x-reqmd.source-line: 239
x-reqmd.symbol-kind: type
trace: []
```

Flatten the work list: (jobIdx, docIdx, reqIdx) triples for docs
with valid schemas.

### IMP-cli-demoteOutdatedVersionPins-1f86521: demoteOutdatedVersionPins
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/check.go
x-reqmd.source-line: 343
x-reqmd.symbol-kind: function
trace: []
```

demoteOutdatedVersionPins demotes Code=="version-pin" + Direction=="outdated"
findings from ERROR to WARNING. Predated findings (Direction=="predated")
are left at ERROR — they indicate a downstream pinning a version that
does not exist on the upstream and are a data integrity issue.

### IMP-cli-validateDir-140fe1c: validateDir
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/check.go
x-reqmd.source-line: 352
x-reqmd.symbol-kind: function
trace: []
```

### IMP-cli-validateSingleFile-de6bca7: validateSingleFile
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/check.go
x-reqmd.source-line: 360
x-reqmd.symbol-kind: function
trace: []
```

### IMP-cli-filterID-fd064a8: filterID
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/check.go
x-reqmd.source-line: 387
x-reqmd.symbol-kind: function
trace: []
```

filterID reports whether reqID passes the active filter set. A nil set
means no filter is active — every requirement passes.

### IMP-cli-formatReport-022f80e: formatReport
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/check.go
x-reqmd.source-line: 396
x-reqmd.symbol-kind: function
trace: []
```

formatReport conditionally returns the report in text or JSON format.

## internal/cli/check_test.go
### IMP-cli-jsonCheckSummary-95c047d: jsonCheckSummary
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/check_test.go
x-reqmd.source-line: 14
x-reqmd.symbol-kind: type
trace: []
```

jsonCheckSummary captures the parts of `check --json` output that the
filter/disjoint tests assert on.

### IMP-cli-TestCheckCmd_FilterScopesValidation-01cdba8: TestCheckCmd_FilterScopesValidation
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/check_test.go
x-reqmd.source-line: 24
x-reqmd.symbol-kind: function
trace: []
```

### IMP-cli-TestCheckCmd_FilterUnknownAttributeFails-fb8a60d: TestCheckCmd_FilterUnknownAttributeFails
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/check_test.go
x-reqmd.source-line: 64
x-reqmd.symbol-kind: function
trace: []
```

### IMP-cli-TestCheckCmd_FilterSyntaxErrorFails-2976cfe: TestCheckCmd_FilterSyntaxErrorFails
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/check_test.go
x-reqmd.source-line: 75
x-reqmd.symbol-kind: function
trace: []
```

### IMP-cli-TestCheckCmd_FilterBuiltinTraceVar-b98cd83: TestCheckCmd_FilterBuiltinTraceVar
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/check_test.go
x-reqmd.source-line: 86
x-reqmd.symbol-kind: function
trace: []
```

TestCheckCmd_FilterBuiltinTraceVar verifies the RFC §2.3 built-in
attribute `trace` is accepted by the typo check even though it is not
declared in the fixture schema.

### IMP-cli-TestCheckCmd_DisjointCheckFlag-e752672: TestCheckCmd_DisjointCheckFlag
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/check_test.go
x-reqmd.source-line: 101
x-reqmd.symbol-kind: function
trace: []
```

### IMP-cli-TestCheckCmd_DisjointCheckClean-b020839: TestCheckCmd_DisjointCheckClean
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/check_test.go
x-reqmd.source-line: 116
x-reqmd.symbol-kind: function
trace: []
```

TestCheckCmd_DisjointCheckClean verifies the disjoint check is off by
default and passes when variants overlap.

### IMP-cli-TestCheckCmd_DisjointCheckSchemaDecl-3074657: TestCheckCmd_DisjointCheckSchemaDecl
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/check_test.go
x-reqmd.source-line: 130
x-reqmd.symbol-kind: function
trace: []
```

TestCheckCmd_DisjointCheckSchemaDecl verifies the check is enabled via
x-reqmd.disjoint-check in schema.yaml (RFC §3.5).

### IMP-cli-TestCheckCmd_FilterWithResults-44c63c4: TestCheckCmd_FilterWithResults
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/check_test.go
x-reqmd.source-line: 151
x-reqmd.symbol-kind: function
trace: []
```

TestCheckCmd_FilterWithResults verifies the --filter + --results
interplay: a measure in the filter set reports its failing verdict even
though the synthesized RESULT node is not itself in the filter set; a
measure excluded by the filter is not reported.

### IMP-cli-TestCheckCmd_RequiresTraceFromDocumentDefault-6e84959: TestCheckCmd_RequiresTraceFromDocumentDefault
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/check_test.go
x-reqmd.source-line: 182
x-reqmd.symbol-kind: function
trace: []
```

TestCheckCmd_RequiresTraceFromDocumentDefault verifies the
x-reqmd.requires-trace-from document default end-to-end through schema.yaml:
a requirement without its own attribute inherits the default.

### IMP-cli-writeUnboundResultsDir-2cc8702: writeUnboundResultsDir
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/check_test.go
x-reqmd.source-line: 220
x-reqmd.symbol-kind: function
trace: []
```

writeUnboundResultsDir creates a results dir containing one CTRF report
whose test declares x-reqmd but binds to nothing (case only). Returns the
results dir path.

### IMP-cli-TestCheckCmd_IgnoreUnboundResults-7e63da7: TestCheckCmd_IgnoreUnboundResults
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/check_test.go
x-reqmd.source-line: 238
x-reqmd.symbol-kind: function
trace: []
```

TestCheckCmd_IgnoreUnboundResults verifies unbound-result warnings surface
by default and are suppressed by --ignore-unbound-results.

### IMP-cli-TestCheckCmd_RelaxedFilterAllowsUndeclared-b2dc1a6: TestCheckCmd_RelaxedFilterAllowsUndeclared
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/check_test.go
x-reqmd.source-line: 267
x-reqmd.symbol-kind: function
trace: []
```

TestCheckCmd_RelaxedFilterAllowsUndeclared pins the --relaxed-filter
contract: an attribute present on requirements but declared in no
schema.yaml is rejected by the default typo check and accepted — with a
warning instead of an error — when the flag is set.

## internal/cli/e2e_test.go
### IMP-cli-bundleFixture-721d374: bundleFixture
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/e2e_test.go
x-reqmd.source-line: 18
x-reqmd.symbol-kind: const
trace: []
```

bundleFixture is a real exchange bundle: sections grouping requirements, an
internal link, an external ref, a declared schema and x-reqmd metadata.

### IMP-cli-renderBundle-16e747f: renderBundle
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/e2e_test.go
x-reqmd.source-line: 22
x-reqmd.symbol-kind: function
trace: []
```

renderBundle runs a bundle through the renderer and writes the files it
produces into dir, exactly as `reqmd-import` does.

### IMP-cli-TestE2E_ExchangeBundleToCheckAndExports-f638a3c: TestE2E_ExchangeBundleToCheckAndExports
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/e2e_test.go
x-reqmd.source-line: 46
x-reqmd.symbol-kind: function
trace: []
```

TestE2E_ExchangeBundleToCheckAndExports is the whole chain in one test: an
exchange bundle becomes reqmd documents, those documents validate, and the
exports carry the structure the bundle declared. It asserts behaviour across
module boundaries (exchange → parser → graph → exporter) instead of inside a
single package, which is where the pipeline actually breaks.

### IMP-cli-TestE2E_SectionsSurviveParseAndFilter-ad82110: TestE2E_SectionsSurviveParseAndFilter
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/e2e_test.go
x-reqmd.source-line: 140
x-reqmd.symbol-kind: function
trace: []
```

TestE2E_SectionsSurviveParseAndFilter covers the section model end to end in
reqmd core: a real document is parsed into a tree, and a filtered view drops
a folder whose requirements were all pruned while keeping leaf sections.

### IMP-cli-TestE2E_CSVImporterToCheckChain-e38dc5a: TestE2E_CSVImporterToCheckChain
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/e2e_test.go
x-reqmd.source-line: 242
x-reqmd.symbol-kind: function
trace: []
```

TestE2E_CSVImporterToCheckChain runs the documented flagship flow when a
Python interpreter is available: DOORS CSV → reqmd_io bundle → renderer →
check. It is skipped rather than failed when python3 is missing, the same way
the client-side JS tests skip without node.

### IMP-cli-TestE2E_ExportJSONTables-000a151: TestE2E_ExportJSONTables
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/e2e_test.go
x-reqmd.source-line: 296
x-reqmd.symbol-kind: function
trace: []
```

TestE2E_ExportJSONTables pins what a plain `export json` guarantees at the
CLI boundary: the four tables always exist, results stay empty without
--results, authored text survives the JSON encoding, and a second export is
byte-identical.

### IMP-cli-TestE2E_ExportJSONFilter-3b5f786: TestE2E_ExportJSONFilter
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/e2e_test.go
x-reqmd.source-line: 362
x-reqmd.symbol-kind: function
trace: []
```

TestE2E_ExportJSONFilter covers --filter at the CLI boundary: it scopes the
requirements table and, because the command hands the filtered docs to the
trace writer, traces.ndjson follows it too. Findings still cover the whole
graph, so that table is written regardless.

### IMP-cli-TestE2E_ExportJSONResults-de998aa: TestE2E_ExportJSONResults
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/e2e_test.go
x-reqmd.source-line: 412
x-reqmd.symbol-kind: function
trace: []
```

TestE2E_ExportJSONResults covers --results: a manual result bound to an
authored measure becomes a verdict row while the requirement and trace
tables keep the authored tree only.

### IMP-cli-TestE2E_ExportJSONDisjointFinding-e09b821: TestE2E_ExportJSONDisjointFinding
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/e2e_test.go
x-reqmd.source-line: 451
x-reqmd.symbol-kind: function
trace: []
```

TestE2E_ExportJSONDisjointFinding covers --disjoint-check: a trace pair with
no overlapping attribute value surfaces as an ERROR in findings.ndjson.

### IMP-cli-ndjsonRows-818dc52: ndjsonRows
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/e2e_test.go
x-reqmd.source-line: 479
x-reqmd.symbol-kind: function
trace: []
```

ndjsonRows reads a JSON-lines table.

### IMP-cli-keysOf-6e4a82e: keysOf
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/e2e_test.go
x-reqmd.source-line: 500
x-reqmd.symbol-kind: function
trace: []
```

keysOf lists map keys for failure messages.

### IMP-cli-TestE2E_ExportHTMLConcurrentPagesDeterministic-ba5414c: TestE2E_ExportHTMLConcurrentPagesDeterministic
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/e2e_test.go
x-reqmd.source-line: 511
x-reqmd.symbol-kind: function
trace: []
```

TestE2E_ExportHTMLConcurrentPagesDeterministic pins the concurrent page
render: a multi-document export runs on a worker pool, so the bytes must not
depend on which worker finished first.

## internal/cli/export_test.go
### IMP-cli-TestExportCSV_Filter-7a4ad02: TestExportCSV_Filter
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/export_test.go
x-reqmd.source-line: 10
x-reqmd.symbol-kind: function
trace: []
```

### IMP-cli-TestExportHTML_Filter-1df7a3f: TestExportHTML_Filter
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/export_test.go
x-reqmd.source-line: 31
x-reqmd.symbol-kind: function
trace: []
```

### IMP-cli-TestExportHTML_ProjectIndex-83dbdfa: TestExportHTML_ProjectIndex
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/export_test.go
x-reqmd.source-line: 54
x-reqmd.symbol-kind: function
trace: []
```

TestExportHTML_ProjectIndex verifies that an export writes the project home
alongside the document pages.

### IMP-cli-TestExportHTML_NestedTraceLinks-484a380: TestExportHTML_NestedTraceLinks
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/export_test.go
x-reqmd.source-line: 75
x-reqmd.symbol-kind: function
trace: []
```

TestExportHTML_NestedTraceLinks verifies that, without -o, cross-document
trace links are relative to each document's own directory.

## internal/cli/filters.go
### IMP-cli-relaxedFilterUsage-25ec00b: relaxedFilterUsage
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/filters.go
x-reqmd.source-line: 7
x-reqmd.symbol-kind: const
trace: []
```

relaxedFilterUsage is the shared help text for --relaxed-filter, kept in
one place so every command that accepts --filter documents it identically.

### IMP-cli-addRelaxedFilterFlag-119dae1: addRelaxedFilterFlag
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/filters.go
x-reqmd.source-line: 11
x-reqmd.symbol-kind: function
trace: []
```

addRelaxedFilterFlag registers --relaxed-filter on cmd, storing the value in
target. Commands pair it with their --filter flag; see filter.CompileForDocs.

## internal/cli/init.go
### IMP-cli-genericSchemaTmpl-1ab7498: genericSchemaTmpl
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/init.go
x-reqmd.source-line: 15
x-reqmd.symbol-kind: var
trace: []
```

go:embed init_templates/generic_schema.yaml

### IMP-cli-genericExampleTmpl-33c3cb1: genericExampleTmpl
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/init.go
x-reqmd.source-line: 18
x-reqmd.symbol-kind: var
trace: []
```

go:embed init_templates/generic_example.md

### IMP-cli-aspiceSchemaTmpl-663a33c: aspiceSchemaTmpl
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/init.go
x-reqmd.source-line: 21
x-reqmd.symbol-kind: var
trace: []
```

go:embed init_templates/aspice_schema.yaml

### IMP-cli-aspiceExampleTmpl-5c7966d: aspiceExampleTmpl
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/init.go
x-reqmd.source-line: 24
x-reqmd.symbol-kind: var
trace: []
```

go:embed init_templates/aspice_example.md

### IMP-cli-resultsSchemaTmpl-5f3dd93: resultsSchemaTmpl
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/init.go
x-reqmd.source-line: 27
x-reqmd.symbol-kind: var
trace: []
```

go:embed init_templates/results_schema.yaml

### IMP-cli-resultsExampleTmpl-143d494: resultsExampleTmpl
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/init.go
x-reqmd.source-line: 30
x-reqmd.symbol-kind: var
trace: []
```

go:embed init_templates/results_example.md

### IMP-cli-initConfig-35500bf: initConfig
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/init.go
x-reqmd.source-line: 32
x-reqmd.symbol-kind: type
trace: []
```

### IMP-cli-newInitCmd-eea24a3: newInitCmd
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/init.go
x-reqmd.source-line: 42
x-reqmd.symbol-kind: function
trace: []
```

### IMP-cli-runInit-6cefd16: runInit
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/init.go
x-reqmd.source-line: 79
x-reqmd.symbol-kind: function
trace: []
```

### IMP-cli-isDirPath-b958bad: isDirPath
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/init.go
x-reqmd.source-line: 183
x-reqmd.symbol-kind: function
trace: []
```

isDirPath reports whether the preset string looks like a filesystem path
(contains a path separator or ends in a known extension), as opposed to a
named preset like "generic", "aspice", or "results".

### IMP-cli-loadCustomPreset-d21d1af: loadCustomPreset
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/init.go
x-reqmd.source-line: 200
x-reqmd.symbol-kind: function
trace: []
```

loadCustomPreset reads schema and example templates from a directory.
It looks for files in this order of preference:
   - schema.yaml.tmpl + example.md.tmpl  (Go template files)
   - schema.yaml       + example.md      (plain files, used verbatim)

The example filename (without .tmpl) is used as the output filename.
If multiple .md files exist, the first one found is used as the example.

### IMP-cli-writeTemplate-7bb9c58: writeTemplate
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/init.go
x-reqmd.source-line: 247
x-reqmd.symbol-kind: function
trace: []
```

## internal/cli/init_test.go
### IMP-cli-TestInit_GenericPreset-0811774: TestInit_GenericPreset
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/init_test.go
x-reqmd.source-line: 12
x-reqmd.symbol-kind: function
trace: []
```

TestInit_GenericPreset checks the default preset renders a schema and an
example document with the requested prefix.

### IMP-cli-TestInit_ResultsPreset-644594f: TestInit_ResultsPreset
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/init_test.go
x-reqmd.source-line: 35
x-reqmd.symbol-kind: function
trace: []
```

TestInit_ResultsPreset checks the results preset overrides level and prefix
so the scaffolded directory is a valid manual-results document.

### IMP-cli-TestInit_CustomPresetDir-75d28d9: TestInit_CustomPresetDir
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/init_test.go
x-reqmd.source-line: 53
x-reqmd.symbol-kind: function
trace: []
```

TestInit_CustomPresetDir checks a preset directory renders both templates
with the schema variables.

### IMP-cli-TestInit_ExistingFilesNeedForce-cd66359: TestInit_ExistingFilesNeedForce
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/init_test.go
x-reqmd.source-line: 76
x-reqmd.symbol-kind: function
trace: []
```

TestInit_ExistingFilesNeedForce checks an existing document aborts the run
unless --force is given.

### IMP-cli-readFileString-e4dc1d4: readFileString
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/init_test.go
x-reqmd.source-line: 91
x-reqmd.symbol-kind: function
trace: []
```

## internal/cli/list.go
### IMP-cli-newListCmd-d11e211: newListCmd
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/list.go
x-reqmd.source-line: 14
x-reqmd.symbol-kind: function
trace: []
```

## internal/cli/list_stats_test.go
### IMP-cli-TestListCmd_Filter-56e0080: TestListCmd_Filter
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/list_stats_test.go
x-reqmd.source-line: 8
x-reqmd.symbol-kind: function
trace: []
```

### IMP-cli-TestListCmd_FilterJSON-907c912: TestListCmd_FilterJSON
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/list_stats_test.go
x-reqmd.source-line: 31
x-reqmd.symbol-kind: function
trace: []
```

### IMP-cli-TestStatsCmd_Filter-d509761: TestStatsCmd_Filter
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/list_stats_test.go
x-reqmd.source-line: 46
x-reqmd.symbol-kind: function
trace: []
```

### IMP-cli-TestListCmd_ShowsItems-69d106c: TestListCmd_ShowsItems
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/list_stats_test.go
x-reqmd.source-line: 66
x-reqmd.symbol-kind: function
trace: []
```

### IMP-cli-TestListCmd_ItemsJSON-971a01c: TestListCmd_ItemsJSON
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/list_stats_test.go
x-reqmd.source-line: 104
x-reqmd.symbol-kind: function
trace: []
```

### IMP-cli-TestStatsCmd_ShowsItems-59e04ad: TestStatsCmd_ShowsItems
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/list_stats_test.go
x-reqmd.source-line: 121
x-reqmd.symbol-kind: function
trace: []
```

## internal/cli/repin.go
### IMP-cli-newRepinCmd-9e593b7: newRepinCmd
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/repin.go
x-reqmd.source-line: 17
x-reqmd.symbol-kind: function
trace: []
```

### IMP-cli-repinOptions-15f9721: repinOptions
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/repin.go
x-reqmd.source-line: 69
x-reqmd.symbol-kind: type
trace: []
```

### IMP-cli-runRepin-77ec084: runRepin
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/repin.go
x-reqmd.source-line: 75
x-reqmd.symbol-kind: function
trace: []
```

### IMP-cli-applyRepin-8795bdf: applyRepin
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/repin.go
x-reqmd.source-line: 132
x-reqmd.symbol-kind: function
trace: []
```

### IMP-cli-writeJSON-771a780: writeJSON
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/repin.go
x-reqmd.source-line: 159
x-reqmd.symbol-kind: function
trace: []
```

### IMP-cli-isTerminal-ca2655b: isTerminal
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/repin.go
x-reqmd.source-line: 168
x-reqmd.symbol-kind: function
trace: []
```

### IMP-cli-promptYes-935399c: promptYes
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/repin.go
x-reqmd.source-line: 170
x-reqmd.symbol-kind: function
trace: []
```

## internal/cli/root.go
### IMP-cli-NewRootCmd-c85094d: NewRootCmd
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/root.go
x-reqmd.source-line: 11
x-reqmd.symbol-kind: function
trace: []
```

NewRootCmd builds the root cobra command. version is injected at build time
via -ldflags "-X main.version=..."; when empty, --version is omitted.

## internal/cli/serve.go
### IMP-cli-pollInterval-2c853d7: pollInterval
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/serve.go
x-reqmd.source-line: 29
x-reqmd.symbol-kind: const
trace: []
```

pollInterval is the file polling interval used when fsnotify is unavailable.

### IMP-cli-newServeCmd-92943ea: newServeCmd
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/serve.go
x-reqmd.source-line: 31
x-reqmd.symbol-kind: function
trace: []
```

### IMP-cli-runServe-213a29f: runServe
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/serve.go
x-reqmd.source-line: 82
x-reqmd.symbol-kind: function
trace: []
```

### IMP-cli-serveData-5584b92: serveData
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/serve.go
x-reqmd.source-line: 167
x-reqmd.symbol-kind: type
trace: []
```

serveData holds shared state for the serve command.

#### IMP-cli-rebuild-231437b: rebuild
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/serve.go
x-reqmd.source-line: 179
x-reqmd.symbol-kind: method
trace: []
```

#### IMP-cli-watchLoop-d8934f5: watchLoop
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/serve.go
x-reqmd.source-line: 316
x-reqmd.symbol-kind: method
trace: []
```

#### IMP-cli-watchFsnotify-dd68ce1: watchFsnotify
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/serve.go
x-reqmd.source-line: 331
x-reqmd.symbol-kind: method
trace: []
```

#### IMP-cli-watchPolling-2c0006a: watchPolling
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/serve.go
x-reqmd.source-line: 420
x-reqmd.symbol-kind: method
trace: []
```

#### IMP-cli-makeRebuildAndNotify-a8e55a7: makeRebuildAndNotify
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/serve.go
x-reqmd.source-line: 512
x-reqmd.symbol-kind: method
trace: []
```

#### IMP-cli-sseHandler-ebdb35b: sseHandler
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/serve.go
x-reqmd.source-line: 587
x-reqmd.symbol-kind: method
trace: []
```

sseHandler implements Server-Sent Events for live-reload.

#### IMP-cli-fileHandler-ebc01a2: fileHandler
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/serve.go
x-reqmd.source-line: 628
x-reqmd.symbol-kind: method
trace: []
```

fileHandler serves the in-memory rendered HTML pages.

### IMP-cli-discoverAndBuildGraph-bd79c15: discoverAndBuildGraph
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/serve.go
x-reqmd.source-line: 232
x-reqmd.symbol-kind: function
trace: []
```

discoverAndBuildGraph runs Pass 0 document discovery, loads ephemeral
verification results (if any), synthesizes them into pseudo-requirements,
and builds the requirement graph from the combined document set.
Returns the original docs (without synthesized results), the verdicts
map for badge rendering, and the graph.

### IMP-cli-exportServePages-7761740: exportServePages
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/serve.go
x-reqmd.source-line: 258
x-reqmd.symbol-kind: function
trace: []
```

exportServePages renders every document to HTML and counts graph checks.
It returns the rendered pages, the total requirement count, and the graph
check results so the caller can report status.

### IMP-cli-publishPages-f2a6b3e: publishPages
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/serve.go
x-reqmd.source-line: 310
x-reqmd.symbol-kind: function
trace: []
```

publishPages atomically swaps the in-memory rendered pages.

### IMP-cli-errFsnotifyUnavailable-6a13a1a: errFsnotifyUnavailable
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/serve.go
x-reqmd.source-line: 329
x-reqmd.symbol-kind: var
trace: []
```

errFsnotifyUnavailable is returned when fsnotify cannot be used.

### IMP-cli-docState-e9b9e58: docState
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/serve.go
x-reqmd.source-line: 424
x-reqmd.symbol-kind: type
trace: []
```

Track last modification time per doc dir

### IMP-cli-openBrowser-c535b25: openBrowser
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/serve.go
x-reqmd.source-line: 532
x-reqmd.symbol-kind: function
trace: []
```

openBrowser opens the given URL in the default system browser.

### IMP-cli-findDocDirsFlat-854fd15: findDocDirsFlat
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/serve.go
x-reqmd.source-line: 553
x-reqmd.symbol-kind: function
trace: []
```

findDocDirsFlat finds all directories under root that contain a schema.yaml.
Entry-level walk errors are skipped: a watcher that sees the readable part
of the tree is more useful than none.

### IMP-cli-findResultWatchDirs-c52a3f8: findResultWatchDirs
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/serve.go
x-reqmd.source-line: 573
x-reqmd.symbol-kind: function
trace: []
```

findResultWatchDirs returns the directories that should be watched to
detect changes in the given results path. For a file path, the parent
directory is watched. For a directory path, the directory itself is
watched (for CTRF files) plus any subdirectories containing a schema.yaml
(manual-results dirs).

## internal/cli/serve_test.go
### IMP-cli-findResult-d74f389: findResult
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/serve_test.go
x-reqmd.source-line: 15
x-reqmd.symbol-kind: function
trace: []
```

findResult reports whether any check result for reqID mentions substr.

### IMP-cli-TestServeDiscover_WholeTree-62ae3f1: TestServeDiscover_WholeTree
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/serve_test.go
x-reqmd.source-line: 27
x-reqmd.symbol-kind: function
trace: []
```

TestServeDiscover_WholeTree verifies the serve pipeline builds the graph
and exports from the full tree: every requirement is served and coverage
resolves across documents.

### IMP-cli-TestServeProjectHome-f425efc: TestServeProjectHome
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/serve_test.go
x-reqmd.source-line: 54
x-reqmd.symbol-kind: function
trace: []
```

TestServeProjectHome verifies `/` serves the project home page instead of
redirecting to the first document.

## internal/cli/stats.go
### IMP-cli-newStatsCmd-a1a9c0e: newStatsCmd
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/stats.go
x-reqmd.source-line: 14
x-reqmd.symbol-kind: function
trace: []
```

## internal/cli/summaries.go
### IMP-cli-buildDocSummaries-91fc2d7: buildDocSummaries
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/summaries.go
x-reqmd.source-line: 11
x-reqmd.symbol-kind: function
trace: []
```

buildDocSummaries converts parsed documents into a slice of DocumentSummary
for use by list and stats commands. Rows contain every node — requirements
and information items (containers + info blocks) — in document order.

## internal/cli/testutil_test.go
### IMP-cli-writeFile-467aec0: writeFile
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/testutil_test.go
x-reqmd.source-line: 16
x-reqmd.symbol-kind: function
trace: []
```

writeFile writes content to rel under dir, creating parent directories.

### IMP-cli-runCmd-d07ae24: runCmd
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/testutil_test.go
x-reqmd.source-line: 28
x-reqmd.symbol-kind: function
trace: []
```

runCmd executes a cobra command with args, capturing stdout+stderr.

### IMP-cli-mdReq-02dc542: mdReq
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/testutil_test.go
x-reqmd.source-line: 39
x-reqmd.symbol-kind: function
trace: []
```

mdReq renders a requirement heading plus its ```attr YAML block.

### IMP-cli-variantSchema-bec263f: variantSchema
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/testutil_test.go
x-reqmd.source-line: 46
x-reqmd.symbol-kind: const
trace: []
```

variantSchema is a minimal schema declaring the custom `variant`
attribute. `status`, `trace`, and `requires-trace-from` are reqmd
built-ins injected automatically, so they need no declaration.

### IMP-cli-verifySchema-d624d19: verifySchema
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/testutil_test.go
x-reqmd.source-line: 60
x-reqmd.symbol-kind: const
trace: []
```

verifySchema adds the custom `verify` attribute used by measures in the
--filter + --results test.

### IMP-cli-writeVariantSpec-296fae0: writeVariantSpec
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/testutil_test.go
x-reqmd.source-line: 77
x-reqmd.symbol-kind: function
trace: []
```

writeVariantSpec creates a single-document spec with SW-001 (Base) and
SW-002 (Premium). Returns the spec root.

### IMP-cli-writeDisjointSpec-f42e401: writeDisjointSpec
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/testutil_test.go
x-reqmd.source-line: 89
x-reqmd.symbol-kind: function
trace: []
```

writeDisjointSpec creates a two-doc spec whose only trace link (SW-001
[Base] → SYS-001 [Premium]) has zero variant overlap. Returns the root.

### IMP-cli-writeDisjointCleanSpec-b19b8f4: writeDisjointCleanSpec
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/testutil_test.go
x-reqmd.source-line: 102
x-reqmd.symbol-kind: function
trace: []
```

writeDisjointCleanSpec is like writeDisjointSpec but the trace link
overlaps (SW-001 [Base, Premium] → SYS-001 [Premium]) and the sw
schema declares `x-reqmd.disjoint-check: variant`.

### IMP-cli-writeCoverageSpec-23183cf: writeCoverageSpec
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/testutil_test.go
x-reqmd.source-line: 115
x-reqmd.symbol-kind: function
trace: []
```

writeCoverageSpec creates a two-doc V-model: SYS-001 (Base) expects
coverage from the software level; its only provider SW-001 (Premium)
traces to it. Returns the root.

### IMP-cli-writeMeasureSpec-22b61dd: writeMeasureSpec
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/testutil_test.go
x-reqmd.source-line: 127
x-reqmd.symbol-kind: function
trace: []
```

writeMeasureSpec creates a single-doc spec with one Base measure
(verify: Test) that a manual result can trace to.

### IMP-cli-writeManualResult-1d27831: writeManualResult
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/testutil_test.go
x-reqmd.source-line: 137
x-reqmd.symbol-kind: function
trace: []
```

writeManualResult creates a manual-results dir with a single result of
the given outcome tracing to measureID. Returns the results dir.

### IMP-cli-gitHelper-0ba8471: gitHelper
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/testutil_test.go
x-reqmd.source-line: 157
x-reqmd.symbol-kind: function
trace: []
```

gitHelper returns a closure that runs git commands in dir with a
hermetic identity, failing the test on any error.

## reqmd-code/internal/cli/extract.go
### IMP-cli-extractFlags-e5f6c43: extractFlags
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract.go
x-reqmd.source-line: 27
x-reqmd.symbol-kind: type
trace: []
```

extractFlags carries the CLI flags for extraction. Using a struct
(rather than free-floating variables) keeps the flag state
self-contained and easy to pass to runExtract.

### IMP-cli-runExtract-7eb6f35: runExtract
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract.go
x-reqmd.source-line: 83
x-reqmd.symbol-kind: function
trace:
  - SW-IMP-001
```

reqmd:trace SW-IMP-001
runExtract is the cobra RunE body. Splitting it out of the constructor
keeps the flag wiring in NewRootCmd and the actual work here, so
this function is straightforward to call from a test if we ever want
to exercise the pipeline without going through cobra.

### IMP-cli-job-b70727a: job
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract.go
x-reqmd.source-line: 135
x-reqmd.symbol-kind: type
trace: []
```

Walk the source tree. We collect (file, plugin) pairs first, then
parse in a second pass; this decouples filesystem I/O from tree-
sitter parsing and keeps the loop body small.

### IMP-cli-parseResult-520e9af: parseResult
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract.go
x-reqmd.source-line: 185
x-reqmd.symbol-kind: type
trace: []
```

Parse every file in parallel, one result slot per job: no channel to
drain and no nil sentinel — a failed job just carries its error. Per-file
failures are non-fatal so one bad file doesn't lose the rest of the run.

Jobs are pre-sorted by path, and symbols are bucketed by package and
re-sorted by (file, line) below, so worker completion order never reaches
the output.

### IMP-cli-extSet-56e3a56: extSet
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract.go
x-reqmd.source-line: 465
x-reqmd.symbol-kind: function
trace: []
```

extSet is a tiny helper for the --lang filter: build a set of the
plugin's claimed extensions once, then look up by extension in O(1).
Keeping it local to this file avoids a public utility for one caller.

### IMP-cli-packagedSymbol-e464227: packagedSymbol
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract.go
x-reqmd.source-line: 475
x-reqmd.symbol-kind: type
trace: []
```

packagedSymbol is a symbol that survived per-package deduplication, with
its stable native ID.

### IMP-cli-traceRefError-29ddef1: traceRefError
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract.go
x-reqmd.source-line: 482
x-reqmd.symbol-kind: type
trace: []
```

traceRefError is one explicit `reqmd:trace` failure, addressed to the
declaring symbol.

### IMP-cli-formatTraceRefErrors-dde83ed: formatTraceRefErrors
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract.go
x-reqmd.source-line: 490
x-reqmd.symbol-kind: function
trace: []
```

formatTraceRefErrors renders the batched failures deterministically.

### IMP-cli-sortSymbols-a49eb78: sortSymbols
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract.go
x-reqmd.source-line: 515
x-reqmd.symbol-kind: function
trace: []
```

sortSymbols orders a package's symbols deterministically: by file, then
start line, then name. Methods therefore follow their parent type.

### IMP-cli-nativeIDMap-7b60f59: nativeIDMap
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract.go
x-reqmd.source-line: 530
x-reqmd.symbol-kind: function
trace: []
```

nativeIDMap maps every symbol's native key to its stable item key. The key
is `<Name>-<hash(pkg, nameOrReceiver.Name)>`: line-free, so edits above the
symbol do not churn IDs.

### IMP-cli-symbolTarget-59d7a46: symbolTarget
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract.go
x-reqmd.source-line: 540
x-reqmd.symbol-kind: type
trace: []
```

symbolTarget is a resolved explicit reference.

### IMP-cli-symbolIndex-0917079: symbolIndex
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract.go
x-reqmd.source-line: 548
x-reqmd.symbol-kind: type
trace: []
```

symbolIndex indexes deduplicated symbols by package and by the exact name
a `reqmd:trace` target can use: the bare name for non-methods, and
"Type.Name" for methods.

#### IMP-cli-resolve-f30272b: resolve
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract.go
x-reqmd.source-line: 573
x-reqmd.symbol-kind: method
trace: []
```

resolve matches an explicit symbol trace exactly: case-sensitive, no prefix
matching. `pkg/Name` and `pkg/Type.Name` search only that package; an
unqualified reference tries the same package, the base package of a test
package (`foo_test` -> `foo`, `test_foo` -> `foo`), then a globally unique
match. On ambiguity it returns the candidate names.

### IMP-cli-newSymbolIndex-3bda229: newSymbolIndex
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract.go
x-reqmd.source-line: 552
x-reqmd.symbol-kind: function
trace: []
```

### IMP-cli-basePackage-a2dd8be: basePackage
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract.go
x-reqmd.source-line: 633
x-reqmd.symbol-kind: function
trace: []
```

basePackage returns the non-test package a test package belongs to, or ""
when the name is not a test package.

### IMP-cli-conventionTarget-77ece95: conventionTarget
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract.go
x-reqmd.source-line: 644
x-reqmd.symbol-kind: type
trace: []
```

conventionTarget is a symbol a test name may refer to by convention.

### IMP-cli-buildConventionPool-8b1796c: buildConventionPool
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract.go
x-reqmd.source-line: 654
x-reqmd.symbol-kind: function
trace: []
```

buildConventionPool collects the symbols convention matching may link to:
production functions, methods, and types. Constants, variables, and tests
(including symbols in test files and inline `#[test]` items) are excluded so
a test can never link to another test or to a constant.

### IMP-cli-hasExplicitDirective-39ca6ca: hasExplicitDirective
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract.go
x-reqmd.source-line: 677
x-reqmd.symbol-kind: function
trace: []
```

hasExplicitDirective reports whether an item already carries an explicit
trace directive. Convention linking only fills the gap when nothing was
written by hand.

### IMP-cli-conventionLink-bdbe21c: conventionLink
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract.go
x-reqmd.source-line: 684
x-reqmd.symbol-kind: function
trace: []
```

conventionLink resolves a test symbol to the production symbol its name
refers to. Lookup is scoped to the same package, then the base package of a
test package — never globally, so unrelated packages cannot produce a link.

### IMP-cli-testStem-a7b827e: testStem
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract.go
x-reqmd.source-line: 715
x-reqmd.symbol-kind: function
trace: []
```

testStem returns the part of a test name that names its target, and whether
the language uses CamelCase (`TestParse`) or snake_case (`test_parse`).
Symbol kinds other than functions/methods never yield a stem.

### IMP-cli-matchConvention-cd50deb: matchConvention
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract.go
x-reqmd.source-line: 739
x-reqmd.symbol-kind: function
trace: []
```

matchConvention finds the longest candidate key that the test stem starts
with at a valid boundary. A tie across distinct targets is ambiguous and
returns the candidate names.

### IMP-cli-boundaryOK-1e734cf: boundaryOK
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract.go
x-reqmd.source-line: 788
x-reqmd.symbol-kind: function
trace: []
```

boundaryOK reports whether the text after a matched candidate is a valid
word boundary: end of name or an underscore for both styles, and a
CamelCase boundary for CamelCase names.

### IMP-cli-isUpperByte-dc15d2b: isUpperByte
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract.go
x-reqmd.source-line: 808
x-reqmd.symbol-kind: function
trace: []
```

### IMP-cli-isLowerByte-353f0c6: isLowerByte
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract.go
x-reqmd.source-line: 809
x-reqmd.symbol-kind: function
trace: []
```

### IMP-cli-nativeKey-89f17af: nativeKey
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract.go
x-reqmd.source-line: 814
x-reqmd.symbol-kind: function
trace: []
```

nativeKey is a symbol's disambiguation key inside its package. Methods are
keyed "Receiver.Name" so a method and a same-named top-level function don't
collapse into one requirement ID; everything else is keyed by name.

### IMP-cli-countRequirements-fa7b952: countRequirements
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract.go
x-reqmd.source-line: 822
x-reqmd.symbol-kind: function
trace: []
```

countRequirements counts requirement-kind items in a document tree.

### IMP-cli-withXReqmd-224aa1e: withXReqmd
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract.go
x-reqmd.source-line: 835
x-reqmd.symbol-kind: function
trace: []
```

withXReqmd returns a copy of doc whose x-reqmd block can be mutated without
affecting the other documents built from the same template.

### IMP-cli-nestMethods-5b30e95: nestMethods
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract.go
x-reqmd.source-line: 852
x-reqmd.symbol-kind: function
trace: []
```

nestMethods attaches collected methods to their type, preserving source order.
A method whose type lives in another file stays at its own file's level: the
file section is the authoritative grouping.

### IMP-cli-relPath-0b117f9: relPath
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract.go
x-reqmd.source-line: 867
x-reqmd.symbol-kind: function
trace: []
```

relPath renders a source file relative to the scanned root, so a per-file
section title is "parser.go" rather than an absolute build path. A file
outside the root keeps its full path.

### IMP-cli-docSections-43bde33: docSections
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract.go
x-reqmd.source-line: 882
x-reqmd.symbol-kind: function
trace: []
```

docSections turns a doc comment into section items: the prose before the
first heading becomes an untitled section (rendered as a plain prose block),
and every heading becomes a child section carrying the prose beneath it.
Headings are consumed here so no body ever contains one — the renderer's
contract is that structure lives in items, never in prose. Fenced code
blocks are skipped so a commented-out "#include" or a Markdown example is not
mistaken for a heading.

### IMP-cli-atxHeading-2b06ea4: atxHeading
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract.go
x-reqmd.source-line: 941
x-reqmd.symbol-kind: function
trace: []
```

atxHeading parses "## Title" into its level and text.

### IMP-cli-isRuleLine-1b1c203: isRuleLine
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract.go
x-reqmd.source-line: 954
x-reqmd.symbol-kind: function
trace: []
```

isRuleLine reports whether a line is a decorative rule: three or more of the
same Markdown rule character, nothing else.

## reqmd-code/internal/cli/extract_test.go
### IMP-cli-runTestExtract-4e9ac61: runTestExtract
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 24
x-reqmd.symbol-kind: function
trace: []
```

runTestExtract runs an extraction and returns the decoded bundle plus the
captured stdout/stderr text.

### IMP-cli-renderTo-67c9ee2: renderTo
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 44
x-reqmd.symbol-kind: function
trace: []
```

renderTo renders the bundle through reqmd-import's own build path — the
same code the real pipeline runs after the JSON hop — and returns the
rendered files keyed by slash-separated path.

### IMP-cli-docByDir-0dbd5c9: docByDir
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 69
x-reqmd.symbol-kind: function
trace: []
```

docByDir returns the bundle document that renders into the given directory.

### IMP-cli-allItems-6d195a2: allItems
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 81
x-reqmd.symbol-kind: function
trace: []
```

allItems flattens a document's item tree in render order.

### IMP-cli-TestRunExtract_HappyPath-574371b: TestRunExtract_HappyPath
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 90
x-reqmd.symbol-kind: function
trace:
  - IMP-cli-runExtract-7eb6f35
```

### IMP-cli-TestRunExtract_RejectsMissingSource-d630987: TestRunExtract_RejectsMissingSource
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 175
x-reqmd.symbol-kind: function
trace:
  - IMP-cli-runExtract-7eb6f35
```

### IMP-cli-TestRunExtract_RejectsFileAsSource-72cc925: TestRunExtract_RejectsFileAsSource
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 186
x-reqmd.symbol-kind: function
trace:
  - IMP-cli-runExtract-7eb6f35
```

### IMP-cli-TestRunExtract_RejectsUnknownLang-8dbede0: TestRunExtract_RejectsUnknownLang
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 199
x-reqmd.symbol-kind: function
trace:
  - IMP-cli-runExtract-7eb6f35
```

### IMP-cli-TestRunExtract_SkipsNoiseDirs-918e8cc: TestRunExtract_SkipsNoiseDirs
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 212
x-reqmd.symbol-kind: function
trace:
  - IMP-cli-runExtract-7eb6f35
```

### IMP-cli-TestRunExtract_LangFilter-d332500: TestRunExtract_LangFilter
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 232
x-reqmd.symbol-kind: function
trace:
  - IMP-cli-runExtract-7eb6f35
```

### IMP-cli-TestRunExtract_Deterministic-9e7d9d0: TestRunExtract_Deterministic
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 245
x-reqmd.symbol-kind: function
trace:
  - IMP-cli-runExtract-7eb6f35
```

### IMP-cli-TestExtSet-acc052a: TestExtSet
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 257
x-reqmd.symbol-kind: function
trace:
  - IMP-cli-extSet-56e3a56
```

### IMP-cli-TestRunExtract_NoSymbolsNotParseFailure-3044ec8: TestRunExtract_NoSymbolsNotParseFailure
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 267
x-reqmd.symbol-kind: function
trace:
  - IMP-cli-runExtract-7eb6f35
```

TestRunExtract_NoSymbolsNotParseFailure verifies that a file with no
extractable symbols (e.g. a package-doc-only .go file) is not counted
as a parse failure in the summary.

### IMP-cli-TestRunExtract_MultiLanguage-f737d9d: TestRunExtract_MultiLanguage
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 280
x-reqmd.symbol-kind: function
trace:
  - IMP-cli-runExtract-7eb6f35
```

TestRunExtract_MultiLanguage walks a source tree containing one file per
registered language and verifies that every language plugin self-registers
and claims its extension.

### IMP-cli-TestRunExtract_OutputFile-1921355: TestRunExtract_OutputFile
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 324
x-reqmd.symbol-kind: function
trace:
  - IMP-cli-runExtract-7eb6f35
```

TestRunExtract_OutputFile verifies -o writes the bundle to a file and keeps
stdout clean.

### IMP-cli-keys-e9c83db: keys
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 351
x-reqmd.symbol-kind: function
trace: []
```

### IMP-cli-TestDocSections-235bd2f: TestDocSections
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 362
x-reqmd.symbol-kind: function
trace:
  - IMP-cli-docSections-43bde33
```

TestDocSections covers moving doc comments out of requirement bodies: the
intro becomes an untitled section, every heading becomes a section, and a
heading inside a fenced block is left alone as prose.

### IMP-cli-TestRunExtract_DocCommentBecomesSection-a750dec: TestRunExtract_DocCommentBecomesSection
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 400
x-reqmd.symbol-kind: function
trace:
  - IMP-cli-runExtract-7eb6f35
```

TestRunExtract_DocCommentBecomesSection verifies the wiring end to end: the
symbol keeps the attr block and its doc prose moves into a child section, and
a Markdown heading in the doc does not break the rendered nesting.

### IMP-cli-TestDocSections_GodocSeparator-f91f647: TestDocSections_GodocSeparator
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 442
x-reqmd.symbol-kind: function
trace:
  - IMP-cli-docSections-43bde33
```

TestDocSections_GodocSeparator guards against a decorative rule becoming a
setext heading: "-----" under a text line renders as an h2, which would turn
a Godoc separator into a section that swallows the symbols after it.

### IMP-cli-runTestExtractErr-5ee29d8: runTestExtractErr
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 467
x-reqmd.symbol-kind: function
trace: []
```

runTestExtractErr runs an extraction that is expected to fail.

### IMP-cli-itemByTitle-26f19cd: itemByTitle
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 478
x-reqmd.symbol-kind: function
trace: []
```

itemByTitle finds one item in a document tree by its title.

### IMP-cli-TestRunExtract_ExplicitSymbolLink-68e02e2: TestRunExtract_ExplicitSymbolLink
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 491
x-reqmd.symbol-kind: function
trace:
  - IMP-cli-runExtract-7eb6f35
```

TestRunExtract_ExplicitSymbolLink covers the explicit marker path: a test
traces to a same-package function by name, and the link survives rendering.

### IMP-cli-TestRunExtract_MethodScopes-cf53ed2: TestRunExtract_MethodScopes
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 519
x-reqmd.symbol-kind: function
trace:
  - IMP-cli-runExtract-7eb6f35
```

TestRunExtract_MethodScopes covers receiver-qualified references and the
C++ `::` spelling.

### IMP-cli-TestRunExtract_BasePackageResolution-b90ca57: TestRunExtract_BasePackageResolution
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 549
x-reqmd.symbol-kind: function
trace:
  - IMP-cli-runExtract-7eb6f35
```

TestRunExtract_BasePackageResolution covers external test packages, where
the test lives in a different document from the symbol it tests.

### IMP-cli-TestRunExtract_UnresolvedSymbolIsFatal-64918f1: TestRunExtract_UnresolvedSymbolIsFatal
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 574
x-reqmd.symbol-kind: function
trace:
  - IMP-cli-runExtract-7eb6f35
```

TestRunExtract_UnresolvedSymbolIsFatal pins the hard-error contract: no
bundle, no stdout, and an existing -o file is left untouched.

### IMP-cli-TestRunExtract_AmbiguousSymbolIsFatal-ac93fec: TestRunExtract_AmbiguousSymbolIsFatal
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 614
x-reqmd.symbol-kind: function
trace:
  - IMP-cli-runExtract-7eb6f35
```

TestRunExtract_AmbiguousSymbolIsFatal covers a reference that matches the
same name in two packages.

### IMP-cli-TestRunExtract_MalformedMarkerIsFatal-d3e1009: TestRunExtract_MalformedMarkerIsFatal
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 641
x-reqmd.symbol-kind: function
trace:
  - IMP-cli-runExtract-7eb6f35
```

TestRunExtract_MalformedMarkerIsFatal covers prose left on the directive.

### IMP-cli-TestRunExtract_QualifiedAndOptOut-74add7a: TestRunExtract_QualifiedAndOptOut
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 659
x-reqmd.symbol-kind: function
trace:
  - IMP-cli-runExtract-7eb6f35
```

TestRunExtract_QualifiedAndOptOut covers an explicit package qualifier and
the `none` opt-out, which must not create a link or an error.

### IMP-cli-TestRunExtract_PythonTestLinksToModule-29fd858: TestRunExtract_PythonTestLinksToModule
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 689
x-reqmd.symbol-kind: function
trace:
  - IMP-cli-runExtract-7eb6f35
```

TestRunExtract_PythonTestLinksToModule covers a test file in its own
document reaching back to the module it tests.

### IMP-cli-TestRunExtract_RequirementIDRefsPassThrough-389c640: TestRunExtract_RequirementIDRefsPassThrough
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 711
x-reqmd.symbol-kind: function
trace:
  - IMP-cli-runExtract-7eb6f35
```

TestRunExtract_RequirementIDRefsPassThrough keeps the ID path intact: an
ID-shaped marker becomes a ref, not a symbol lookup, and renders into trace.

### IMP-cli-TestRunExtract_CaseMismatchIsFatal-b652bd8: TestRunExtract_CaseMismatchIsFatal
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 734
x-reqmd.symbol-kind: function
trace:
  - IMP-cli-runExtract-7eb6f35
```

TestRunExtract_CaseMismatchIsFatal pins that explicit matching is
case-sensitive: a near-miss is an error, not a silent skip.

### IMP-cli-TestRunExtract_BuildTagDuplicateNotAmbiguous-9263624: TestRunExtract_BuildTagDuplicateNotAmbiguous
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 753
x-reqmd.symbol-kind: function
trace:
  - IMP-cli-runExtract-7eb6f35
```

TestRunExtract_BuildTagDuplicateNotAmbiguous pins that build-tag twins are
deduplicated before indexing, so a reference to them resolves once.

### IMP-cli-TestRunExtract_DuplicateLinksDeduped-563ad25: TestRunExtract_DuplicateLinksDeduped
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 783
x-reqmd.symbol-kind: function
trace:
  - IMP-cli-runExtract-7eb6f35
```

TestRunExtract_DuplicateLinksDeduped pins that two spellings of the same
target produce one link.

### IMP-cli-TestFormatTraceRefErrors-a3aaa11: TestFormatTraceRefErrors
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 804
x-reqmd.symbol-kind: function
trace:
  - IMP-cli-formatTraceRefErrors-dde83ed
```

TestFormatTraceRefErrors pins deterministic sort order and the cap.

### IMP-cli-testLinks-90933f8: testLinks
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 834
x-reqmd.symbol-kind: function
trace: []
```

testLinks returns an item's links for assertions.

### IMP-cli-TestRunExtract_LinkTestsOffByDefault-b997766: TestRunExtract_LinkTestsOffByDefault
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 840
x-reqmd.symbol-kind: function
trace:
  - IMP-cli-runExtract-7eb6f35
```

TestRunExtract_LinkTestsOffByDefault pins that conventions are opt-in.

### IMP-cli-TestRunExtract_ConventionGo-d487a43: TestRunExtract_ConventionGo
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 860
x-reqmd.symbol-kind: function
trace:
  - IMP-cli-runExtract-7eb6f35
```

TestRunExtract_ConventionGo covers CamelCase matching: longest candidate,
CamelCase boundaries, initial-case folding, methods, types, and the target
pool exclusions.

### IMP-cli-TestRunExtract_ConventionGoExternalTestPackage-6a8f55c: TestRunExtract_ConventionGoExternalTestPackage
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 922
x-reqmd.symbol-kind: function
trace:
  - IMP-cli-runExtract-7eb6f35
```

TestRunExtract_ConventionGoExternalTestPackage covers the `_test` half of
basePackage: a test in package alpha_test links to package alpha.

### IMP-cli-TestRunExtract_ConventionPython-abc2a7f: TestRunExtract_ConventionPython
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 942
x-reqmd.symbol-kind: function
trace:
  - IMP-cli-runExtract-7eb6f35
```

TestRunExtract_ConventionPython covers snake_case across test documents.

### IMP-cli-TestRunExtract_ConventionRustInline-99d6be6: TestRunExtract_ConventionRustInline
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 960
x-reqmd.symbol-kind: function
trace:
  - IMP-cli-runExtract-7eb6f35
```

TestRunExtract_ConventionRustInline covers an inline #[test] module.

### IMP-cli-TestRunExtract_ConventionExplicitWins-9826cea: TestRunExtract_ConventionExplicitWins
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 979
x-reqmd.symbol-kind: function
trace:
  - IMP-cli-runExtract-7eb6f35
```

TestRunExtract_ConventionExplicitWins pins that any explicit directive
(an ID ref or none) suppresses convention linking.

### IMP-cli-TestRunExtract_ConventionAmbiguousWarns-2ddfda4: TestRunExtract_ConventionAmbiguousWarns
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 1007
x-reqmd.symbol-kind: function
trace:
  - IMP-cli-runExtract-7eb6f35
```

TestRunExtract_ConventionAmbiguousWarns pins non-fatal tie handling.

### IMP-cli-TestRunExtract_ConventionNoGlobalFallback-b664803: TestRunExtract_ConventionNoGlobalFallback
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/cli/extract_test.go
x-reqmd.source-line: 1033
x-reqmd.symbol-kind: function
trace:
  - IMP-cli-runExtract-7eb6f35
```

TestRunExtract_ConventionNoGlobalFallback pins that a test in one package
cannot link to a same-named symbol in an unrelated package.

## reqmd-import/internal/cli/root.go
### IMP-cli-runOptions-7da2fca: runOptions
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/cli/root.go
x-reqmd.source-line: 62
x-reqmd.symbol-kind: type
trace: []
```

runOptions carries the flags of one invocation.

### IMP-cli-parseStatusMap-306167e: parseStatusMap
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/cli/root.go
x-reqmd.source-line: 71
x-reqmd.symbol-kind: function
trace: []
```

parseStatusMap turns "Source=reqmd" pairs into a lookup table.

### IMP-cli-run-710e194: run
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/cli/root.go
x-reqmd.source-line: 86
x-reqmd.symbol-kind: function
trace: []
```

## reqmd-import/internal/cli/root_test.go
### IMP-cli-schemaURL-339429d: schemaURL
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/cli/root_test.go
x-reqmd.source-line: 11
x-reqmd.symbol-kind: const
trace: []
```

### IMP-cli-bundleJSON-158cd41: bundleJSON
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/cli/root_test.go
x-reqmd.source-line: 13
x-reqmd.symbol-kind: const
trace: []
```

### IMP-cli-TestRunFromFile-0822526: TestRunFromFile
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/cli/root_test.go
x-reqmd.source-line: 24
x-reqmd.symbol-kind: function
trace:
  - IMP-cli-run-710e194
```

### IMP-cli-TestRunFromStdin-1dd082a: TestRunFromStdin
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/cli/root_test.go
x-reqmd.source-line: 51
x-reqmd.symbol-kind: function
trace:
  - IMP-cli-run-710e194
```

### IMP-cli-TestRunValidateOnly-63cc920: TestRunValidateOnly
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/cli/root_test.go
x-reqmd.source-line: 70
x-reqmd.symbol-kind: function
trace:
  - IMP-cli-run-710e194
```

TestRunValidateOnly covers the conformance path for third-party producers:
report every violation, write nothing.

### IMP-cli-TestRunStatusMap-4f860ec: TestRunStatusMap
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/cli/root_test.go
x-reqmd.source-line: 109
x-reqmd.symbol-kind: function
trace:
  - IMP-cli-run-710e194
```

TestRunStatusMap covers the renderer-side status mapping flag.

### IMP-cli-TestRunRejectsBadBundle-0628673: TestRunRejectsBadBundle
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/cli/root_test.go
x-reqmd.source-line: 132
x-reqmd.symbol-kind: function
trace:
  - IMP-cli-run-710e194
```

