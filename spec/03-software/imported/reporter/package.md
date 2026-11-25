# Package: reporter

<!-- reqmd: generated -->

## internal/reporter/diff.go
### IMP-reporter-colorBold-1d71556: colorBold
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/diff.go
x-reqmd.source-line: 12
x-reqmd.symbol-kind: const
trace: []
```

### IMP-reporter-colorCyan-ce187da: colorCyan
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/diff.go
x-reqmd.source-line: 12
x-reqmd.symbol-kind: const
trace: []
```

### IMP-reporter-colorGreen-2c55b9c: colorGreen
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/diff.go
x-reqmd.source-line: 12
x-reqmd.symbol-kind: const
trace: []
```

### IMP-reporter-colorRed-a0f63fc: colorRed
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/diff.go
x-reqmd.source-line: 12
x-reqmd.symbol-kind: const
trace: []
```

### IMP-reporter-colorReset-7cd7b40: colorReset
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/diff.go
x-reqmd.source-line: 12
x-reqmd.symbol-kind: const
trace: []
```

### IMP-reporter-colorYellow-55a4909: colorYellow
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/diff.go
x-reqmd.source-line: 12
x-reqmd.symbol-kind: const
trace: []
```

### IMP-reporter-FormatDiff-2a9528e: FormatDiff
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/diff.go
x-reqmd.source-line: 22
x-reqmd.symbol-kind: function
trace: []
```

FormatDiff produces a human-readable colored text diff.

### IMP-reporter-shortHash-b263d54: shortHash
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/diff.go
x-reqmd.source-line: 110
x-reqmd.symbol-kind: function
trace: []
```

shortHash returns the first 7 characters of a SHA, or the full SHA if shorter.

### IMP-reporter-formatSubmoduleChange-bdeb209: formatSubmoduleChange
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/diff.go
x-reqmd.source-line: 118
x-reqmd.symbol-kind: function
trace: []
```

formatSubmoduleChange returns the formatted line and color for a submodule change.

### IMP-reporter-renderSubmoduleSection-dcf35fb: renderSubmoduleSection
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/diff.go
x-reqmd.source-line: 133
x-reqmd.symbol-kind: function
trace: []
```

renderSubmoduleSection renders the submodule changes section.
Returns an empty string if there are no submodule changes.

### IMP-reporter-FormatDiffJSON-ba26c18: FormatDiffJSON
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/diff.go
x-reqmd.source-line: 149
x-reqmd.symbol-kind: function
trace: []
```

FormatDiffJSON produces JSON output.

### IMP-reporter-formatValue-a56317e: formatValue
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/diff.go
x-reqmd.source-line: 158
x-reqmd.symbol-kind: function
trace: []
```

formatValue renders an attribute value for display.

## internal/reporter/diff\_test.go
### IMP-reporter-TestFormatDiff\_Submodules-7ebb990: TestFormatDiff\_Submodules
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/diff_test.go
x-reqmd.source-line: 10
x-reqmd.symbol-kind: function
trace:
  - IMP-reporter-FormatDiff-2a9528e
```

### IMP-reporter-TestFormatDiff\_NoSubmodules-321d12f: TestFormatDiff\_NoSubmodules
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/diff_test.go
x-reqmd.source-line: 39
x-reqmd.symbol-kind: function
trace:
  - IMP-reporter-FormatDiff-2a9528e
```

### IMP-reporter-TestShortHash-38bc934: TestShortHash
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/diff_test.go
x-reqmd.source-line: 51
x-reqmd.symbol-kind: function
trace:
  - IMP-reporter-shortHash-b263d54
```

## internal/reporter/reporter.go
### IMP-reporter-ValidationError-7751545: ValidationError
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter.go
x-reqmd.source-line: 16
x-reqmd.symbol-kind: type
trace: []
```

ValidationError describes a single requirement that failed validation.

### IMP-reporter-ParseError-d4d4381: ParseError
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter.go
x-reqmd.source-line: 23
x-reqmd.symbol-kind: type
trace: []
```

ParseError describes a structural problem (file unreadable, bad YAML, etc.).

### IMP-reporter-ExitCodeError-1585eb7: ExitCodeError
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter.go
x-reqmd.source-line: 30
x-reqmd.symbol-kind: type
trace: []
```

ExitCodeError wraps a desired process exit code so that main.go
can set exit code via cobra error handling instead of os.Exit in command handlers.

#### IMP-reporter-Error-d7589b1: Error
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter.go
x-reqmd.source-line: 34
x-reqmd.symbol-kind: method
trace: []
```

#### IMP-reporter-ExitCode-0014db4: ExitCode
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter.go
x-reqmd.source-line: 39
x-reqmd.symbol-kind: method
trace: []
```

ExitCode returns the process exit code this error carries.

### IMP-reporter-DocHeader-4245769: DocHeader
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter.go
x-reqmd.source-line: 42
x-reqmd.symbol-kind: type
trace: []
```

DocHeader holds per-document metadata for the validate output header.

### IMP-reporter-Report-2c840c8: Report
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter.go
x-reqmd.source-line: 50
x-reqmd.symbol-kind: type
trace: []
```

Report holds all results from a validation run.

#### IMP-reporter-NewIndex-266d9a6: NewIndex
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter.go
x-reqmd.source-line: 65
x-reqmd.symbol-kind: method
trace: []
```

NewIndex populates the pre-indexed maps from the flat slices.
Bucketing by reqID makes the per-reqID lookups in Format()/FormatJSON() O(1)
instead of scanning all errors/checks per requirement.

#### IMP-reporter-ExitCode-22aced1: ExitCode
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter.go
x-reqmd.source-line: 81
x-reqmd.symbol-kind: method
trace: []
```

ExitCode returns the appropriate exit code per the spec:

	0 = all valid
	1 = validation errors
	2 = parse errors

#### IMP-reporter-counts-5083288: counts
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter.go
x-reqmd.source-line: 99
x-reqmd.symbol-kind: method
trace: []
```

counts tallies WARNING and ERROR-level graph checks in a single pass.
Shared by ExitCode, Format, and FormatJSON so the summary numbers are
computed the same way everywhere.

#### IMP-reporter-resultFindings-da40197: resultFindings
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter.go
x-reqmd.source-line: 116
x-reqmd.symbol-kind: method
trace: []
```

resultFindings returns graph checks that belong to the verification
results layer rather than to an authored requirement: findings attributed
to synthesized nodes (RESULT:/TC: IDs) and loader warnings attached
without a ReqID (e.g. unbound-result). These are rendered in a dedicated
results section, never as requirement-attributed findings.

#### IMP-reporter-Format-dde0447: Format
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter.go
x-reqmd.source-line: 131
x-reqmd.symbol-kind: method
trace: []
```

Format produces the per-file, per-requirement validation report
as specified in the spec tree under spec/.

#### IMP-reporter-FormatJSON-7d0f915: FormatJSON
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter.go
x-reqmd.source-line: 480
x-reqmd.symbol-kind: method
trace: []
```

FormatJSON serializes the full check report as JSON.

### IMP-reporter-FormatList-a6f5d1d: FormatList
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter.go
x-reqmd.source-line: 266
x-reqmd.symbol-kind: function
trace: []
```

FormatList produces a text table for `reqmd list`. Every node appears
in document order with a leading Type column: requirements carry their
ID (+ title) and attributes, containers and info items show their
heading text with empty attribute cells.

### IMP-reporter-FormatStats-2b22ce0: FormatStats
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter.go
x-reqmd.source-line: 312
x-reqmd.symbol-kind: function
trace: []
```

FormatStats produces a stats breakdown. The per-document header shows
requirement and item counts; a type breakdown section precedes the
attribute-value breakdowns.

### IMP-reporter-DocumentSummary-360b609: DocumentSummary
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter.go
x-reqmd.source-line: 358
x-reqmd.symbol-kind: type
trace: []
```

DocumentSummary is the data needed for list/stats display.

### IMP-reporter-NodeRow-a2eea22: NodeRow
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter.go
x-reqmd.source-line: 368
x-reqmd.symbol-kind: type
trace: []
```

NodeRow is a single row in list/stats output: a requirement or an
information item (container/info), in document order.

### IMP-reporter-attributeCounts-eb7d80a: attributeCounts
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter.go
x-reqmd.source-line: 379
x-reqmd.symbol-kind: function
trace: []
```

attributeCounts tallies string attribute values per declared property
across a document's rows. Properties with no string values are omitted.
Shared by the text and JSON stats formatters so both report identical
breakdowns.

### IMP-reporter-sortedKeys-a407aab: sortedKeys
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter.go
x-reqmd.source-line: 395
x-reqmd.symbol-kind: function
trace: []
```

### IMP-reporter-jsonReport-d472f7e: jsonReport
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter.go
x-reqmd.source-line: 408
x-reqmd.symbol-kind: type
trace: []
```

JSON output types (private — serialized via json.Marshal)

### IMP-reporter-jsonSummary-86ada34: jsonSummary
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter.go
x-reqmd.source-line: 417
x-reqmd.symbol-kind: type
trace: []
```

### IMP-reporter-jsonDocSection-6795f5d: jsonDocSection
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter.go
x-reqmd.source-line: 426
x-reqmd.symbol-kind: type
trace: []
```

### IMP-reporter-jsonReqResult-b005c6d: jsonReqResult
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter.go
x-reqmd.source-line: 433
x-reqmd.symbol-kind: type
trace: []
```

### IMP-reporter-jsonChk-655adcc: jsonChk
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter.go
x-reqmd.source-line: 439
x-reqmd.symbol-kind: type
trace: []
```

### IMP-reporter-jsonParseErr-341f8fe: jsonParseErr
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter.go
x-reqmd.source-line: 449
x-reqmd.symbol-kind: type
trace: []
```

### IMP-reporter-jsonListReport-0de3b56: jsonListReport
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter.go
x-reqmd.source-line: 454
x-reqmd.symbol-kind: type
trace: []
```

### IMP-reporter-jsonListDoc-8485cea: jsonListDoc
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter.go
x-reqmd.source-line: 458
x-reqmd.symbol-kind: type
trace: []
```

### IMP-reporter-jsonListRow-72c7e70: jsonListRow
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter.go
x-reqmd.source-line: 464
x-reqmd.symbol-kind: type
trace: []
```

### IMP-reporter-jsonStatsDoc-6293ad3: jsonStatsDoc
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter.go
x-reqmd.source-line: 472
x-reqmd.symbol-kind: type
trace: []
```

### IMP-reporter-FormatListJSON-be34f2b: FormatListJSON
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter.go
x-reqmd.source-line: 579
x-reqmd.symbol-kind: function
trace: []
```

FormatListJSON serializes list output as JSON. All nodes appear in
document order in the `rows` array with a `type` discriminator.

### IMP-reporter-FormatStatsJSON-e3ea6f2: FormatStatsJSON
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter.go
x-reqmd.source-line: 610
x-reqmd.symbol-kind: function
trace: []
```

FormatStatsJSON serializes stats output as JSON.

### IMP-reporter-jsonStatsOut-e998dd9: jsonStatsOut
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter.go
x-reqmd.source-line: 611
x-reqmd.symbol-kind: type
trace: []
```

## internal/reporter/reporter\_test.go
### IMP-reporter-mustUnmarshal-cbdc45e: mustUnmarshal
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter_test.go
x-reqmd.source-line: 12
x-reqmd.symbol-kind: function
trace: []
```

mustUnmarshal decodes JSON string into a map for assertion checks.

### IMP-reporter-TestExitCode-67062c7: TestExitCode
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter_test.go
x-reqmd.source-line: 23
x-reqmd.symbol-kind: function
trace: []
```

--- ExitCode ---

### IMP-reporter-TestFormat\_Empty-4b3c13f: TestFormat\_Empty
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter_test.go
x-reqmd.source-line: 85
x-reqmd.symbol-kind: function
trace:
  - IMP-reporter-Format-dde0447
```

--- Format (empty) ---

### IMP-reporter-TestFormat\_WithValidationErrors-0fb0226: TestFormat\_WithValidationErrors
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter_test.go
x-reqmd.source-line: 95
x-reqmd.symbol-kind: function
trace:
  - IMP-reporter-Format-dde0447
```

--- Format (validation errors) ---

### IMP-reporter-TestFormat\_WithNewFormat-2c91a63: TestFormat\_WithNewFormat
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter_test.go
x-reqmd.source-line: 133
x-reqmd.symbol-kind: function
trace:
  - IMP-reporter-Format-dde0447
```

--- Format (new format with DocHeaders) ---

### IMP-reporter-TestFormat\_DuplicateReqIDPrintsFindingsOnce-1ad55df: TestFormat\_DuplicateReqIDPrintsFindingsOnce
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter_test.go
x-reqmd.source-line: 189
x-reqmd.symbol-kind: function
trace:
  - IMP-reporter-Format-dde0447
```

--- Duplicate requirement IDs ---
A document holding the same requirement ID three times lists that ID three
times in ReqIDs; every finding bucketed to the ID must still print once.

### IMP-reporter-TestFormatJSON\_DuplicateReqIDRendersOnce-ce6bf9a: TestFormatJSON\_DuplicateReqIDRendersOnce
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter_test.go
x-reqmd.source-line: 213
x-reqmd.symbol-kind: function
trace:
  - IMP-reporter-FormatJSON-7d0f915
```

### IMP-reporter-TestFormat\_EmptyWithDocHeaders-654afa3: TestFormat\_EmptyWithDocHeaders
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter_test.go
x-reqmd.source-line: 237
x-reqmd.symbol-kind: function
trace:
  - IMP-reporter-Format-dde0447
```

--- Format (empty with DocHeaders) ---

### IMP-reporter-TestExitCode\_GraphErrorChangesExitCode-b8da1e6: TestExitCode\_GraphErrorChangesExitCode
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter_test.go
x-reqmd.source-line: 262
x-reqmd.symbol-kind: function
trace: []
```

--- ExitCode with graph errors ---

### IMP-reporter-TestFormat\_WithParseErrors-219535a: TestFormat\_WithParseErrors
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter_test.go
x-reqmd.source-line: 277
x-reqmd.symbol-kind: function
trace:
  - IMP-reporter-Format-dde0447
```

--- Format (parse errors) ---

### IMP-reporter-TestFormatList-7bfaedd: TestFormatList
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter_test.go
x-reqmd.source-line: 303
x-reqmd.symbol-kind: function
trace:
  - IMP-reporter-FormatList-a6f5d1d
```

--- FormatList ---

### IMP-reporter-TestFormatList\_WithTitle-3a5f904: TestFormatList\_WithTitle
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter_test.go
x-reqmd.source-line: 352
x-reqmd.symbol-kind: function
trace:
  - IMP-reporter-FormatList-a6f5d1d
```

### IMP-reporter-TestFormatListJSON\_WithTitle-ce5e1e2: TestFormatListJSON\_WithTitle
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter_test.go
x-reqmd.source-line: 375
x-reqmd.symbol-kind: function
trace:
  - IMP-reporter-FormatListJSON-be34f2b
```

### IMP-reporter-TestFormatStats-fe92ba0: TestFormatStats
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter_test.go
x-reqmd.source-line: 413
x-reqmd.symbol-kind: function
trace:
  - IMP-reporter-FormatStats-2b22ce0
```

--- FormatStats ---

### IMP-reporter-TestFormatJSON\_Simple-25c543a: TestFormatJSON\_Simple
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter_test.go
x-reqmd.source-line: 455
x-reqmd.symbol-kind: function
trace:
  - IMP-reporter-FormatJSON-7d0f915
```

--- JSON output ---

### IMP-reporter-TestFormatJSON\_Empty-cedf301: TestFormatJSON\_Empty
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter_test.go
x-reqmd.source-line: 540
x-reqmd.symbol-kind: function
trace:
  - IMP-reporter-FormatJSON-7d0f915
```

### IMP-reporter-TestFormatJSON\_ParseErrors-9e07194: TestFormatJSON\_ParseErrors
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter_test.go
x-reqmd.source-line: 558
x-reqmd.symbol-kind: function
trace:
  - IMP-reporter-FormatJSON-7d0f915
```

### IMP-reporter-TestFormatJSON\_GraphErrors-936fba3: TestFormatJSON\_GraphErrors
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter_test.go
x-reqmd.source-line: 580
x-reqmd.symbol-kind: function
trace:
  - IMP-reporter-FormatJSON-7d0f915
```

### IMP-reporter-TestFormatListJSON-6daeeeb: TestFormatListJSON
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter_test.go
x-reqmd.source-line: 604
x-reqmd.symbol-kind: function
trace:
  - IMP-reporter-FormatListJSON-be34f2b
```

### IMP-reporter-TestFormatStatsJSON-8a521b2: TestFormatStatsJSON
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter_test.go
x-reqmd.source-line: 636
x-reqmd.symbol-kind: function
trace:
  - IMP-reporter-FormatStatsJSON-e3ea6f2
```

### IMP-reporter-TestFormat\_ResultsSection-45a5597: TestFormat\_ResultsSection
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/reporter/reporter_test.go
x-reqmd.source-line: 670
x-reqmd.symbol-kind: function
trace:
  - IMP-reporter-Format-dde0447
```

