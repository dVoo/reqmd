# Package: parser

<!-- reqmd: generated -->

## internal/parser/attrscan.go
### IMP-parser-errUnsupported-5dd0116: errUnsupported
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/attrscan.go
x-reqmd.source-line: 16
x-reqmd.symbol-kind: var
trace: []
```

errUnsupported is returned by the scanner for input outside the subset it
can parse *exactly*: nested maps, multi-line scalars, duplicate keys, and
scalars yaml.v3 would type differently (floats, timestamps, hex, big
integers). decodeAttrs then re-parses the whole block with yaml.v3, so the
result never depends on which path produced it.

### IMP-parser-decodeAttrs-639cc16: decodeAttrs
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/attrscan.go
x-reqmd.source-line: 21
x-reqmd.symbol-kind: function
trace: []
```

decodeAttrs parses the contents of an ```reqmd block into a map. It runs the
fast line scanner first and falls back to yaml.v3 for anything the scanner
cannot reproduce byte-for-byte (see errUnsupported).

### IMP-parser-decodeAttrYAML-b0e14ba: decodeAttrYAML
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/attrscan.go
x-reqmd.source-line: 63
x-reqmd.symbol-kind: function
trace: []
```

decodeAttrYAML is the fast path: it parses a small flat YAML map (the
contents of an ```reqmd block) into map[string]any without building the
intermediate YAML document AST that yaml.v3 allocates.

Supported syntax (covers all real-world reqmd blocks):
   - key: value          (scalar: string, int, bool)
   - key: "quoted value" (quoted string — quotes stripped)
   - key:                (list — following lines are "  - item")
   - item1
   - item2
   - # comment           (ignored)
   - blank lines          (ignored)

Values are typed as yaml.v3 would type them:
   - unquoted → string (unless it's an int or bool literal)
   - "true"/"false" → bool
   - integer → int
   - quoted → string (always, even if contents look like a number)
   - list items → []any of strings

Anything it cannot reproduce exactly — malformed input, nested maps, flow
mappings, duplicate keys, or a scalar yaml.v3 would retype (float,
timestamp, hex, big integer) — returns errUnsupported so decodeAttrs
re-parses the block with yaml.v3. Callers should use decodeAttrs, not this
function directly.

## internal/parser/attrscan\_test.go
### IMP-parser-TestDecodeAttrYAML\_Scalars-c181e5c: TestDecodeAttrYAML\_Scalars
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/attrscan_test.go
x-reqmd.source-line: 11
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-decodeAttrYAML-b0e14ba
```

### IMP-parser-TestDecodeAttrYAML\_QuotedStrings-552fc27: TestDecodeAttrYAML\_QuotedStrings
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/attrscan_test.go
x-reqmd.source-line: 31
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-decodeAttrYAML-b0e14ba
```

### IMP-parser-TestDecodeAttrYAML\_BlockList-691e48c: TestDecodeAttrYAML\_BlockList
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/attrscan_test.go
x-reqmd.source-line: 47
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-decodeAttrYAML-b0e14ba
```

### IMP-parser-TestDecodeAttrYAML\_FlowSeq-e8e0bbe: TestDecodeAttrYAML\_FlowSeq
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/attrscan_test.go
x-reqmd.source-line: 68
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-decodeAttrYAML-b0e14ba
```

### IMP-parser-TestDecodeAttrYAML\_EmptyValue-6e824d9: TestDecodeAttrYAML\_EmptyValue
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/attrscan_test.go
x-reqmd.source-line: 89
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-decodeAttrYAML-b0e14ba
```

### IMP-parser-TestDecodeAttrYAML\_Comments-b820f3d: TestDecodeAttrYAML\_Comments
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/attrscan_test.go
x-reqmd.source-line: 100
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-decodeAttrYAML-b0e14ba
```

### IMP-parser-TestDecodeAttrYAML\_BlankLines-e425642: TestDecodeAttrYAML\_BlankLines
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/attrscan_test.go
x-reqmd.source-line: 117
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-decodeAttrYAML-b0e14ba
```

### IMP-parser-TestDecodeAttrYAML\_EmptyKey-279d177: TestDecodeAttrYAML\_EmptyKey
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/attrscan_test.go
x-reqmd.source-line: 128
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-decodeAttrYAML-b0e14ba
```

### IMP-parser-TestDecodeAttrYAML\_BadYAML-92dd0bd: TestDecodeAttrYAML\_BadYAML
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/attrscan_test.go
x-reqmd.source-line: 135
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-decodeAttrYAML-b0e14ba
```

### IMP-parser-TestDecodeAttrYAML\_MixedTypes-54633eb: TestDecodeAttrYAML\_MixedTypes
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/attrscan_test.go
x-reqmd.source-line: 142
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-decodeAttrYAML-b0e14ba
```

### IMP-parser-TestDecodeAttrYAML\_ParityWithYAMLv3-fe54232: TestDecodeAttrYAML\_ParityWithYAMLv3
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/attrscan_test.go
x-reqmd.source-line: 163
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-decodeAttrYAML-b0e14ba
```

### IMP-parser-TestDecodeAttrs\_DuplicateKey-ac3785f: TestDecodeAttrs\_DuplicateKey
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/attrscan_test.go
x-reqmd.source-line: 201
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-decodeAttrs-639cc16
```

TestDecodeAttrs_DuplicateKey rejects a repeated key instead of silently
keeping the last value (yaml.v3 rejects it too, with a line number).

### IMP-parser-TestDecodeAttrYAML\_UnsupportedScalarIsReported-6da2475: TestDecodeAttrYAML\_UnsupportedScalarIsReported
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/attrscan_test.go
x-reqmd.source-line: 210
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-decodeAttrYAML-b0e14ba
```

TestDecodeAttrYAML_UnsupportedScalarIsReported guards the fast path: a
scalar it cannot type must be reported as errUnsupported so decodeAttrs
falls back, never silently stringified.

### IMP-parser-TestParseScalar-c3f7acf: TestParseScalar
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/attrscan_test.go
x-reqmd.source-line: 219
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseScalar-0fec6c3
```

TestParseScalar covers the plain-scalar typing the scanner must reproduce
exactly; anything it cannot type is rejected so the block re-parses with
yaml.v3 instead of silently becoming a string.

### IMP-parser-TestIsPlainStringStart-e9d2853: TestIsPlainStringStart
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/attrscan_test.go
x-reqmd.source-line: 259
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-isPlainStringStart-c2e4d67
```

### IMP-parser-TestIsIntegerString-0ff5297: TestIsIntegerString
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/attrscan_test.go
x-reqmd.source-line: 279
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-isIntegerString-264690b
```

### IMP-parser-TestUnquoteYAML-6bee43a: TestUnquoteYAML
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/attrscan_test.go
x-reqmd.source-line: 302
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-unquoteYAML-4ed7ece
```

TestUnquoteYAML covers both quote styles: doubled single quotes are a
literal quote, double quotes carry YAML escapes, and an unterminated quote
is an error (which send the block to yaml.v3).

### IMP-parser-TestUnescapeYAMLDouble-ba46c0a: TestUnescapeYAMLDouble
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/attrscan_test.go
x-reqmd.source-line: 333
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-unescapeYAMLDouble-9b443ee
```

### IMP-parser-TestLeadingSpaces-662b141: TestLeadingSpaces
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/attrscan_test.go
x-reqmd.source-line: 342
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-leadingSpaces-223b365
```

### IMP-parser-TestParseFlowSeq-ee2b166: TestParseFlowSeq
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/attrscan_test.go
x-reqmd.source-line: 357
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseFlowSeq-f6009f2
```

TestParseFlowSeq covers flow sequences, including quoted items with commas
(which a naive string split would break) and invalid input.

### IMP-parser-TestSplitFlowItems-f1aac68: TestSplitFlowItems
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/attrscan_test.go
x-reqmd.source-line: 389
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-splitFlowItems-277c16c
```

TestSplitFlowItems covers the quote- and bracket-aware comma split.

## internal/parser/attrvalue.go
### IMP-parser-parseScalar-0fec6c3: parseScalar
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/attrvalue.go
x-reqmd.source-line: 15
x-reqmd.symbol-kind: function
trace: []
```

parseScalar converts an unquoted YAML plain scalar to the Go type yaml.v3
would produce: true/false → bool, null/~ → nil, integer → int, anything
else → string. Scalars yaml.v3 would retype as something this function
cannot produce (float, timestamp, hex, big integer) report errUnsupported
so decodeAttrs re-parses the block with yaml.v3 instead of silently
turning them into strings.

### IMP-parser-isPlainStringStart-c2e4d67: isPlainStringStart
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/attrvalue.go
x-reqmd.source-line: 48
x-reqmd.symbol-kind: function
trace: []
```

isPlainStringStart reports whether c can begin a plain YAML scalar that
yaml.v3 resolves as a string.

### IMP-parser-isIntegerString-264690b: isIntegerString
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/attrvalue.go
x-reqmd.source-line: 54
x-reqmd.symbol-kind: function
trace: []
```

isIntegerString reports whether s is a plain integer literal (optional
sign followed by digits). Used as the gate before strconv.Atoi.

### IMP-parser-unquoteYAML-4ed7ece: unquoteYAML
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/attrvalue.go
x-reqmd.source-line: 74
x-reqmd.symbol-kind: function
trace: []
```

unquoteYAML strips single or double quotes from a YAML scalar.

### IMP-parser-unescapeYAMLDouble-9b443ee: unescapeYAMLDouble
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/attrvalue.go
x-reqmd.source-line: 94
x-reqmd.symbol-kind: function
trace: []
```

### IMP-parser-leadingSpaces-223b365: leadingSpaces
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/attrvalue.go
x-reqmd.source-line: 124
x-reqmd.symbol-kind: function
trace: []
```

### IMP-parser-parseFlowSeq-f6009f2: parseFlowSeq
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/attrvalue.go
x-reqmd.source-line: 135
x-reqmd.symbol-kind: function
trace: []
```

parseFlowSeq parses a YAML flow sequence [item1, item2, item3] into []any.
Items are parsed as scalars (same rules as parseScalar). Quoted items
are supported.

### IMP-parser-splitFlowItems-277c16c: splitFlowItems
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/attrvalue.go
x-reqmd.source-line: 169
x-reqmd.symbol-kind: function
trace: []
```

splitFlowItems splits a flow sequence body on commas, respecting quotes.

## internal/parser/bench\_test.go
### IMP-parser-benchAttr-94dfd1d: benchAttr
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/bench_test.go
x-reqmd.source-line: 14
x-reqmd.symbol-kind: const
trace: []
```

benchAttr is a representative minimal reqmd block: two scalars, one int, one
list with pins. Both benchmarks decode it so the scanner's advantage stays
measurable — it is ~15x faster than yaml.v3 (1784 vs 27608 ns/op, 14 vs 140
allocs) and that is why the fast path exists at all. If the gap ever closes,
delete the scanner and keep yaml.v3.

### IMP-parser-BenchmarkDecodeAttrs-26d49bf: BenchmarkDecodeAttrs
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/bench_test.go
x-reqmd.source-line: 16
x-reqmd.symbol-kind: function
trace: []
```

### IMP-parser-BenchmarkDecodeAttrsYAMLv3-a71c0c3: BenchmarkDecodeAttrsYAMLv3
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/bench_test.go
x-reqmd.source-line: 25
x-reqmd.symbol-kind: function
trace: []
```

### IMP-parser-BenchmarkParseMD-2a9f1cb: BenchmarkParseMD
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/bench_test.go
x-reqmd.source-line: 38
x-reqmd.symbol-kind: function
trace: []
```

BenchmarkParseMD measures the whole file→content-tree path, which is what
`reqmd check`/`ls` spend nearly all their time in.

### IMP-parser-syntheticDoc-ccb009f: syntheticDoc
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/bench_test.go
x-reqmd.source-line: 63
x-reqmd.symbol-kind: function
trace: []
```

syntheticDoc builds a document of roughly n bytes made of root-level
requirements with reqmd blocks — the shape that exercises both the
root-split pre-scan (above parseParallelThreshold) and the chunked
parallel parse.

## internal/parser/git.go
### IMP-parser-DiscoverAtTagWithSchemas-5d84f4e: DiscoverAtTagWithSchemas
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/git.go
x-reqmd.source-line: 19
x-reqmd.symbol-kind: function
trace: []
```

DiscoverAtTagWithSchemas extracts the repo at the given git tag into a temp
directory, runs the standard Discover pipeline, and returns the parsed
documents plus the raw schema.yaml content for each document directory.

### IMP-parser-discoverAtTag-a32da82: discoverAtTag
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/git.go
x-reqmd.source-line: 25
x-reqmd.symbol-kind: function
trace: []
```

discoverAtTag extracts the repo at the given git tag and returns both the
parsed documents and the raw schema maps per document path.

### IMP-parser-rewriteNodeSources-3e563ee: rewriteNodeSources
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/git.go
x-reqmd.source-line: 98
x-reqmd.symbol-kind: function
trace: []
```

rewriteNodeSources applies fn to the Source of every node in the tree.

### IMP-parser-relToTmpDir-abe620e: relToTmpDir
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/git.go
x-reqmd.source-line: 106
x-reqmd.symbol-kind: function
trace: []
```

relToTmpDir returns path relative to the temp extraction directory.

### IMP-parser-ListSubmodulesAtTag-f8e2142: ListSubmodulesAtTag
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/git.go
x-reqmd.source-line: 122
x-reqmd.symbol-kind: function
trace: []
```

ListSubmodulesAtTag returns a map of submodule path → pinned commit SHA
at the given git tag. Uses `git ls-tree -t <tag>` and filters for
submodule entries (mode 160000, type commit). Returns an empty map
if the tag has no submodules. Returns an error if the tag is invalid
or the git command fails.

### IMP-parser-parseLsTreeOutput-ac82723: parseLsTreeOutput
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/git.go
x-reqmd.source-line: 153
x-reqmd.symbol-kind: function
trace: []
```

parseLsTreeOutput parses the output of `git ls-tree -t <tag>` and
returns only submodule entries (mode 160000, type commit) as
path → SHA. Exposed as a separate function for unit testability.

git ls-tree -t output format (one entry per line):

	<mode> <type> <object>\t<file>

Example:

	160000 commit abc123def456...\tvendor/spec-a
	100644 blob   789ghi012jkl...\tREADME.md

## internal/parser/parser.go
### IMP-parser-attrFenceInfo-2495c08: attrFenceInfo
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser.go
x-reqmd.source-line: 27
x-reqmd.symbol-kind: const
trace: []
```

### IMP-parser-mdParser-069e9b6: mdParser
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser.go
x-reqmd.source-line: 44
x-reqmd.symbol-kind: var
trace: []
```

mdParser is the shared goldmark parser used by parseMD. It only parses: the
content tree is built by walking the AST, so no renderer is attached.

Structural extensions are the only ones registered here — tables and task
lists must become real block nodes or their markers would be flattened into
paragraph text — plus the YAML frontmatter parser, which stores the document's
metadata on the AST. Purely presentational extensions (emoji, syntax
highlighting, alerts, diagrams) live on the exporter's renderer instead: the
body is stored as Markdown, so they apply when that Markdown is rendered.

### IMP-parser-Discover-cc87d2f: Discover
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser.go
x-reqmd.source-line: 53
x-reqmd.symbol-kind: function
trace: []
```

Discover walks root recursively, locating schema.yaml files.
Every directory containing schema.yaml is a document directory.
Returns one Document per document directory.
Document directories are loaded in parallel for performance.

### IMP-parser-findDocDirs-d29d277: findDocDirs
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser.go
x-reqmd.source-line: 78
x-reqmd.symbol-kind: function
trace: []
```

findDocDirs walks root and collects all directories containing schema.yaml.

### IMP-parser-LoadSchema-cb5dcb7: LoadSchema
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser.go
x-reqmd.source-line: 100
x-reqmd.symbol-kind: function
trace: []
```

LoadSchema reads and parses the schema.yaml file from a document directory.
It returns the raw schema map before any x-reqmd extraction.

### IMP-parser-loadDocument-aad11cd: loadDocument
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser.go
x-reqmd.source-line: 112
x-reqmd.symbol-kind: function
trace: []
```

### IMP-parser-parseFiles-1d993d6: parseFiles
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser.go
x-reqmd.source-line: 153
x-reqmd.symbol-kind: function
trace: []
```

parseFiles parses all .md files in a document directory and returns the
concatenated content tree roots in mdFiles order (deterministic), the
document title (the first file's h1 heading, if any), and any per-file
errors. Files are parsed in parallel;
results are collected by file index so the returned order never depends
on goroutine completion order.

### IMP-parser-result-426f339: result
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser.go
x-reqmd.source-line: 158
x-reqmd.symbol-kind: type
trace: []
```

### IMP-parser-parseParallelThreshold-93b90ab: parseParallelThreshold
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser.go
x-reqmd.source-line: 218
x-reqmd.symbol-kind: const
trace: []
```

parseParallelThreshold is the minimum file size (in bytes) for which the
root-split pre-scan + intra-file chunk parallelism is worth it. Below
this size the file is parsed serially: the pre-scan is an extra full
line scan, and a small file parses in
well under a millisecond so chunking cannot help. Cross-file parallelism
(the worker pool in parseFiles) still applies.

### IMP-parser-parseMD-626e5e6: parseMD
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser.go
x-reqmd.source-line: 237
x-reqmd.symbol-kind: function
trace: []
```

parseMD parses a single .md file and returns the assembled content tree
roots, the file title (the first h1 heading, when present), and any YAML
frontmatter as a map.

Level-1 headings are document titles, not content: they are never nodes
and can never be requirements (only h2+ can). The first h1 in the file is
captured as the file title; the tree assembled from the remaining nodes
starts at the content heading level.

For large files with multiple root-level requirements (parent=nil), the
file is split at root requirement boundaries and chunks are parsed in
parallel. Each chunk yields a flat node stream in source order; the
streams are concatenated and assembled into the tree afterwards so
nesting across chunk boundaries (e.g. a container in chunk 0 with a
requirement in chunk 1) is preserved. The pre-scan to find split
points is a cheap line scan that replicates the stack logic without
building a goldmark AST.

### IMP-parser-chunkResult-928de6c: chunkResult
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser.go
x-reqmd.source-line: 279
x-reqmd.symbol-kind: type
trace: []
```

Compute chunk boundaries: which split indices start/end each chunk.
Chunk 0 additionally covers the file preamble [0, splits[0]) so
leading containers and prose (including the h1 title) are never
dropped.

### IMP-parser-parseMDChunk-5dc2f33: parseMDChunk
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser.go
x-reqmd.source-line: 336
x-reqmd.symbol-kind: function
trace: []
```

parseMDChunk parses a single chunk of markdown (possibly the whole file)
using goldmark and returns the flat node stream in source order, the file
title (first h1 heading), and any YAML frontmatter. Every heading except a
level-1 document title becomes a node; prose and other blocks attach to the
most recently opened node's Body. Nodes are NOT assembled into a tree here —
callers must assemble the concatenated streams so nesting across chunk
boundaries is preserved.

### IMP-parser-assembleTree-d304ba8: assembleTree
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser.go
x-reqmd.source-line: 507
x-reqmd.symbol-kind: function
trace:
  - SW-PAR-002
```

reqmd:trace SW-PAR-002
assembleTree builds the document content tree from a flat, ordered node
stream. A node's parent is the nearest preceding heading with a
shallower level; requirements additionally get ParentID pointing at the
nearest preceding requirement heading. Sections keep KindSection whether
they end up with children or not: folder-ness is derived from Children, so
adding a subsection never changes a node's kind.

### IMP-parser-popByLevel-b6b3780: popByLevel
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser.go
x-reqmd.source-line: 535
x-reqmd.symbol-kind: function
trace: []
```

popByLevel removes stack entries whose level is >= target, mirroring
the heading-nesting rule: a heading at level L closes everything at
level L or deeper.

### IMP-parser-reqStackEntry-c41a8b4: reqStackEntry
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser.go
x-reqmd.source-line: 544
x-reqmd.symbol-kind: type
trace: []
```

reqStackEntry tracks an open requirement at a heading level during the
root-split pre-scan.

### IMP-parser-popStackToLevel-0f8087e: popStackToLevel
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser.go
x-reqmd.source-line: 550
x-reqmd.symbol-kind: function
trace: []
```

popStackToLevel removes entries whose level is >= targetLevel.
Used by the root-split pre-scan to replicate requirement nesting.

### IMP-parser-findRootSplits-2059702: findRootSplits
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser.go
x-reqmd.source-line: 567
x-reqmd.symbol-kind: function
trace: []
```

findRootSplits scans src line-by-line to find byte offsets where root-level
requirements begin. A root requirement is one where the requirement stack
is empty when the requirement heading is encountered. The scan replicates
the same stack logic as the assembly step but without building a goldmark
AST.

Returns the byte offsets of root requirement starts (including the heading
line) and any YAML frontmatter parsed from the file header.

If only 0 or 1 roots are found, the caller parses the whole file serially.

### IMP-parser-frontmatterEnd-c0380eb: frontmatterEnd
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser.go
x-reqmd.source-line: 604
x-reqmd.symbol-kind: function
trace: []
```

frontmatterEnd returns the offset just past a leading YAML frontmatter block,
or 0 when the file does not start with one. ok is false when the block is
opened on the first line and never closed.

### IMP-parser-isDashLine-05eacd6: isDashLine
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser.go
x-reqmd.source-line: 623
x-reqmd.symbol-kind: function
trace: []
```

isDashLine reports whether line is a non-empty run of dashes, surrounding
spaces ignored. It mirrors the frontmatter extension's own separator rule so
the split scan and the parser agree on where a metadata block starts and ends
(the extension accepts "--", "----", and any longer run, not just "---").

### IMP-parser-nextLine-3ad066b: nextLine
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser.go
x-reqmd.source-line: 639
x-reqmd.symbol-kind: function
trace: []
```

nextLine returns the line at byte offset pos (without the trailing \n)
and whether there was a newline. The returned slice references src —
no allocation.

### IMP-parser-headingLevelBytes-c6b847f: headingLevelBytes
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser.go
x-reqmd.source-line: 650
x-reqmd.symbol-kind: function
trace: []
```

headingLevelBytes returns the heading level (1-6) and true if the line
is a markdown heading (e.g. "## Title"). Returns 0, false otherwise.
Works on []byte to avoid string allocation.

### IMP-parser-nextNonBlankIsAttrBytes-b179628: nextNonBlankIsAttrBytes
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser.go
x-reqmd.source-line: 666
x-reqmd.symbol-kind: function
trace: []
```

nextNonBlankIsAttrBytes checks if the next non-blank line after byte
offset pos is a ```reqmd fence. Scans src directly — no allocation.

### IMP-parser-isNextAttrBlock-2901c53: isNextAttrBlock
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser.go
x-reqmd.source-line: 685
x-reqmd.symbol-kind: function
trace: []
```

isNextAttrBlock returns true if n's next sibling is a ```reqmd fence.
Blank lines between heading and fence are consumed by goldmark and don't
create AST nodes, so they don't affect this check.

### IMP-parser-splitHeadingID-bf6d963: splitHeadingID
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser.go
x-reqmd.source-line: 700
x-reqmd.symbol-kind: function
trace: []
```

splitHeadingID splits a heading like "REQ-001: This is a title"
into ("REQ-001", "This is a title"). If there's no colon+space,
returns (heading, "").

### IMP-parser-collectInlineText-88fd5e8: collectInlineText
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser.go
x-reqmd.source-line: 709
x-reqmd.symbol-kind: function
trace: []
```

--- text extraction helpers ---

### IMP-parser-codeInfo-ca47cb5: codeInfo
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser.go
x-reqmd.source-line: 726
x-reqmd.symbol-kind: function
trace: []
```

codeInfo returns the fenced block's full info string ("go title=\"x\""), not
just its language token: the body keeps its Markdown, so anything after the
language must survive the round-trip.

### IMP-parser-appendBodyBlock-1bd723d: appendBodyBlock
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser.go
x-reqmd.source-line: 734
x-reqmd.symbol-kind: function
trace: []
```

appendBodyBlock appends a block to the open node's body, separated from any
preceding prose by a blank line. Without that separator a fenced code block or
a thematic break glues onto the paragraph above it, and the fence stops being
a fence once the body Markdown is rendered again.

### IMP-parser-renderFence-4f29349: renderFence
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser.go
x-reqmd.source-line: 746
x-reqmd.symbol-kind: function
trace: []
```

renderFence re-emits a code block as a fenced block, so the stored body keeps
the delimiters the exporter needs to render it again.

### IMP-parser-paragraphText-2c38d4d: paragraphText
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser.go
x-reqmd.source-line: 760
x-reqmd.symbol-kind: function
trace: []
```

### IMP-parser-renderBlock-1867761: renderBlock
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser.go
x-reqmd.source-line: 778
x-reqmd.symbol-kind: function
trace: []
```

renderBlock returns the source markdown of a block that has no dedicated
case: lists, tables, blockquotes, and raw HTML. The body of a node is
markdown, so the block's original source is the faithful representation —
concatenating the line segments instead would drop the list markers and the
line breaks between items, turning a list into one run-on paragraph.

### IMP-parser-ParseSingleFile-250a119: ParseSingleFile
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser.go
x-reqmd.source-line: 815
x-reqmd.symbol-kind: function
trace: []
```

ParseSingleFile is used by `reqmd check <file> --schema <schema>`.

### IMP-parser-generatedMarker-3ef3883: generatedMarker
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser.go
x-reqmd.source-line: 827
x-reqmd.symbol-kind: const
trace: []
```

generatedMarker is the sentinel the renderer writes into every file it owns.
Its pre-rename spelling is still recognized so trees generated before the
renderer moved to reqmd-import keep parsing cleanly.

### IMP-parser-legacyGeneratedMarker-640e6be: legacyGeneratedMarker
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser.go
x-reqmd.source-line: 827
x-reqmd.symbol-kind: const
trace: []
```

generatedMarker is the sentinel the renderer writes into every file it owns.
Its pre-rename spelling is still recognized so trees generated before the
renderer moved to reqmd-import keep parsing cleanly.

### IMP-parser-isGeneratedMarker-22d244b: isGeneratedMarker
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser.go
x-reqmd.source-line: 835
x-reqmd.symbol-kind: function
trace: []
```

isGeneratedMarker reports whether an HTML block is one of the renderer's
generated-content markers. Such a block is tool metadata: keeping it would
open a phantom headingless content node in every generated document.

## internal/parser/parser\_test.go
### IMP-parser-parseReqTree-07a3a65: parseReqTree
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 15
x-reqmd.symbol-kind: function
trace: []
```

parseReqTree is a test helper: parses src and returns the assembled
content tree roots.

### IMP-parser-reqsOf-4fb7484: reqsOf
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 25
x-reqmd.symbol-kind: function
trace: []
```

reqsOf returns the flat requirement list of an assembled tree.

### IMP-parser-TestParseMD\_Full-9d1bcca: TestParseMD\_Full
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 31
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_Full verifies a normal requirement with heading, reqmd block,
prose body, and *Rationale: line.

### IMP-parser-TestParseMD\_NoRationale-a066ea1: TestParseMD\_NoRationale
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 61
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_NoRationale verifies that a requirement without *Rationale:*
has an empty Rational field.

### IMP-parser-TestParseMD\_MultipleRequirements-af22eff: TestParseMD\_MultipleRequirements
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 74
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_MultipleRequirements verifies that multiple requirements
in one document are each parsed correctly.

### IMP-parser-TestParseMD\_HeadingWithoutAttrIsInfo-c482966: TestParseMD\_HeadingWithoutAttrIsInfo
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 97
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_HeadingWithoutAttrIsInfo verifies that a heading without an
reqmd block produces an info item instead of a requirement.

### IMP-parser-TestParseMD\_BadYAML-b6e68d1: TestParseMD\_BadYAML
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 120
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_BadYAML verifies that invalid YAML inside an ```reqmd``` block
returns an error.

### IMP-parser-TestParseMD\_NonAttrCodeBlock-81dc505: TestParseMD\_NonAttrCodeBlock
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 133
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_NonAttrCodeBlock verifies that a fenced code block with a
non-attr info string (e.g. ```json) inside a requirement body is preserved.

### IMP-parser-TestParseMD\_EmptyFile-1ed701a: TestParseMD\_EmptyFile
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 149
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_EmptyFile verifies that parsing an empty source returns
no nodes and no error.

### IMP-parser-TestParseMD\_ThematicBreak-858144f: TestParseMD\_ThematicBreak
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 161
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_ThematicBreak verifies that thematic breaks (---) inside
a requirement body are preserved and don't disrupt body accumulation.

### IMP-parser-TestParseMD\_ArrayAttr-3b2533f: TestParseMD\_ArrayAttr
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 175
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_ArrayAttr verifies that an attribute with a YAML array value
is parsed into []interface{} with the correct elements.

### IMP-parser-TestParseMD\_SubReqUnderParent-8cbf9e9: TestParseMD\_SubReqUnderParent
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 198
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_SubReqUnderParent verifies that a ### heading with a reqmd block
immediately following a ## heading creates a sub-requirement with ParentID set.

### IMP-parser-TestParseMD\_MultipleSubReqs-1879aa2: TestParseMD\_MultipleSubReqs
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 229
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_MultipleSubReqs verifies that multiple ### sub-requirements
under the same ## parent all get the correct ParentID.

### IMP-parser-TestParseMD\_SubReqDynamicLevel-0a560b0: TestParseMD\_SubReqDynamicLevel
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 248
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_SubReqDynamicLevel verifies that a ### heading with reqmd block
as the FIRST requirement is a top-level requirement, not an orphan child.

### IMP-parser-TestParseMD\_SubHeadingBecomesInfo-345b983: TestParseMD\_SubHeadingBecomesInfo
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 264
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_SubHeadingBecomesInfo verifies that a ### heading without an
reqmd block under a requirement becomes an info item child, not body text.

### IMP-parser-TestParseMD\_DynamicReqLevel-7f422c2: TestParseMD\_DynamicReqLevel
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 295
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_DynamicReqLevel verifies that requirements chain by heading
level when the content tree starts at level 2 (h1 is reserved for the
document title).

### IMP-parser-TestParseMD\_H1IsDocumentTitle-9cbe059: TestParseMD\_H1IsDocumentTitle
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 319
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_H1IsDocumentTitle verifies that a level-1 heading is treated
as the document title: an h1 with a reqmd block is rejected (requirements
must be level 2+), and a plain h1 is captured as the title and never
becomes a content node.

### IMP-parser-TestParseMD\_SecondH1Rejected-1c7406e: TestParseMD\_SecondH1Rejected
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 353
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_SecondH1Rejected verifies that a second level-1 heading is a
parse error. Before this rule the heading text was silently dropped: only
the first h1 became the title, and the second one's prose was re-attached
to a fresh headingless node.

### IMP-parser-TestParseSingleFile-541ac55: TestParseSingleFile
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 371
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-ParseSingleFile-250a119
```

TestParseSingleFile verifies ParseSingleFile reads a real file from disk
and produces the correct results. Uses t.TempDir() for the temp file.

### IMP-parser-TestParseMD\_Title-adb89a7: TestParseMD\_Title
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 410
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_Title verifies that a heading with "ID: Title" format
splits the ID and Title correctly.

### IMP-parser-TestParseMD\_NoTitle-57898c2: TestParseMD\_NoTitle
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 425
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_NoTitle verifies that a heading without ": " has empty Title.

### IMP-parser-TestParseMD\_TitleWithColonInTitle-a90e081: TestParseMD\_TitleWithColonInTitle
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 441
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_TitleWithColonInTitle verifies that only the first ": " splits
(IDs can't contain colons per the ID pattern, but titles may).

### IMP-parser-TestParseMD\_ContainerHoldsRequirements-c119338: TestParseMD\_ContainerHoldsRequirements
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 458
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_ContainerHoldsRequirements verifies that a container heading
(no reqmd block) becomes a KindSection node that holds the requirements
nested under it, which stay top-level (no requirement parent).

### IMP-parser-TestParseMD\_UserExample1-4fe01b9: TestParseMD\_UserExample1
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 492
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_UserExample1 reproduces the scenario the user described:
a container at the same level breaks the requirement chain; requirements
nest under containers and stay top-level.

### IMP-parser-TestParseMD\_DeepNesting-264a679: TestParseMD\_DeepNesting
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 536
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_DeepNesting verifies requirements at arbitrary depth
(3+ levels) chain correctly when no containers intervene.

### IMP-parser-TestParseMD\_ContainerDoesNotBreakDeeperChain-8ed979f: TestParseMD\_ContainerDoesNotBreakDeeperChain
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 556
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_ContainerDoesNotBreakDeeperChain verifies that a container
heading deeper than the current requirement does NOT break the chain:
requirements after it still link to the nearest shallower requirement.

### IMP-parser-TestParseMD\_LevelGapNoContainer-efe3d45: TestParseMD\_LevelGapNoContainer
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 602
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_LevelGapNoContainer verifies that a heading gap without a
container still links children to the nearest shallower requirement.

### IMP-parser-TestParseMD\_SiblingContainersBreakChain-c8e8243: TestParseMD\_SiblingContainersBreakChain
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 618
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_SiblingContainersBreakChain verifies that multiple containers
at successive levels produce only top-level requirements.

### IMP-parser-TestParseMD\_Preamble-342cd93: TestParseMD\_Preamble
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 635
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_Preamble verifies that leading prose before the first heading
is captured as an info node and not dropped, and that an h1 title does not
become a wrapping container.

### IMP-parser-TestParseMD\_RequirementlessFile-04022cd: TestParseMD\_RequirementlessFile
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 672
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_RequirementlessFile verifies that a file with no requirement
headings at all is captured entirely as info nodes, with the h1 title kept
out of the content tree.

### IMP-parser-TestParseMD\_ChunkedPreamble-5ad1399: TestParseMD\_ChunkedPreamble
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 699
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_ChunkedPreamble verifies that the parallel chunked path
(multiple root requirements) still captures the file preamble and title.

### IMP-parser-TestParseMD\_LevelGapUnderContainer-8b98274: TestParseMD\_LevelGapUnderContainer
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 726
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_LevelGapUnderContainer verifies that a requirement inside a
container under another requirement gets the right ParentID (nearest
shallower requirement) and container nesting.

### IMP-parser-TestParseLsTreeOutput\_Empty-77f5401: TestParseLsTreeOutput\_Empty
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 740
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseLsTreeOutput-ac82723
```

### IMP-parser-TestParseLsTreeOutput\_NoSubmodules-4532e94: TestParseLsTreeOutput\_NoSubmodules
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 747
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseLsTreeOutput-ac82723
```

### IMP-parser-TestParseLsTreeOutput\_OneSubmodule-b8eea82: TestParseLsTreeOutput\_OneSubmodule
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 756
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseLsTreeOutput-ac82723
```

### IMP-parser-TestParseLsTreeOutput\_MultipleSubmodules-1076783: TestParseLsTreeOutput\_MultipleSubmodules
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 769
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseLsTreeOutput-ac82723
```

### IMP-parser-TestParseLsTreeOutput\_Malformed-52b1abd: TestParseLsTreeOutput\_Malformed
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 789
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseLsTreeOutput-ac82723
```

### IMP-parser-bigDoc-6733460: bigDoc
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 810
x-reqmd.symbol-kind: function
trace: []
```

bigDoc builds a document large enough to take the chunked parse path: a
title, preamble prose, a root container holding requirements, then many root
requirements. Its size matters — parseMD only splits above
parseParallelThreshold.

### IMP-parser-nodeSummary-15b10d4: nodeSummary
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 833
x-reqmd.symbol-kind: function
trace: []
```

nodeSummary renders a tree into a comparable string so the chunked result can
be compared with the serial one without walking structs field by field.

### IMP-parser-TestParseMD\_ChunkedMatchesSerial-40309b2: TestParseMD\_ChunkedMatchesSerial
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 845
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_ChunkedMatchesSerial is the invariant behind the parallel path:
batching roots into chunks must produce exactly the tree the single-pass
parser produces, including the preamble (chunk 0 covers [0, first root)).

### IMP-parser-TestParseMD\_ChunkedKeepsPreambleAndBoundaries-20ab508: TestParseMD\_ChunkedKeepsPreambleAndBoundaries
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 870
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_ChunkedKeepsPreambleAndBoundaries checks the two properties the
chunking exists for: leading prose and the container it opens must not be
dropped, and no requirement may be lost or duplicated at a chunk boundary.

### IMP-parser-TestParseMD\_WrappedParagraphKeepsLineBreaks-9fdd92a: TestParseMD\_WrappedParagraphKeepsLineBreaks
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 915
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_WrappedParagraphKeepsLineBreaks pins that a source-wrapped
paragraph keeps its line breaks in the stored body. Dropping them glued the
last word of one line onto the first word of the next in the HTML export.

### IMP-parser-TestParseMD\_ListBodyPreserved-976e285: TestParseMD\_ListBodyPreserved
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 936
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_ListBodyPreserved pins the body-capture contract for blocks that
have no dedicated case: a list in a requirement body must keep its markers and
its line breaks, and must be separated from the prose above it. Concatenating
the AST line segments dropped both, which turned every bullet list into one
run-on paragraph in the HTML export.

### IMP-parser-TestParseMD\_TableBodyPreserved-bad15b3: TestParseMD\_TableBodyPreserved
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 955
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_TableBodyPreserved covers the other block that used to be
flattened: a GFM table must survive as markdown, rows intact.

### IMP-parser-TestParseMD\_MermaidFencePreserved-73e98f9: TestParseMD\_MermaidFencePreserved
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 974
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_MermaidFencePreserved verifies that a mermaid block survives the
parser with its delimiters intact. Without them the exporter cannot turn it
into a diagram and renders the diagram source as prose.

### IMP-parser-TestParseMD\_FrontmatterParsed-5dd7032: TestParseMD\_FrontmatterParsed
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 991
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_FrontmatterParsed verifies that a leading YAML frontmatter block
is parsed into the document's metadata and contributes no content: no node,
no body prose.

### IMP-parser-TestParseFiles\_FrontmatterFirstFileWins-dc22cae: TestParseFiles\_FrontmatterFirstFileWins
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 1018
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseFiles-1d993d6
```

TestParseFiles_FrontmatterFirstFileWins pins the merge rule across the files
of one document directory.

### IMP-parser-TestParseMD\_ChunkedFrontmatter-46e1b7d: TestParseMD\_ChunkedFrontmatter
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 1046
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_ChunkedFrontmatter covers the same rule on the parallel path:
chunk 0 carries the file preamble, so it is the chunk that yields the
metadata, and the block must not reach any node's body.

### IMP-parser-TestParseMD\_FrontmatterUnterminatedIsError-c4c6945: TestParseMD\_FrontmatterUnterminatedIsError
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 1072
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_FrontmatterUnterminatedIsError pins the one failure shape where
content would otherwise vanish without a diagnostic: a file whose first line
opens a metadata block that is never closed. The extension accepts any run of
dashes as a separator, so the guard has to as well.

### IMP-parser-TestParseMD\_IndentedCodeBlockBecomesFence-d343855: TestParseMD\_IndentedCodeBlockBecomesFence
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 1093
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_IndentedCodeBlockBecomesFence verifies the v2 code-block merge:
an indented block is re-emitted as a fence so the body keeps code semantics
instead of losing its indentation markers.

### IMP-parser-TestParseMD\_FenceInfoPreserved-a762138: TestParseMD\_FenceInfoPreserved
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 1108
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_FenceInfoPreserved verifies that everything after the language in
a fence's info string survives the body round-trip, not just the language.

### IMP-parser-TestParseMD\_ChunkedFrontmatterSeparatorRule-db428aa: TestParseMD\_ChunkedFrontmatterSeparatorRule
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/parser/parser_test.go
x-reqmd.source-line: 1125
x-reqmd.symbol-kind: function
trace:
  - IMP-parser-parseMD-626e5e6
```

TestParseMD_ChunkedFrontmatterSeparatorRule keeps the split scan and the
extension in agreement: the extension opens a metadata block on any run of
dashes, so the scan must skip the same block. A heading-looking line inside
the block would otherwise be taken for the document title.

