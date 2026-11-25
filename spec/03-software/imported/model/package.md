# Package: model

<!-- reqmd: generated -->

## internal/model/builtin.go
### IMP-model-AttrDisposition-acdff3a: AttrDisposition
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/builtin.go
x-reqmd.source-line: 11
x-reqmd.symbol-kind: const
trace: []
```

Built-in attribute keys used in requirement attr blocks. They are
injected into every schema by internal/schema.Compile and accessed by
graph checks and HTML export.

### IMP-model-AttrDispositionReason-f8355ee: AttrDispositionReason
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/builtin.go
x-reqmd.source-line: 11
x-reqmd.symbol-kind: const
trace: []
```

Built-in attribute keys used in requirement attr blocks. They are
injected into every schema by internal/schema.Compile and accessed by
graph checks and HTML export.

### IMP-model-AttrRequiresTraceFrom-261c85c: AttrRequiresTraceFrom
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/builtin.go
x-reqmd.source-line: 11
x-reqmd.symbol-kind: const
trace: []
```

Built-in attribute keys used in requirement attr blocks. They are
injected into every schema by internal/schema.Compile and accessed by
graph checks and HTML export.

### IMP-model-AttrStatus-0033655: AttrStatus
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/builtin.go
x-reqmd.source-line: 11
x-reqmd.symbol-kind: const
trace: []
```

Built-in attribute keys used in requirement attr blocks. They are
injected into every schema by internal/schema.Compile and accessed by
graph checks and HTML export.

### IMP-model-AttrTrace-e4c0e28: AttrTrace
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/builtin.go
x-reqmd.source-line: 11
x-reqmd.symbol-kind: const
trace: []
```

Built-in attribute keys used in requirement attr blocks. They are
injected into every schema by internal/schema.Compile and accessed by
graph checks and HTML export.

### IMP-model-AttrVerify-394826e: AttrVerify
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/builtin.go
x-reqmd.source-line: 11
x-reqmd.symbol-kind: const
trace: []
```

Built-in attribute keys used in requirement attr blocks. They are
injected into every schema by internal/schema.Compile and accessed by
graph checks and HTML export.

### IMP-model-AttrVersion-253d1dd: AttrVersion
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/builtin.go
x-reqmd.source-line: 11
x-reqmd.symbol-kind: const
trace: []
```

Built-in attribute keys used in requirement attr blocks. They are
injected into every schema by internal/schema.Compile and accessed by
graph checks and HTML export.

### IMP-model-AttrResultEvidence-cca1fcb: AttrResultEvidence
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/builtin.go
x-reqmd.source-line: 25
x-reqmd.symbol-kind: const
trace: []
```

Tool-owned attribute keys used on synthesized verification-result
pseudo-requirements (internal/verify). They are namespaced under
x-reqmd.* — the same convention as reqmd-code's provenance attributes —
so they can never collide with authored schema attributes.

### IMP-model-AttrResultOutcome-5a2e39e: AttrResultOutcome
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/builtin.go
x-reqmd.source-line: 25
x-reqmd.symbol-kind: const
trace: []
```

Tool-owned attribute keys used on synthesized verification-result
pseudo-requirements (internal/verify). They are namespaced under
x-reqmd.* — the same convention as reqmd-code's provenance attributes —
so they can never collide with authored schema attributes.

### IMP-model-AttrResultVerifiedAt-30b0e27: AttrResultVerifiedAt
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/builtin.go
x-reqmd.source-line: 25
x-reqmd.symbol-kind: const
trace: []
```

Tool-owned attribute keys used on synthesized verification-result
pseudo-requirements (internal/verify). They are namespaced under
x-reqmd.* — the same convention as reqmd-code's provenance attributes —
so they can never collide with authored schema attributes.

### IMP-model-AttrResultVerifier-5a573d0: AttrResultVerifier
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/builtin.go
x-reqmd.source-line: 25
x-reqmd.symbol-kind: const
trace: []
```

Tool-owned attribute keys used on synthesized verification-result
pseudo-requirements (internal/verify). They are namespaced under
x-reqmd.* — the same convention as reqmd-code's provenance attributes —
so they can never collide with authored schema attributes.

### IMP-model-StripPin-f9a274d: StripPin
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/builtin.go
x-reqmd.source-line: 42
x-reqmd.symbol-kind: function
trace: []
```

StripPin removes a trailing `~N` version pin from a requirement ref.
Returns the bare ref, the pin value, and true if a numeric pin was
present. A `~` followed by a non-numeric suffix is NOT treated as a
pin (e.g. "UP-001~abc" returns ("UP-001~abc", 0, false)) — this
matches the graph layer's parsing and prevents mis-stripping refs
that contain `~` for other reasons.

This is the single source of truth for `~N` stripping across the
codebase: graph (Pass 2 pin extraction), repin (newRefFrom, stripPin
for display), and verify (MergeLatest key normalization).

### IMP-model-StatusApproved-be78e3d: StatusApproved
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/builtin.go
x-reqmd.source-line: 58
x-reqmd.symbol-kind: const
trace: []
```

Built-in status enum values. The default (when no status attr is
declared) is StatusApproved — see StatusDefault.

### IMP-model-StatusDraft-a123da1: StatusDraft
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/builtin.go
x-reqmd.source-line: 58
x-reqmd.symbol-kind: const
trace: []
```

Built-in status enum values. The default (when no status attr is
declared) is StatusApproved — see StatusDefault.

### IMP-model-StatusDefault-2219a8f: StatusDefault
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/builtin.go
x-reqmd.source-line: 66
x-reqmd.symbol-kind: const
trace: []
```

StatusDefault is the implicit status when none is declared on a
requirement. The built-in `status` attribute is enum [draft, approved];
`approved` is the coverage-providing value.

## internal/model/builtin_test.go
### IMP-model-TestStripPin-4488901: TestStripPin
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/builtin_test.go
x-reqmd.source-line: 5
x-reqmd.symbol-kind: function
trace:
  - IMP-model-StripPin-f9a274d
```

## internal/model/format.go
### IMP-model-FormatScalar-41a318f: FormatScalar
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/format.go
x-reqmd.source-line: 16
x-reqmd.symbol-kind: function
trace: []
```

FormatScalar converts a primitive scalar value to its string
representation. It handles the types that YAML decoding produces for
attribute values: string, bool, int, int64, float64. Non-scalar types
(slices, maps) return "" — callers should handle those with
encoding/json or a domain-specific renderer.

This is the shared scalar formatter used by the exporter and reporter
packages to avoid duplicating the same type switch in four places.

### IMP-model-AttrString-ac410c6: AttrString
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/format.go
x-reqmd.source-line: 41
x-reqmd.symbol-kind: function
trace: []
```

Attribute accessors. Attr values come from the YAML attr block and are
typed by the parser (string, bool, int, []any of those). These helpers
are the single place that type-switches on them; every consumer (graph
checks, exporters, reporters) reads attrs through them instead of
reimplementing the same switch.
AttrString returns the string value of attrs[key], or "" when the key
is absent or holds a non-string value.

### IMP-model-AttrStrings-b28638f: AttrStrings
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/format.go
x-reqmd.source-line: 50
x-reqmd.symbol-kind: function
trace: []
```

AttrStrings reads an array attr as []string. Returns nil when the key is
absent (so callers can tell "not set" from "explicitly empty") and
[]string{} when the key is present but the array is empty. Non-string
items are dropped.

### IMP-model-AttrList-6a7d22b: AttrList
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/format.go
x-reqmd.source-line: 76
x-reqmd.symbol-kind: function
trace: []
```

AttrList returns an array attr in its native []any form, converting a
[]string. Returns nil when the key is absent or not an array.

### IMP-model-AttrInt-1b173ec: AttrInt
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/format.go
x-reqmd.source-line: 93
x-reqmd.symbol-kind: function
trace: []
```

AttrInt reads an integer attr, handling the types that appear in
practice: int and int64 (YAML decode) and float64 (JSON round-trip).
Returns 0 when the key is absent or undecodable.

### IMP-model-FormatAttr-04da9dc: FormatAttr
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/format.go
x-reqmd.source-line: 108
x-reqmd.symbol-kind: function
trace: []
```

FormatAttr renders an attribute value for display: scalars through
FormatScalar, slices and maps as compact JSON. Unknown types render as
"" so a table cell never shows Go syntax.

## internal/model/node.go
### IMP-model-Kind-382a382: Kind
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/node.go
x-reqmd.source-line: 9
x-reqmd.symbol-kind: type
trace: []
```

Kind discriminates the node types in a document's content tree.

#### IMP-model-String-0301365: String
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/node.go
x-reqmd.source-line: 25
x-reqmd.symbol-kind: method
trace: []
```

String returns the lowercase display name used in text and JSON output.

### IMP-model-KindRequirement-ace5645: KindRequirement
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/node.go
x-reqmd.source-line: 11
x-reqmd.symbol-kind: const
trace: []
```

### IMP-model-KindSection-031c53d: KindSection
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/node.go
x-reqmd.source-line: 11
x-reqmd.symbol-kind: const
trace: []
```

### IMP-model-Node-aaeb5ec: Node
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/node.go
x-reqmd.source-line: 40
x-reqmd.symbol-kind: type
trace: []
```

Node is a single content node in a parsed Markdown document. Every
heading becomes a node; a node is a requirement iff it has an adjacent
attr block. Prose and other blocks attach to the nearest open heading's
Body. Nodes form a tree via Children, preserving document order.

#### IMP-model-IsRequirement-4ca5fff: IsRequirement
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/node.go
x-reqmd.source-line: 56
x-reqmd.symbol-kind: method
trace: []
```

IsRequirement reports whether the node is a requirement (as opposed to
a section).

### IMP-model-XReqmd-25ba149: XReqmd
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/node.go
x-reqmd.source-line: 62
x-reqmd.symbol-kind: type
trace: []
```

XReqmd holds reqmd-specific directory metadata from schema.yaml.
All fields map directly to the x-reqmd extension block.

### IMP-model-TraceUpstream-c4a21d4: TraceUpstream
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/node.go
x-reqmd.source-line: 78
x-reqmd.symbol-kind: type
trace: []
```

TraceUpstream declares the expected upstream layer and source directories.

### IMP-model-SourceConfig-1cf24c7: SourceConfig
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/node.go
x-reqmd.source-line: 84
x-reqmd.symbol-kind: type
trace: []
```

SourceConfig describes where to find and how to parse the originating artefact.

### IMP-model-CollectSectionPaths-2a19657: CollectSectionPaths
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/node.go
x-reqmd.source-line: 94
x-reqmd.symbol-kind: function
trace: []
```

CollectSectionPaths returns, for every requirement in the tree, the titles of
its ancestor sections joined by " / " (empty when it has none). It is the one
implementation of "where does this requirement sit in the document?" shared by
the NDJSON export (the section_path column) and the baseline diff (a parent
delta), so the two can never disagree.

### IMP-model-sectionPathSeparator-8ff6927: sectionPathSeparator
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/node.go
x-reqmd.source-line: 116
x-reqmd.symbol-kind: const
trace: []
```

sectionPathSeparator joins ancestor section titles in a section path.

### IMP-model-Document-57ebc69: Document
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/node.go
x-reqmd.source-line: 128
x-reqmd.symbol-kind: type
trace: []
```

Document groups all content sharing a single schema.yaml. Nodes holds
the ordered content tree; every heading in every .md file is a node
except the level-1 document title (see parser). Title is the first
file's h1 heading, when present.

Synthetic marks documents that do not originate from authored Markdown
on disk — currently the pseudo-document of verification results
synthesized by internal/verify, later also synthesized test cases. Nodes
of a synthetic document carry no schema, no ID prefix, and no source file
that tooling may rewrite (e.g. repin must skip them).

#### IMP-model-Requirements-83ad233: Requirements
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/node.go
x-reqmd.source-line: 142
x-reqmd.symbol-kind: method
trace: []
```

Requirements returns all requirement-kind nodes in document order
(depth-first). The returned pointers reference the tree itself, so
in-place mutation updates the document.

### IMP-model-CollectRequirements-a8770c2: CollectRequirements
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/node.go
x-reqmd.source-line: 148
x-reqmd.symbol-kind: function
trace: []
```

CollectRequirements walks a node tree depth-first and returns all
requirement-kind nodes in document order.

## internal/model/node_test.go
### IMP-model-TestNodeAllFields-4f66d3a: TestNodeAllFields
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/node_test.go
x-reqmd.source-line: 5
x-reqmd.symbol-kind: function
trace:
  - IMP-model-Node-aaeb5ec
```

### IMP-model-TestNodeEdgeCases-7f138f6: TestNodeEdgeCases
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/node_test.go
x-reqmd.source-line: 48
x-reqmd.symbol-kind: function
trace:
  - IMP-model-Node-aaeb5ec
```

### IMP-model-TestKindString-4623907: TestKindString
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/node_test.go
x-reqmd.source-line: 73
x-reqmd.symbol-kind: function
trace:
  - IMP-model-Kind-382a382
```

### IMP-model-TestCollectRequirements-e08b61e: TestCollectRequirements
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/node_test.go
x-reqmd.source-line: 88
x-reqmd.symbol-kind: function
trace:
  - IMP-model-CollectRequirements-a8770c2
```

### IMP-model-TestDocumentRequirements-0d84ca1: TestDocumentRequirements
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/node_test.go
x-reqmd.source-line: 108
x-reqmd.symbol-kind: function
trace:
  - IMP-model-Document-57ebc69
```

### IMP-model-TestDocumentZeroNodes-71758c0: TestDocumentZeroNodes
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/model/node_test.go
x-reqmd.source-line: 134
x-reqmd.symbol-kind: function
trace:
  - IMP-model-Document-57ebc69
```

## reqmd-code/internal/model/short_hash.go
### IMP-model-ShortHash-5aa3366: ShortHash
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/model/short_hash.go
x-reqmd.source-line: 21
x-reqmd.symbol-kind: function
trace: []
```

ShortHash produces a stable 7-hex-char FNV-1a-64 hash over (pkg, name),
masked to 28 bits. It exists as a single shared utility so the cobra
`extract` command and any future entry point produce byte-identical
IDs for the same source. Two callers previously duplicated this
function; keeping it here removes the maintenance hazard.

The FNV-1a-64 offset basis and prime are the standard constants
(0xcbf29ce484222325 and 0x100000001b3, respectively). The 28-bit
mask is "7 hex chars" — long enough to make accidental collisions
negligible for any realistic source tree, short enough to keep IDs
readable in generated .md files.

A single XOR of the ':' separator is mixed in between the pkg and
name halves: adding the separator as a normal hashed byte would
self-cancel under XOR, so the explicit XOR is what disambiguates
("x:y") from ("x::y") and similar.

## reqmd-code/internal/model/symbol.go
### IMP-model-SymbolKind-9dece46: SymbolKind
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/model/symbol.go
x-reqmd.source-line: 8
x-reqmd.symbol-kind: type
trace: []
```

SymbolKind enumerates the top-level declarations a language plugin can
extract. Each language plugin maps its AST node kinds to one of these
values; the writer treats them uniformly.

### IMP-model-SymbolConst-0d7f4e7: SymbolConst
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/model/symbol.go
x-reqmd.source-line: 10
x-reqmd.symbol-kind: const
trace: []
```

### IMP-model-SymbolFunction-45ca65c: SymbolFunction
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/model/symbol.go
x-reqmd.source-line: 10
x-reqmd.symbol-kind: const
trace: []
```

### IMP-model-SymbolMethod-866e67d: SymbolMethod
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/model/symbol.go
x-reqmd.source-line: 10
x-reqmd.symbol-kind: const
trace: []
```

### IMP-model-SymbolType-a1c089c: SymbolType
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/model/symbol.go
x-reqmd.source-line: 10
x-reqmd.symbol-kind: const
trace: []
```

### IMP-model-SymbolVar-0a710d5: SymbolVar
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/model/symbol.go
x-reqmd.source-line: 10
x-reqmd.symbol-kind: const
trace: []
```

### IMP-model-Symbol-ef16294: Symbol
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/model/symbol.go
x-reqmd.source-line: 32
x-reqmd.symbol-kind: type
trace: []
```

Symbol describes one top-level declaration extracted from a source file.
It is produced by language plugins and consumed by the extractor's
normalizer/tracer, which turn it into a GeneratedReq for the writer.

Coordinates are 1-based for human readability (matching what editors show).

### IMP-model-StripReceiverName-9a492f8: StripReceiverName
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/model/symbol.go
x-reqmd.source-line: 70
x-reqmd.symbol-kind: function
trace: []
```

StripReceiverName normalizes a receiver string to its parent type name:
strips an optional leading '*' (pointer receiver) and any surrounding
brackets the AST may emit. Used to look up a method's parent type in
the per-package ID map. Uses TrimPrefix/TrimSuffix, not TrimLeft with a
cutset, so interior characters are not stripped.

