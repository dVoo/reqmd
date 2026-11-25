# Package: export

<!-- reqmd: generated -->

## internal/cli/export/csv.go
### IMP-export-newCsvCmd-b7e44bd: newCsvCmd
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/export/csv.go
x-reqmd.source-line: 17
x-reqmd.symbol-kind: function
trace: []
```

## internal/cli/export/export.go
### IMP-export-NewExportCmd-44591a9: NewExportCmd
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/export/export.go
x-reqmd.source-line: 10
x-reqmd.symbol-kind: function
trace: []
```

NewExportCmd builds the `reqmd export` command with its subcommands.

## internal/cli/export/graph.go
### IMP-export-newGraphCmd-34aaec5: newGraphCmd
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/export/graph.go
x-reqmd.source-line: 17
x-reqmd.symbol-kind: function
trace: []
```

## internal/cli/export/html.go
### IMP-export-outputDirPerm-499f0e1: outputDirPerm
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/export/html.go
x-reqmd.source-line: 23
x-reqmd.symbol-kind: const
trace: []
```

outputDirPerm keeps output directories readable by the exporting user's
group, following the CLI's 0750 convention.

### IMP-export-newHTMLCmd-d5201e8: newHTMLCmd
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/export/html.go
x-reqmd.source-line: 25
x-reqmd.symbol-kind: function
trace: []
```

### IMP-export-runExportHTML-0d522f9: runExportHTML
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/export/html.go
x-reqmd.source-line: 50
x-reqmd.symbol-kind: function
trace: []
```

runExportHTML writes one page per document plus the project home.

### IMP-export-discoverHTMLDocs-d84b278: discoverHTMLDocs
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/export/html.go
x-reqmd.source-line: 79
x-reqmd.symbol-kind: function
trace: []
```

discoverHTMLDocs parses the tree and, when a filter is active, returns the
matching requirement IDs for graph scoping plus the pruned export set. The
graph is still built from the full docs so trace links resolve.

### IMP-export-analyzeHTMLGraph-e1968e2: analyzeHTMLGraph
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/export/html.go
x-reqmd.source-line: 104
x-reqmd.symbol-kind: function
trace: []
```

analyzeHTMLGraph builds the trace graph and adapts it into the presentation
analysis every HTML view renders from. Ephemeral verification results are
synthesized into pseudo-requirements before the graph is built.

### IMP-export-htmlOutput-19b55b1: htmlOutput
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/export/html.go
x-reqmd.source-line: 125
x-reqmd.symbol-kind: type
trace: []
```

htmlOutput resolves output paths once for a document set and writes the
document pages and the project home.

#### IMP-export-resolveFor-2baaf0e: resolveFor
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/export/html.go
x-reqmd.source-line: 165
x-reqmd.symbol-kind: method
trace: []
```

resolveFor maps a requirement ID to an href relative to current's
directory, so trace links work in both output layouts.

#### IMP-export-writeDocumentPages-589104f: writeDocumentPages
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/export/html.go
x-reqmd.source-line: 186
x-reqmd.symbol-kind: method
trace: []
```

writeDocumentPages renders every exported document page, one worker per CPU
when there is more than one document. All of the expensive work (render data,
page model, body Markdown, the embedded view model JSON) is per document,
while the analysis, render context, boundaries and output maps are read-only
after construction, so they are shared. Each worker keeps its own
exporter.HTML, which carries the per-page state (boundary, document chain,
link resolver).

#### IMP-export-writeDocumentPage-c42d086: writeDocumentPage
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/export/html.go
x-reqmd.source-line: 238
x-reqmd.symbol-kind: method
trace: []
```

writeDocumentPage renders one document page, wiring its boundary badge,
document chain, and per-page trace link resolver. It returns the output path
so the caller can report it in document order.

#### IMP-export-writeProjectIndex-65ab62f: writeProjectIndex
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/export/html.go
x-reqmd.source-line: 257
x-reqmd.symbol-kind: method
trace: []
```

writeProjectIndex writes the project home page.

### IMP-export-newHTMLOutput-781a8ad: newHTMLOutput
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/export/html.go
x-reqmd.source-line: 137
x-reqmd.symbol-kind: function
trace: []
```

newHTMLOutput computes the output layout: flat files in outputDir when set,
otherwise one page inside each document directory.

### IMP-export-writeHTMLFile-7e55c53: writeHTMLFile
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/export/html.go
x-reqmd.source-line: 279
x-reqmd.symbol-kind: function
trace: []
```

writeHTMLFile creates path, runs write, and reports both the write and close
error so a truncated export is never reported as written.

## internal/cli/export/json.go
### IMP-export-tableFindings-91b72c1: tableFindings
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/export/json.go
x-reqmd.source-line: 19
x-reqmd.symbol-kind: const
trace: []
```

jsonTable names the four NDJSON files written by `export json`.

### IMP-export-tableRequirements-0aa7d47: tableRequirements
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/export/json.go
x-reqmd.source-line: 19
x-reqmd.symbol-kind: const
trace: []
```

jsonTable names the four NDJSON files written by `export json`.

### IMP-export-tableResults-a3efb1b: tableResults
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/export/json.go
x-reqmd.source-line: 19
x-reqmd.symbol-kind: const
trace: []
```

jsonTable names the four NDJSON files written by `export json`.

### IMP-export-tableTraces-ff7b141: tableTraces
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/export/json.go
x-reqmd.source-line: 19
x-reqmd.symbol-kind: const
trace: []
```

jsonTable names the four NDJSON files written by `export json`.

### IMP-export-newJSONCmd-bb9f639: newJSONCmd
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/cli/export/json.go
x-reqmd.source-line: 26
x-reqmd.symbol-kind: function
trace: []
```

