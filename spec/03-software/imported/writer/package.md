# Package: writer

<!-- reqmd: generated -->

## reqmd-import/internal/writer/writer.go
### IMP-writer-GeneratedMarker-c980113: GeneratedMarker
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/writer/writer.go
x-reqmd.source-line: 22
x-reqmd.symbol-kind: const
trace: []
```

GeneratedMarker is the HTML-comment sentinel embedded in the header of
every generated file. The idempotent write path uses it to distinguish
generated content from hand-authored content in the same tree.

### IMP-writer-legacyMarker-4a7dcdd: legacyMarker
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/writer/writer.go
x-reqmd.source-line: 28
x-reqmd.symbol-kind: const
trace: []
```

legacyMarker is the pre-rename marker emitted by reqmd-code before the
renderer moved to reqmd-import. It is still recognized so a re-run
overwrites an existing generated tree instead of mistaking it for
hand-authored content.

### IMP-writer-KindRequirement-b12f9b5: KindRequirement
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/writer/writer.go
x-reqmd.source-line: 32
x-reqmd.symbol-kind: const
trace: []
```

Kind discriminates a rendered item. A requirement gets a reqmd block; a
section never does, which is the only difference reqmd cares about.

### IMP-writer-KindSection-c5fd98d: KindSection
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/writer/writer.go
x-reqmd.source-line: 32
x-reqmd.symbol-kind: const
trace: []
```

Kind discriminates a rendered item. A requirement gets a reqmd block; a
section never does, which is the only difference reqmd cares about.

### IMP-writer-Item-c445e3f: Item
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/writer/writer.go
x-reqmd.source-line: 40
x-reqmd.symbol-kind: type
trace: []
```

Item is one resolved content item to render: its heading depth, its final
reqmd ID (requirements only), the trace refs, and the remaining reqmd-block
entries.

#### IMP-writer-IsRequirement-440dc92: IsRequirement
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/writer/writer.go
x-reqmd.source-line: 54
x-reqmd.symbol-kind: method
trace: []
```

IsRequirement reports whether the item renders a reqmd block. Only an
explicit KindSection makes a section: the zero value is a requirement, so an
item built without a kind can never silently lose its reqmd block.

### IMP-writer-File-24cd9e6: File
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/writer/writer.go
x-reqmd.source-line: 59
x-reqmd.symbol-kind: type
trace: []
```

File is one rendered output file, relative to the output root.

### IMP-writer-IsGenerated-a943579: IsGenerated
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/writer/writer.go
x-reqmd.source-line: 65
x-reqmd.symbol-kind: function
trace: []
```

IsGenerated reports whether content carries the generated marker.

### IMP-writer-Write-44afbbf: Write
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/writer/writer.go
x-reqmd.source-line: 73
x-reqmd.symbol-kind: function
trace: []
```

Write writes every file under outDir, creating directories as needed. A
file is only rewritten when its content changed; existing content without
the generated marker is left untouched so hand-authored files are never
clobbered.

### IMP-writer-writeIfChanged-a5c5088: writeIfChanged
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/writer/writer.go
x-reqmd.source-line: 82
x-reqmd.symbol-kind: function
trace: []
```

### IMP-writer-RenderDocument-9c37809: RenderDocument
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/writer/writer.go
x-reqmd.source-line: 106
x-reqmd.symbol-kind: function
trace:
  - SW-IMP-005
```

reqmd:trace SW-IMP-005
RenderDocument renders a document: title, generated marker, optional
preamble prose, then the items in the order given. Depth is carried per item,
so a section's children must follow it in the slice.

### IMP-writer-headingEscaper-3d460e9: headingEscaper
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/writer/writer.go
x-reqmd.source-line: 136
x-reqmd.symbol-kind: var
trace: []
```

RenderItem renders one item: heading, then (for a requirement) the attr
block, then the body. Depth is the nesting level: 0 renders "##", 1 renders
"###", and so on (reqmd accepts any heading level from h2 down). An item
without a title renders as a plain prose block, which is how the headingless
document preamble and untitled sections are represented.
headingEscaper escapes the characters that would start inline Markdown
inside a generated heading. A symbol or package named `__init__` would
otherwise parse as emphasis and vanish from the ID (`IMP-init-…`), breaking
the document's id-prefix. A backslash before an ASCII punctuation character
is a CommonMark escape, so the parsed heading text is identical to the
authored one.

### IMP-writer-escapeHeading-6d2f897: escapeHeading
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/writer/writer.go
x-reqmd.source-line: 150
x-reqmd.symbol-kind: function
trace: []
```

### IMP-writer-RenderItem-136994f: RenderItem
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/writer/writer.go
x-reqmd.source-line: 152
x-reqmd.symbol-kind: function
trace: []
```

### IMP-writer-writeAttr-19d41f8: writeAttr
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/writer/writer.go
x-reqmd.source-line: 185
x-reqmd.symbol-kind: function
trace: []
```

writeAttr writes the reqmd-block body: status first, then the remaining
attributes in sorted key order (deterministic output), then the trace list.

### IMP-writer-writeAttrValue-844e649: writeAttrValue
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/writer/writer.go
x-reqmd.source-line: 203
x-reqmd.symbol-kind: function
trace: []
```

### IMP-writer-writeList-54a61a7: writeList
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/writer/writer.go
x-reqmd.source-line: 232
x-reqmd.symbol-kind: function
trace: []
```

### IMP-writer-scalar-57c235c: scalar
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/writer/writer.go
x-reqmd.source-line: 248
x-reqmd.symbol-kind: function
trace: []
```

### IMP-writer-writeRefList-c76c778: writeRefList
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/writer/writer.go
x-reqmd.source-line: 264
x-reqmd.symbol-kind: function
trace: []
```

writeRefList writes a YAML block list, or `key: []` when empty.

### IMP-writer-quoteIfNeeded-5d24016: quoteIfNeeded
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/writer/writer.go
x-reqmd.source-line: 282
x-reqmd.symbol-kind: function
trace: []
```

quoteIfNeeded formats a value as a YAML scalar that reqmd reads back
unchanged. A plain scalar is unsafe when it contains YAML punctuation, when
it has leading or trailing whitespace, or when reqmd's attr scanner
(internal/parser/attrscan.go, parseScalar) would retype it as null, a bool,
or an integer — so all of those are double-quoted with the escapes that
scanner understands. Over-quoting is harmless; under-quoting silently
changes the value's type.

### IMP-writer-plainScalarSafe-aada755: plainScalarSafe
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/writer/writer.go
x-reqmd.source-line: 302
x-reqmd.symbol-kind: function
trace: []
```

plainScalarSafe reports whether s can be written as a plain (unquoted) YAML
scalar without changing its meaning for reqmd's attr scanner. The retype
list must stay a superset of that scanner's parseScalar: if the scanner
starts resolving something new to a non-string, a value that this function
calls safe becomes corrupt.

## reqmd-import/internal/writer/writer\_test.go
### IMP-writer-firstLine-5d9de94: firstLine
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/writer/writer_test.go
x-reqmd.source-line: 15
x-reqmd.symbol-kind: function
trace: []
```

### IMP-writer-TestRenderDocument-13d0035: TestRenderDocument
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/writer/writer_test.go
x-reqmd.source-line: 20
x-reqmd.symbol-kind: function
trace:
  - IMP-writer-RenderDocument-9c37809
```

### IMP-writer-TestRenderItemHeadingDepth-c23dda8: TestRenderItemHeadingDepth
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/writer/writer_test.go
x-reqmd.source-line: 72
x-reqmd.symbol-kind: function
trace:
  - IMP-writer-RenderItem-136994f
```

### IMP-writer-TestRenderItemFieldsDeterministic-dfa5678: TestRenderItemFieldsDeterministic
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/writer/writer_test.go
x-reqmd.source-line: 89
x-reqmd.symbol-kind: function
trace:
  - IMP-writer-RenderItem-136994f
```

### IMP-writer-TestRenderItemUnsupportedFieldType-6cb9e56: TestRenderItemUnsupportedFieldType
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/writer/writer_test.go
x-reqmd.source-line: 119
x-reqmd.symbol-kind: function
trace:
  - IMP-writer-RenderItem-136994f
```

### IMP-writer-TestIsGenerated-7b1177d: TestIsGenerated
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/writer/writer_test.go
x-reqmd.source-line: 131
x-reqmd.symbol-kind: function
trace:
  - IMP-writer-IsGenerated-a943579
```

### IMP-writer-TestWriteIdempotent-90fb814: TestWriteIdempotent
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/writer/writer_test.go
x-reqmd.source-line: 150
x-reqmd.symbol-kind: function
trace:
  - IMP-writer-Write-44afbbf
```

TestWriteIdempotent verifies a second Write produces byte-identical files.

### IMP-writer-TestWriteDoesNotClobberHandAuthored-44e4a21: TestWriteDoesNotClobberHandAuthored
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/writer/writer_test.go
x-reqmd.source-line: 170
x-reqmd.symbol-kind: function
trace:
  - IMP-writer-Write-44afbbf
```

TestWriteDoesNotClobberHandAuthored verifies a hand-authored file (no
generated marker) is left untouched.

### IMP-writer-readAll-288c5bf: readAll
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/writer/writer_test.go
x-reqmd.source-line: 188
x-reqmd.symbol-kind: function
trace: []
```

### IMP-writer-TestRenderSection-3b14bb3: TestRenderSection
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/writer/writer_test.go
x-reqmd.source-line: 217
x-reqmd.symbol-kind: function
trace: []
```

TestRenderSection covers the two section shapes: a titled section renders a
heading with no reqmd block, and an untitled section renders as a plain prose
block (the document preamble and untitled sections use the same form).

### IMP-writer-TestQuoteIfNeeded-3ed73fa: TestQuoteIfNeeded
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/writer/writer_test.go
x-reqmd.source-line: 239
x-reqmd.symbol-kind: function
trace:
  - IMP-writer-quoteIfNeeded-5d24016
```

### IMP-writer-TestQuoteIfNeededRoundTrip-bf7ce49: TestQuoteIfNeededRoundTrip
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/writer/writer_test.go
x-reqmd.source-line: 290
x-reqmd.symbol-kind: function
trace:
  - IMP-writer-quoteIfNeeded-5d24016
```

### IMP-writer-TestRenderItemEscapesHeadingMarkdown-88484c4: TestRenderItemEscapesHeadingMarkdown
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/writer/writer_test.go
x-reqmd.source-line: 316
x-reqmd.symbol-kind: function
trace:
  - IMP-writer-RenderItem-136994f
```

TestRenderItemEscapesHeadingMarkdown pins that heading text carrying inline
Markdown metacharacters stays literal, so a package or symbol named
`__init__` keeps its underscores in the parsed ID and title.

### IMP-writer-TestRenderDocumentEscapesTitle-7a45697: TestRenderDocumentEscapesTitle
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/internal/writer/writer_test.go
x-reqmd.source-line: 340
x-reqmd.symbol-kind: function
trace:
  - IMP-writer-RenderDocument-9c37809
```

TestRenderDocumentEscapesTitle pins the document h1 escaping: an
emphasis-like title must not lose its underscores.

