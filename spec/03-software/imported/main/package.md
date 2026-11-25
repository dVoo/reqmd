# Package: main

<!-- reqmd: generated -->

## cmd/reqmd-bench-gen/main.go
### IMP-main-attrFence-16ccbea: attrFence
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-gen/main.go
x-reqmd.source-line: 40
x-reqmd.symbol-kind: const
trace: []
```

attrFence marks a requirement heading in the generated Markdown.

### IMP-main-placeholderFmt-37c3dfa: placeholderFmt
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-gen/main.go
x-reqmd.source-line: 42
x-reqmd.symbol-kind: const
trace: []
```

### IMP-main-outputDirPerm-cfba3d4: outputDirPerm
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-gen/main.go
x-reqmd.source-line: 45
x-reqmd.symbol-kind: const
trace: []
```

Generated corpora are private to the exporting user.

### IMP-main-outputFilePerm-baa63e3: outputFilePerm
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-gen/main.go
x-reqmd.source-line: 45
x-reqmd.symbol-kind: const
trace: []
```

Generated corpora are private to the exporting user.

### IMP-main-main-93bd65b: main
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-gen/main.go
x-reqmd.source-line: 50
x-reqmd.symbol-kind: function
trace: []
```

### IMP-main-hasAttrBlock-dfc2c6e: hasAttrBlock
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-gen/main.go
x-reqmd.source-line: 162
x-reqmd.symbol-kind: function
trace: []
```

hasAttrBlock reports whether the heading on line i is followed by an
`reqmd` fence, which is what makes it a requirement heading.

### IMP-main-rewriteTree-0a497be: rewriteTree
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-gen/main.go
x-reqmd.source-line: 173
x-reqmd.symbol-kind: function
trace: []
```

### IMP-main-replicatePerDoc-aa9b58a: replicatePerDoc
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-gen/main.go
x-reqmd.source-line: 220
x-reqmd.symbol-kind: function
trace: []
```

replicatePerDoc walks each top-level doc dir in `root` and, for every
.md file, repeats the requirement blocks `multiplier - 1` more times.
Each repeated block has its IDs and trace references rewritten to keep
the trace graph connected.

### IMP-main-expandContent-2e9ff07: expandContent
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-gen/main.go
x-reqmd.source-line: 274
x-reqmd.symbol-kind: function
trace: []
```

expandContent extracts the frontmatter, the preamble, and the heading
blocks from `body`, then concatenates `multiplier` copies of the heading
blocks. The first copy uses the original IDs; subsequent copies have
their IDs and references rewritten.

### IMP-main-splitOnH2-aacd271: splitOnH2
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-gen/main.go
x-reqmd.source-line: 305
x-reqmd.symbol-kind: function
trace: []
```

splitOnH2 splits body on lines starting with "## ". The first chunk
includes the preamble (possibly empty); subsequent chunks start with
"## " (preserved as their first line).

### IMP-main-rewriteBlockForCopy-d34487f: rewriteBlockForCopy
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-gen/main.go
x-reqmd.source-line: 327
x-reqmd.symbol-kind: function
trace: []
```

rewriteBlockForCopy returns a copy of `block` with the heading ID
rewritten to have a "_c<copyIdx>" suffix, and the `trace:` and
`requires-trace-from:` lines rewritten to point at the renamed IDs.
Other fields are left untouched (we don't want to corrupt enum values
like `process: SYS` or `status: approved`).

### IMP-main-rewriteIDListInPlace-239c077: rewriteIDListInPlace
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-gen/main.go
x-reqmd.source-line: 379
x-reqmd.symbol-kind: function
trace: []
```

rewriteIDListInPlace rewrites each ID token in a comma-separated list
like "SYS-001, stakeholder/SYS-001~3" by appending `suffix` to the ID
portion.

### IMP-main-applyRewrite-bf0dc4a: applyRewrite
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-gen/main.go
x-reqmd.source-line: 406
x-reqmd.symbol-kind: function
trace: []
```

applyRewrite is the per-copy rename pass. Same as rewriteBlockForCopy
in spirit but operates on the whole file body and uses the per-copy
rename map (e.g. "SYS-001" -> "T00-SYS-001").

## cmd/reqmd-bench-gen/main\_test.go
### IMP-main-TestHeadingIDReMatchesRequirementHeadingLevels-e509041: TestHeadingIDReMatchesRequirementHeadingLevels
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-gen/main_test.go
x-reqmd.source-line: 7
x-reqmd.symbol-kind: function
trace: []
```

The source spec keeps most requirement IDs on h3/h4 headings, so every
heading level the parser accepts as a requirement must be discovered.

### IMP-main-TestApplyRewritePreservesHeadingLevel-545c287: TestApplyRewritePreservesHeadingLevel
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-gen/main_test.go
x-reqmd.source-line: 45
x-reqmd.symbol-kind: function
trace:
  - IMP-main-applyRewrite-bf0dc4a
```

### IMP-main-TestApplyRewriteTagsSchemaPrefixOnce-37b9fa1: TestApplyRewriteTagsSchemaPrefixOnce
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-gen/main_test.go
x-reqmd.source-line: 56
x-reqmd.symbol-kind: function
trace:
  - IMP-main-applyRewrite-bf0dc4a
```

A document's declared prefix must receive the copy tag exactly once. Tagging
it twice makes every requirement in that document fail validation.

### IMP-main-TestRewriteBlockForCopyPreservesHeadingLevel-9cd553c: TestRewriteBlockForCopyPreservesHeadingLevel
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-gen/main_test.go
x-reqmd.source-line: 65
x-reqmd.symbol-kind: function
trace:
  - IMP-main-rewriteBlockForCopy-d34487f
```

### IMP-main-TestRewriteBlockForCopySkipsProseHeading-6186a62: TestRewriteBlockForCopySkipsProseHeading
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-gen/main_test.go
x-reqmd.source-line: 77
x-reqmd.symbol-kind: function
trace:
  - IMP-main-rewriteBlockForCopy-d34487f
```

A prose heading that happens to contain a dash is not a requirement and must
be repeated verbatim.

## cmd/reqmd-bench-run/main.go
### IMP-main-scale-9176ac8: scale
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-run/main.go
x-reqmd.source-line: 35
x-reqmd.symbol-kind: type
trace: []
```

### IMP-main-measurement-ac5778e: measurement
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-run/main.go
x-reqmd.source-line: 42
x-reqmd.symbol-kind: type
trace: []
```

### IMP-main-hardware-b0842f6: hardware
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-run/main.go
x-reqmd.source-line: 56
x-reqmd.symbol-kind: type
trace: []
```

### IMP-main-report-05e6068: report
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-run/main.go
x-reqmd.source-line: 68
x-reqmd.symbol-kind: type
trace: []
```

### IMP-main-peakRSS-242622f: peakRSS
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-run/main.go
x-reqmd.source-line: 77
x-reqmd.symbol-kind: var
trace: []
```

### IMP-main-cmd-df5a3aa: cmd
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-run/main.go
x-reqmd.source-line: 215
x-reqmd.symbol-kind: type
trace: []
```

### IMP-main-benchCommands-42e1ab6: benchCommands
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-run/main.go
x-reqmd.source-line: 220
x-reqmd.symbol-kind: function
trace: []
```

### IMP-main-measureN-b937a80: measureN
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-run/main.go
x-reqmd.source-line: 232
x-reqmd.symbol-kind: function
trace: []
```

measureN runs the command `runs` times and returns a measurement with the
median wall time and the peak RSS across all runs.

### IMP-main-flagVerbose-964213a: flagVerbose
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-run/main.go
x-reqmd.source-line: 261
x-reqmd.symbol-kind: var
trace: []
```

### IMP-main-measureOnce-ddf35cb: measureOnce
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-run/main.go
x-reqmd.source-line: 263
x-reqmd.symbol-kind: function
trace: []
```

### IMP-main-reqHeadingRe-6de9915: reqHeadingRe
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-run/main.go
x-reqmd.source-line: 287
x-reqmd.symbol-kind: var
trace: []
```

reqHeadingRe matches a heading that can carry a requirement ID. The dash is
required so a prose heading like "Notes" is not mistaken for an ID.

### IMP-main-scaleTargets-648bf86: scaleTargets
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-run/main.go
x-reqmd.source-line: 293
x-reqmd.symbol-kind: function
trace: []
```

scaleTargets turns total requirement targets into the generator's two knobs.
A copy multiplies the document tree; per-doc replication multiplies each
document's content. The corpus therefore ends up with roughly
copies x reqsPerDoc x srcDocDirs requirements, which is what is solved for.

### IMP-main-countTree-f3291b1: countTree
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-run/main.go
x-reqmd.source-line: 314
x-reqmd.symbol-kind: function
trace: []
```

countTree reports how many requirements and document directories a tree
holds, using the same rules as the parser: a document directory holds a
schema.yaml, and a requirement is a heading followed by a `reqmd` fence.

### IMP-main-hasAttrFence-27ce782: hasAttrFence
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-run/main.go
x-reqmd.source-line: 350
x-reqmd.symbol-kind: function
trace: []
```

hasAttrFence reports whether the heading on line i opens a requirement.

### IMP-main-sampleRSS-c8fc27c: sampleRSS
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-run/main.go
x-reqmd.source-line: 361
x-reqmd.symbol-kind: function
trace: []
```

### IMP-main-readRSSKB-b69e42f: readRSSKB
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-run/main.go
x-reqmd.source-line: 377
x-reqmd.symbol-kind: function
trace: []
```

### IMP-main-collectHardware-587ae74: collectHardware
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: cmd/reqmd-bench-run/main.go
x-reqmd.source-line: 397
x-reqmd.symbol-kind: function
trace: []
```

collectHardware gathers CPU, OS, and runtime metadata for reproducibility.

## cmd/reqmd/main.go
### IMP-main-exitCoder-3b8cb2f: exitCoder
```yaml reqmd
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
```yaml reqmd
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

