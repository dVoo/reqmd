# Package: main

<!-- reqmd: generated -->

## cmd/reqmd-bench-gen/main.go
### IMP-main-placeholderFmt-37c3dfa: placeholderFmt
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-gen/main.go
x-reqmd.source-line: 39
x-reqmd.symbol-kind: const
trace: []
```

### IMP-main-outputDirPerm-cfba3d4: outputDirPerm
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-gen/main.go
x-reqmd.source-line: 42
x-reqmd.symbol-kind: const
trace: []
```

Generated corpora are private to the exporting user.

### IMP-main-outputFilePerm-baa63e3: outputFilePerm
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-gen/main.go
x-reqmd.source-line: 42
x-reqmd.symbol-kind: const
trace: []
```

Generated corpora are private to the exporting user.

### IMP-main-main-93bd65b: main
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-gen/main.go
x-reqmd.source-line: 47
x-reqmd.symbol-kind: function
trace: []
```

### IMP-main-rewriteTree-0a497be: rewriteTree
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-gen/main.go
x-reqmd.source-line: 156
x-reqmd.symbol-kind: function
trace: []
```

### IMP-main-replicatePerDoc-aa9b58a: replicatePerDoc
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-gen/main.go
x-reqmd.source-line: 203
x-reqmd.symbol-kind: function
trace: []
```

replicatePerDoc walks each top-level doc dir in `root` and, for every
 .md file, repeats the requirement blocks `multiplier - 1` more times.
 Each repeated block has its IDs and trace references rewritten to keep
 the trace graph connected.

### IMP-main-expandContent-2e9ff07: expandContent
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-gen/main.go
x-reqmd.source-line: 257
x-reqmd.symbol-kind: function
trace: []
```

expandContent extracts the frontmatter, the preamble, and the heading
 blocks from `body`, then concatenates `multiplier` copies of the heading
 blocks. The first copy uses the original IDs; subsequent copies have
 their IDs and references rewritten.

### IMP-main-splitOnH2-aacd271: splitOnH2
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-gen/main.go
x-reqmd.source-line: 288
x-reqmd.symbol-kind: function
trace: []
```

splitOnH2 splits body on lines starting with "## ". The first chunk
 includes the preamble (possibly empty); subsequent chunks start with
 "## " (preserved as their first line).

### IMP-main-rewriteBlockForCopy-d34487f: rewriteBlockForCopy
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-gen/main.go
x-reqmd.source-line: 310
x-reqmd.symbol-kind: function
trace: []
```

rewriteBlockForCopy returns a copy of `block` with the heading ID
 rewritten to have a "_c<copyIdx>" suffix, and the `trace:` and
 `requires-trace-from:` lines rewritten to point at the renamed IDs.
 Other fields are left untouched (we don't want to corrupt enum values
 like `process: SYS` or `status: approved`).

### IMP-main-rewriteIDListInPlace-239c077: rewriteIDListInPlace
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-gen/main.go
x-reqmd.source-line: 360
x-reqmd.symbol-kind: function
trace: []
```

rewriteIDListInPlace rewrites each ID token in a comma-separated list
 like "SYS-001, stakeholder/SYS-001~3" by appending `suffix` to the ID
 portion.

### IMP-main-applyRewrite-bf0dc4a: applyRewrite
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-gen/main.go
x-reqmd.source-line: 387
x-reqmd.symbol-kind: function
trace: []
```

applyRewrite is the per-copy rename pass. Same as rewriteBlockForCopy
 in spirit but operates on the whole file body and uses the per-copy
 rename map (e.g. "SYS-001" -> "T00-SYS-001").

## cmd/reqmd-bench-run/main.go
### IMP-main-scale-9176ac8: scale
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-run/main.go
x-reqmd.source-line: 32
x-reqmd.symbol-kind: type
trace: []
```

### IMP-main-measurement-ac5778e: measurement
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-run/main.go
x-reqmd.source-line: 38
x-reqmd.symbol-kind: type
trace: []
```

### IMP-main-hardware-b0842f6: hardware
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-run/main.go
x-reqmd.source-line: 52
x-reqmd.symbol-kind: type
trace: []
```

### IMP-main-report-05e6068: report
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-run/main.go
x-reqmd.source-line: 64
x-reqmd.symbol-kind: type
trace: []
```

### IMP-main-peakRSS-242622f: peakRSS
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-run/main.go
x-reqmd.source-line: 73
x-reqmd.symbol-kind: var
trace: []
```

### IMP-main-cmd-df5a3aa: cmd
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-run/main.go
x-reqmd.source-line: 216
x-reqmd.symbol-kind: type
trace: []
```

### IMP-main-benchCommands-42e1ab6: benchCommands
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-run/main.go
x-reqmd.source-line: 221
x-reqmd.symbol-kind: function
trace: []
```

### IMP-main-measureN-b937a80: measureN
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-run/main.go
x-reqmd.source-line: 233
x-reqmd.symbol-kind: function
trace: []
```

measureN runs the command `runs` times and returns a measurement with the
 median wall time and the peak RSS across all runs.

### IMP-main-flagVerbose-964213a: flagVerbose
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-run/main.go
x-reqmd.source-line: 262
x-reqmd.symbol-kind: var
trace: []
```

### IMP-main-measureOnce-ddf35cb: measureOnce
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-run/main.go
x-reqmd.source-line: 264
x-reqmd.symbol-kind: function
trace: []
```

### IMP-main-countReqs-ef98ccc: countReqs
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-run/main.go
x-reqmd.source-line: 286
x-reqmd.symbol-kind: function
trace: []
```

### IMP-main-countDirs-8b5f4a1: countDirs
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-run/main.go
x-reqmd.source-line: 315
x-reqmd.symbol-kind: function
trace: []
```

### IMP-main-sampleRSS-c8fc27c: sampleRSS
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-run/main.go
x-reqmd.source-line: 328
x-reqmd.symbol-kind: function
trace: []
```

### IMP-main-readRSSKB-b69e42f: readRSSKB
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-run/main.go
x-reqmd.source-line: 344
x-reqmd.symbol-kind: function
trace: []
```

### IMP-main-collectHardware-587ae74: collectHardware
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-run/main.go
x-reqmd.source-line: 364
x-reqmd.symbol-kind: function
trace: []
```

collectHardware gathers CPU, OS, and runtime metadata for reproducibility.

## cmd/reqmd/main.go
### IMP-main-exitCoder-3b8cb2f: exitCoder
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd/main.go
x-reqmd.source-line: 19
x-reqmd.symbol-kind: type
trace: []
```

exitCoder is implemented by errors that carry a process exit code
 (reporter.ExitCodeError). main depends on the behavior, not the
 concrete type, so the entry point stays decoupled from the reporter
 package.

### IMP-main-version-610805e: version
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd/main.go
x-reqmd.source-line: 28
x-reqmd.symbol-kind: var
trace: []
```

version is overridden at release time via:

	go build -ldflags "-X main.version=v0.1.0"

 When empty (dev builds), `reqmd --version` reports "(devel)".

## reqmd-code/cmd/reqmd-code/main.go
## reqmd-import/cmd/reqmd-import/main.go
