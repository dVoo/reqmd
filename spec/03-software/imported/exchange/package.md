# Package: exchange

<!-- reqmd: generated -->

## reqmd-import/exchange/build.go
### IMP-exchange-maxProblems-7f31d48: maxProblems
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 18
x-reqmd.symbol-kind: const
trace: []
```

maxProblems caps how many violations one run reports, so a large broken
bundle produces a readable list instead of thousands of lines.

### IMP-exchange-maxRenderDepth-376c8d1: maxRenderDepth
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 23
x-reqmd.symbol-kind: const
trace: []
```

maxRenderDepth is the deepest nesting the renderer accepts: depth 0 renders
"##", so depth 4 is "######" and depth 5 would be seven hashes, which
CommonMark does not treat as a heading.

### IMP-exchange-Options-676fff4: Options
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 26
x-reqmd.symbol-kind: type
trace: []
```

Options carries the renderer's non-format configuration.

### IMP-exchange-ValidationError-1f2139d: ValidationError
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 35
x-reqmd.symbol-kind: type
trace: []
```

ValidationError reports every violation found in one pass. Producers fix a
batch, not one error at a time.

#### IMP-exchange-Error-422078f: Error
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 39
x-reqmd.symbol-kind: method
trace: []
```

### IMP-exchange-statusAdditionPattern-5a9cfce: statusAdditionPattern
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 50
x-reqmd.symbol-kind: var
trace: []
```

statusAdditionPattern matches x-reqmd.additional-status-values entries.

### IMP-exchange-Build-c781f0a: Build
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 56
x-reqmd.symbol-kind: method
trace: []
```

Build validates the bundle and renders every document to output files. A
document without items produces only its schema.yaml; one with items also
produces its .md file. Items keep the order they were given in, at every
level — the renderer never reorders, normalizes or drops anything.

### IMP-exchange-BuildWith-bce6596: BuildWith
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 61
x-reqmd.symbol-kind: method
trace: []
```

BuildWith is Build with renderer options (status mapping).

### IMP-exchange-validator-a99a8b0: validator
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 71
x-reqmd.symbol-kind: type
trace: []
```

validator accumulates violations and the per-document render state.

#### IMP-exchange-addf-8b7cd0f: addf
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 87
x-reqmd.symbol-kind: method
trace: []
```

#### IMP-exchange-validate-3741726: validate
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 94
x-reqmd.symbol-kind: method
trace: []
```

#### IMP-exchange-totalProblems-5548a8e: totalProblems
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 154
x-reqmd.symbol-kind: method
trace: []
```

totalProblems reports how many violations were seen, including the ones
dropped by the cap.

#### IMP-exchange-validateDocument-9261b7b: validateDocument
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 156
x-reqmd.symbol-kind: method
trace: []
```

#### IMP-exchange-validateItems-52154c2: validateItems
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 193
x-reqmd.symbol-kind: method
trace: []
```

#### IMP-exchange-validateFields-e963123: validateFields
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 246
x-reqmd.symbol-kind: method
trace: []
```

#### IMP-exchange-validateLinks-6e1c8d7: validateLinks
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 264
x-reqmd.symbol-kind: method
trace: []
```

#### IMP-exchange-lookup-181ae86: lookup
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 287
x-reqmd.symbol-kind: method
trace: []
```

lookup finds an item by key in a document and returns its kind.

#### IMP-exchange-validateStatus-7d55228: validateStatus
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 292
x-reqmd.symbol-kind: method
trace: []
```

#### IMP-exchange-recordAddition-a0a5a79: recordAddition
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 317
x-reqmd.symbol-kind: method
trace: []
```

recordAddition extends a document's status enum for a non-built-in status.

#### IMP-exchange-render-d44c420: render
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 339
x-reqmd.symbol-kind: method
trace: []
```

#### IMP-exchange-renderItems-871fa34: renderItems
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 381
x-reqmd.symbol-kind: method
trace: []
```

renderItems flattens the item tree into render order, resolving IDs, links,
the status mapping and each item's heading depth.

#### IMP-exchange-mapStatus-97e68a8: mapStatus
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 426
x-reqmd.symbol-kind: method
trace: []
```

mapStatus returns the reqmd status to render and, when the source value could
not be mapped, the original value to preserve as provenance.

#### IMP-exchange-buildRefIndex-d719c53: buildRefIndex
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 443
x-reqmd.symbol-kind: method
trace: []
```

#### IMP-exchange-resolveLink-d8a7f2c: resolveLink
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 459
x-reqmd.symbol-kind: method
trace: []
```

#### IMP-exchange-renderSchema-5b3321b: renderSchema
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 491
x-reqmd.symbol-kind: method
trace: []
```

### IMP-exchange-isBuiltinStatus-7d51a2b: isBuiltinStatus
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 325
x-reqmd.symbol-kind: function
trace: []
```

isBuiltinStatus reports whether a status is one of reqmd's built-ins.

### IMP-exchange-isReqmdStatus-de40e09: isReqmdStatus
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 335
x-reqmd.symbol-kind: function
trace: []
```

isReqmdStatus reports whether a value is a status reqmd can carry: a built-in
or a valid x-reqmd.additional-status-values token.

### IMP-exchange-refIndex-4488c29: refIndex
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 441
x-reqmd.symbol-kind: type
trace: []
```

refIndex maps a key to the reqmd ID of the requirement it names, per document
key, so cross-document links resolve.

### IMP-exchange-qualifyRef-44316ee: qualifyRef
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 481
x-reqmd.symbol-kind: function
trace: []
```

qualifyRef renders a cross-document trace ref. reqmd resolves
"<document-id>/<id>", falling back to "<dir>/<id>", so the emitted ref uses
the target's document-id when it has one and its directory otherwise.

### IMP-exchange-add-8d99cc1: add
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 559
x-reqmd.symbol-kind: function
trace: []
```

add appends a rendered file, rejecting a second document claiming the same
output path.

### IMP-exchange-docName-cb3f89f: docName
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 569
x-reqmd.symbol-kind: function
trace: []
```

### IMP-exchange-Dir-d3c2de4: Dir
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 578
x-reqmd.symbol-kind: method
trace: []
```

Dir is the output directory, from x-reqmd.file's sibling config. It defaults
to the document key so a neutral producer needs no layout knowledge.

### IMP-exchange-FileName-0f9a098: FileName
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 586
x-reqmd.symbol-kind: method
trace: []
```

FileName is the .md filename, defaulting to requirements.md.

### IMP-exchange-prefix-d7ae479: prefix
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 594
x-reqmd.symbol-kind: method
trace: []
```

prefix is the ID prefix, empty for a document without x-reqmd.

### IMP-exchange-itemLabel-09da739: itemLabel
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 604
x-reqmd.symbol-kind: function
trace: []
```

itemLabel is how an item is addressed in an error message: its title when it
has one (that is what a producer or reviewer recognises), else its key, else
a placeholder for a headingless block.

### IMP-exchange-cutDocRef-77c5799: cutDocRef
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 615
x-reqmd.symbol-kind: function
trace: []
```

cutDocRef splits a "doc-key/key" link. A bare key is document-local.

### IMP-exchange-checkSingleLine-fa0f9c0: checkSingleLine
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 626
x-reqmd.symbol-kind: function
trace: []
```

checkSingleLine rejects a producer string that would break the one-line
structure of a heading or document title: a line break injects extra Markdown
lines and can steal the attr block of the next item.

### IMP-exchange-checkAttrKey-f4ed8b0: checkAttrKey
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 637
x-reqmd.symbol-kind: function
trace: []
```

checkAttrKey rejects an attribute name that cannot be written as "key: value".
reqmd's attr scanner splits each line at the first colon and never unquotes
keys, so a key containing a colon would silently arrive under a different
name ("a:b: v" is read as attribute "a").

### IMP-exchange-checkScalarValue-ef036e3: checkScalarValue
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 656
x-reqmd.symbol-kind: function
trace: []
```

checkScalarValue rejects an attribute value the renderer cannot write: reqmd
attr blocks hold scalars and flat lists of scalars, never nested structures.

### IMP-exchange-cleanRelDir-6f2bf1d: cleanRelDir
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 678
x-reqmd.symbol-kind: function
trace: []
```

cleanRelDir normalizes a document directory and rejects paths that escape the
output root.

### IMP-exchange-cleanFileName-82ff162: cleanFileName
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build.go
x-reqmd.source-line: 694
x-reqmd.symbol-kind: function
trace: []
```

cleanFileName rejects a .md name that is not a plain file name in the
document directory.

## reqmd-import/exchange/build\_test.go
### IMP-exchange-bundle-0de44b2: bundle
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build_test.go
x-reqmd.source-line: 15
x-reqmd.symbol-kind: function
trace: []
```

bundle wraps documents in a valid envelope.

### IMP-exchange-render-fbe40ce: render
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build_test.go
x-reqmd.source-line: 20
x-reqmd.symbol-kind: function
trace: []
```

render builds the bundle and returns rendered files keyed by path.

### IMP-exchange-renderErr-1bfb191: renderErr
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build_test.go
x-reqmd.source-line: 34
x-reqmd.symbol-kind: function
trace: []
```

renderErr builds and returns the validation error text.

### IMP-exchange-wantProblems-571ea1c: wantProblems
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build_test.go
x-reqmd.source-line: 45
x-reqmd.symbol-kind: function
trace: []
```

wantProblems asserts that every fragment appears in the error report, i.e.
that violations are batched rather than reported one at a time.

### IMP-exchange-TestBuild\_RendersDocument-7fc2521: TestBuild\_RendersDocument
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build_test.go
x-reqmd.source-line: 54
x-reqmd.symbol-kind: function
trace:
  - IMP-exchange-Build-c781f0a
```

### IMP-exchange-TestBuild\_NestedSectionsAndDepth-c5a4b4e: TestBuild\_NestedSectionsAndDepth
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build_test.go
x-reqmd.source-line: 115
x-reqmd.symbol-kind: function
trace:
  - IMP-exchange-Build-c781f0a
```

TestBuild_NestedSectionsAndDepth covers the tree: a section renders a heading
with children one level deeper, and no attr block.

### IMP-exchange-TestBuild\_LinksResolve-10292b4: TestBuild\_LinksResolve
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build_test.go
x-reqmd.source-line: 162
x-reqmd.symbol-kind: function
trace:
  - IMP-exchange-Build-c781f0a
```

TestBuild_LinksResolve covers document-local and cross-document links: a link
becomes a trace ref, qualified by the target document's id when it has one.

### IMP-exchange-TestBuild\_CrossDocumentLinkFallsBackToDir-aa509d7: TestBuild\_CrossDocumentLinkFallsBackToDir
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build_test.go
x-reqmd.source-line: 192
x-reqmd.symbol-kind: function
trace:
  - IMP-exchange-Build-c781f0a
```

TestBuild_CrossDocumentLinkFallsBackToDir covers a target document without a
document-id: reqmd resolves "<dir>/<id>", so the ref must use the dir.

### IMP-exchange-TestBuild\_StatusMapping-48d7331: TestBuild\_StatusMapping
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build_test.go
x-reqmd.source-line: 223
x-reqmd.symbol-kind: function
trace:
  - IMP-exchange-Build-c781f0a
```

TestBuild_StatusMapping covers the renderer-side status map: a mapped value is
used and extends the document enum; an unmapped one falls back to the default
and keeps the original as provenance.

### IMP-exchange-TestBuild\_SchemaOnlyDocument-3ae4b47: TestBuild\_SchemaOnlyDocument
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build_test.go
x-reqmd.source-line: 264
x-reqmd.symbol-kind: function
trace:
  - IMP-exchange-Build-c781f0a
```

TestBuild_SchemaOnlyDocument covers a document without items.

### IMP-exchange-TestBuild\_DecodeRejectsUnknownFields-42e71c2: TestBuild\_DecodeRejectsUnknownFields
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build_test.go
x-reqmd.source-line: 274
x-reqmd.symbol-kind: function
trace:
  - IMP-exchange-Build-c781f0a
```

### IMP-exchange-TestBuild\_ItemTreeKeepsInputOrder-82fb12a: TestBuild\_ItemTreeKeepsInputOrder
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build_test.go
x-reqmd.source-line: 282
x-reqmd.symbol-kind: function
trace:
  - IMP-exchange-Build-c781f0a
```

### IMP-exchange-TestBuild\_BatchesViolations-6858d71: TestBuild\_BatchesViolations
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build_test.go
x-reqmd.source-line: 299
x-reqmd.symbol-kind: function
trace:
  - IMP-exchange-Build-c781f0a
```

TestBuild_BatchesViolations is the core producer-facing guarantee: every
violation is reported in one run instead of one at a time.

### IMP-exchange-TestBuild\_ValidationRules-5c56e6f: TestBuild\_ValidationRules
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build_test.go
x-reqmd.source-line: 321
x-reqmd.symbol-kind: function
trace:
  - IMP-exchange-Build-c781f0a
```

### IMP-exchange-nestItems-5ea8ad8: nestItems
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build_test.go
x-reqmd.source-line: 454
x-reqmd.symbol-kind: function
trace: []
```

nestItems builds a chain of sections deep enough to exceed the render budget.

### IMP-exchange-TestDecodeRejectsDeepNesting-4e1659c: TestDecodeRejectsDeepNesting
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build_test.go
x-reqmd.source-line: 461
x-reqmd.symbol-kind: function
trace:
  - IMP-exchange-Decode-a1934ec
```

### IMP-exchange-TestDecodeRejectsTrailingData-e75bfe0: TestDecodeRejectsTrailingData
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build_test.go
x-reqmd.source-line: 479
x-reqmd.symbol-kind: function
trace:
  - IMP-exchange-Decode-a1934ec
```

### IMP-exchange-TestBuild\_IsIdempotent-929e54c: TestBuild\_IsIdempotent
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build_test.go
x-reqmd.source-line: 489
x-reqmd.symbol-kind: function
trace:
  - IMP-exchange-Build-c781f0a
```

TestBuild_IsIdempotent verifies byte-determinism: the same bundle renders the
same bytes, so re-running an import is a no-op.

### IMP-exchange-TestValidationError\_ListsEveryProblem-fbe4c3d: TestValidationError\_ListsEveryProblem
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/build_test.go
x-reqmd.source-line: 525
x-reqmd.symbol-kind: function
trace:
  - IMP-exchange-ValidationError-1f2139d
```

TestValidationError_ListsEveryProblem pins the report shape producers rely on.

## reqmd-import/exchange/exchange.go
### IMP-exchange-SchemaURL-f088c9e: SchemaURL
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/exchange.go
x-reqmd.source-line: 21
x-reqmd.symbol-kind: const
trace: []
```

SchemaURL is the value a bundle must declare in its "$schema" field. It is
also the identity of the format: a breaking change moves to a new URL
instead of growing an in-place version.

### IMP-exchange-KindRequirement-90a4237: KindRequirement
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/exchange.go
x-reqmd.source-line: 25
x-reqmd.symbol-kind: const
trace: []
```

Item kinds. A requirement renders a heading plus an attr block and takes part
in validation, tracing and coverage; a section renders structure only.

### IMP-exchange-KindSection-131ac5b: KindSection
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/exchange.go
x-reqmd.source-line: 25
x-reqmd.symbol-kind: const
trace: []
```

Item kinds. A requirement renders a heading plus an attr block and takes part
in validation, tracing and coverage; a section renders structure only.

### IMP-exchange-maxDepth-8eb5541: maxDepth
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/exchange.go
x-reqmd.source-line: 33
x-reqmd.symbol-kind: const
trace: []
```

maxDepth bounds bundle nesting at decode time. The renderer separately
rejects anything deeper than six heading levels; this guard only keeps a
hostile bundle from exhausting the stack.

### IMP-exchange-Bundle-18665d2: Bundle
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/exchange.go
x-reqmd.source-line: 37
x-reqmd.symbol-kind: type
trace: []
```

Bundle is the top-level exchange document: the documents to render, in
output order.

### IMP-exchange-Document-dfafec7: Document
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/exchange.go
x-reqmd.source-line: 44
x-reqmd.symbol-kind: type
trace: []
```

Document describes one reqmd document directory: one schema.yaml plus, when
it has items, one .md file.

### IMP-exchange-XReqmd-eb3fed7: XReqmd
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/exchange.go
x-reqmd.source-line: 65
x-reqmd.symbol-kind: type
trace: []
```

XReqmd is every reqmd-specific decision for one document. A producer that
only knows the neutral part of the format omits it.

### IMP-exchange-Upstream-39ed6e7: Upstream
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/exchange.go
x-reqmd.source-line: 87
x-reqmd.symbol-kind: type
trace: []
```

Upstream is a document's x-reqmd.upstream block.

### IMP-exchange-Item-095f5f1: Item
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/exchange.go
x-reqmd.source-line: 95
x-reqmd.symbol-kind: type
trace: []
```

Item is one content item. Nesting replaces a parent pointer entirely: a child
cannot reference a missing parent, cannot be listed before it, and cannot
form a cycle.

### IMP-exchange-Decode-a1934ec: Decode
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/exchange.go
x-reqmd.source-line: 124
x-reqmd.symbol-kind: function
trace: []
```

Decode reads one exchange bundle. Unknown fields are rejected so a producer
typo fails loudly instead of silently dropping data, and nesting deeper than
maxDepth is refused before it can exhaust the stack.

### IMP-exchange-checkNesting-2671524: checkNesting
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/exchange.go
x-reqmd.source-line: 146
x-reqmd.symbol-kind: function
trace: []
```

checkNesting rejects a bundle nested deeper than maxDepth.

## reqmd-import/exchange/fixtures\_test.go
### IMP-exchange-schemaPath-fbc19b6: schemaPath
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/fixtures_test.go
x-reqmd.source-line: 17
x-reqmd.symbol-kind: const
trace: []
```

schemaPath is the published contract this format claims to implement. The
fixtures below are the shared corpus a third-party producer can copy: every
valid bundle must satisfy both the schema (shape) and Build (semantics), and
every invalid bundle names the rule it breaks.

### IMP-exchange-validFixtures-200e162: validFixtures
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/fixtures_test.go
x-reqmd.source-line: 20
x-reqmd.symbol-kind: var
trace: []
```

validFixtures must pass the schema and Build.

### IMP-exchange-invalidFixtures-ed4f6df: invalidFixtures
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/fixtures_test.go
x-reqmd.source-line: 32
x-reqmd.symbol-kind: var
trace: []
```

invalidFixtures must be rejected. Which rule each one breaks, and with which
message, is asserted once in build_test.go: the corpus's job is that the
published schema and the renderer agree about what is legal, not to restate
the wording of every error.

### IMP-exchange-TestFixtures\_SchemaAcceptsValidBundles-86a7784: TestFixtures\_SchemaAcceptsValidBundles
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/fixtures_test.go
x-reqmd.source-line: 56
x-reqmd.symbol-kind: function
trace: []
```

TestFixtures_SchemaAcceptsValidBundles is the neutrality check: the published
schema alone accepts every well-formed bundle, with no Go-specific knowledge.

### IMP-exchange-TestFixtures\_ValidBundlesBuild-54be37a: TestFixtures\_ValidBundlesBuild
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/fixtures_test.go
x-reqmd.source-line: 84
x-reqmd.symbol-kind: function
trace: []
```

TestFixtures_ValidBundlesBuild verifies every valid fixture renders.

### IMP-exchange-TestFixtures\_InvalidBundlesAreRejected-a84c3e7: TestFixtures\_InvalidBundlesAreRejected
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/fixtures_test.go
x-reqmd.source-line: 101
x-reqmd.symbol-kind: function
trace: []
```

TestFixtures_InvalidBundlesAreRejected verifies the corpus's negative half:
nothing in it renders. A rejection at decode time (unknown field, nesting
guard) counts the same as one at build time — both are loud.

### IMP-exchange-readFixture-e1b130f: readFixture
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/fixtures_test.go
x-reqmd.source-line: 116
x-reqmd.symbol-kind: function
trace: []
```

readFixture loads one fixture from the corpus.

## reqmd-import/exchange/outline\_test.go
### IMP-exchange-outlineHeading-a41963a: outlineHeading
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/outline_test.go
x-reqmd.source-line: 17
x-reqmd.symbol-kind: type
trace: []
```

This file holds the structural assertions the rendering tests use instead of
substring matching. A rendered document is asserted through its outline —
heading level, heading text, and the parsed attr block — so a change in
formatting (spacing, quoting style, key order) does not break a test that is
really about structure, while a change in nesting or identity does.
outlineHeading is one rendered heading.

### IMP-exchange-outline-e94c7e4: outline
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/outline_test.go
x-reqmd.source-line: 26
x-reqmd.symbol-kind: function
trace: []
```

outline parses a rendered document into its headings. It only understands the
shape the renderer produces (ATX headings, opaque ```attr blocks), which is
exactly what the assertions need.

### IMP-exchange-atxLevel-450428d: atxLevel
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/outline_test.go
x-reqmd.source-line: 63
x-reqmd.symbol-kind: function
trace: []
```

atxLevel parses "### Text" into its level and text.

### IMP-exchange-heading-c8a7344: heading
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/outline_test.go
x-reqmd.source-line: 76
x-reqmd.symbol-kind: function
trace: []
```

heading returns the outline entry whose text contains want, failing when it
is absent or ambiguous.

### IMP-exchange-attrString-c97aca4: attrString
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/outline_test.go
x-reqmd.source-line: 95
x-reqmd.symbol-kind: function
trace: []
```

attrString reads a string attribute from a parsed heading.

### IMP-exchange-attrList-6998c75: attrList
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-import/exchange/outline_test.go
x-reqmd.source-line: 109
x-reqmd.symbol-kind: function
trace: []
```

attrList reads a list attribute from a parsed heading.

