# Package: rust

<!-- reqmd: generated -->

## reqmd-code/internal/lang/rust/rust.go
### IMP-rust-queriesFS-0116fdc: queriesFS
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/rust/rust.go
x-reqmd.source-line: 45
x-reqmd.symbol-kind: var
trace: []
```

go:embed queries/*.scm

### IMP-rust-RustLanguage-34e24e1: RustLanguage
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/rust/rust.go
x-reqmd.source-line: 48
x-reqmd.symbol-kind: type
trace: []
```

RustLanguage is the Rust language plugin. Implements lang.Language.

#### IMP-rust-Name-226a996: Name
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/rust/rust.go
x-reqmd.source-line: 54
x-reqmd.symbol-kind: method
trace: []
```

Name returns the canonical language identifier.

#### IMP-rust-Extensions-cb2e0fb: Extensions
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/rust/rust.go
x-reqmd.source-line: 57
x-reqmd.symbol-kind: method
trace: []
```

Extensions returns the file extensions this plugin handles.

#### IMP-rust-Grammar-51e70ce: Grammar
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/rust/rust.go
x-reqmd.source-line: 60
x-reqmd.symbol-kind: method
trace: []
```

Grammar returns the tree-sitter-rust grammar language.

#### IMP-rust-Query-ee356f3: Query
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/rust/rust.go
x-reqmd.source-line: 67
x-reqmd.symbol-kind: method
trace: []
```

Query returns the embedded tree-sitter query for symbol extraction.
If the query file is missing (a packaging error), it returns an empty
query rather than panicking; Parse will then emit zero symbols.

#### IMP-rust-BindDoc-c2956d0: BindDoc
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/rust/rust.go
x-reqmd.source-line: 80
x-reqmd.symbol-kind: method
trace: []
```

BindDoc returns the cleaned doc comment for the given declaration
node. It walks backwards over any contiguous run of `//` line
comments and `/* ... */` block comments, joins them with newlines,
and strips comment markers (including `///` and `//!` doc-comment
forms). Returns the empty string when no preceding comment is present.

#### IMP-rust-Parse-569329c: Parse
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/rust/rust.go
x-reqmd.source-line: 212
x-reqmd.symbol-kind: method
trace: []
```

Parse parses the given Rust source bytes and returns a flat list of
top-level symbols, in source order. It uses the embedded tree-sitter
query to find function, method, type, const, and static declarations,
and walks the parent chain to drop impl/trait-body function items from
the free-function pass (those are methods) and to populate a method's
Receiver from its enclosing impl/trait.

##### Dedup strategy
The query produces intentional double-matches for the same source
node: a method's `function_item` matches both the bare
`function_item` pattern (@func.decl) and the impl-body-anchored
pattern (@method.decl). To collapse these to a single emitted symbol
we track emitted method decl nodes by `*sitter.Node` identity, then
skip the bare @func.decl match that points at the same node.

### IMP-rust-cleanComment-9cb394b: cleanComment
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/rust/rust.go
x-reqmd.source-line: 125
x-reqmd.symbol-kind: function
trace: []
```

cleanComment strips Rust comment markers from a comment node's text
and trims per-line whitespace. It accepts both `//`-style (including
doc forms `///` and `//!`) and `/* ... */`-style (including `/**`
and `/*!`) comments. The conventional single space after the marker
(`// foo` → `foo`) is dropped.

### IMP-rust-testAttrRe-54f98b7: testAttrRe
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/rust/rust.go
x-reqmd.source-line: 165
x-reqmd.symbol-kind: var
trace: []
```

testAttrRe matches a test attribute, including common paths such as
`#[tokio::test]`; it deliberately does not match `#[cfg(test)]`.

### IMP-rust-hasTestAttribute-4807e65: hasTestAttribute
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/rust/rust.go
x-reqmd.source-line: 171
x-reqmd.symbol-kind: function
trace: []
```

hasTestAttribute reports whether the item carries a `#[test]` attribute.
It walks the same sibling run BindDoc does: attributes and comments may sit
between the item and the attribute, while an inner attribute or inner doc
comment ends the run.

### IMP-rust-declNode-87c33dd: declNode
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/rust/rust.go
x-reqmd.source-line: 296
x-reqmd.symbol-kind: function
trace: []
```

declNode returns the primary declaration node for a match, if any.

### IMP-rust-symbolFromMatch-b138cba: symbolFromMatch
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/rust/rust.go
x-reqmd.source-line: 319
x-reqmd.symbol-kind: function
trace: []
```

symbolFromMatch builds a model.Symbol from the per-match capture map.
The map keys are the @-names from queries/symbols.scm. Returns false
when the match has no usable captures or refers to a function-local
item.

### IMP-rust-enclosingTypeName-9e3106a: enclosingTypeName
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/rust/rust.go
x-reqmd.source-line: 375
x-reqmd.symbol-kind: function
trace: []
```

enclosingTypeName walks up the parent chain from a method node to the
enclosing impl_item or trait_item and returns the owner's type name
(e.g. "Point" for `impl Point { fn scale(&self) }` and "Shape" for
`trait Shape { fn area(&self) }`). Generic and scoped impl types
(`impl Point<T>`, `impl crate::Point`) are reduced to their base
type identifier so the receiver matches the parent type's name.
Returns "" when the method has no enclosing impl/trait.

### IMP-rust-init-d1988fb: init
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/rust/rust.go
x-reqmd.source-line: 420
x-reqmd.symbol-kind: function
trace: []
```

## reqmd-code/internal/lang/rust/rust\_test.go
### IMP-rust-TestRustLanguageRegisters-ae8e5ad: TestRustLanguageRegisters
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/rust/rust_test.go
x-reqmd.source-line: 13
x-reqmd.symbol-kind: function
trace:
  - IMP-rust-RustLanguage-34e24e1
```

TestRustLanguageRegisters verifies that RustLanguage self-registered
itself with the global lang.Registry on package init.

### IMP-rust-TestBindDocAboveTestAttribute-1ffbc53: TestBindDocAboveTestAttribute
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/rust/rust_test.go
x-reqmd.source-line: 26
x-reqmd.symbol-kind: function
trace:
  - IMP-rust-BindDoc-c2956d0
```

TestBindDocAboveTestAttribute verifies that a doc comment above a
`#[test]` attribute still binds to the function: the attribute sits
between the comment and the item.

### IMP-rust-TestBindDocStopsAtInnerAttribute-a057f04: TestBindDocStopsAtInnerAttribute
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/rust/rust_test.go
x-reqmd.source-line: 50
x-reqmd.symbol-kind: function
trace:
  - IMP-rust-BindDoc-c2956d0
```

TestBindDocStopsAtInnerAttribute verifies that a module-level inner
attribute does not let the crate doc bind to the first item.

### IMP-rust-TestBindDocStopsAtModuleDoc-e011a42: TestBindDocStopsAtModuleDoc
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/rust/rust_test.go
x-reqmd.source-line: 74
x-reqmd.symbol-kind: function
trace:
  - IMP-rust-BindDoc-c2956d0
```

TestBindDocStopsAtModuleDoc verifies that an inner doc comment (`//!`)
does not bind to the first item: module docs are not item docs.

### IMP-rust-TestParseSample-051aa5e: TestParseSample
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/rust/rust_test.go
x-reqmd.source-line: 99
x-reqmd.symbol-kind: function
trace:
  - IMP-rust-Parse-569329c
```

TestParseSample verifies that Parse correctly identifies and classifies
all top-level symbols in a representative Rust file: a const, a struct,
an enum, a free function, a trait with a method signature, a static,
and an impl block method. The impl method must not be double-counted as
a free function, and its Receiver must be the impl'd type.

### IMP-rust-TestParseTraitDefaultMethod-3c13c58: TestParseTraitDefaultMethod
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/rust/rust_test.go
x-reqmd.source-line: 188
x-reqmd.symbol-kind: function
trace:
  - IMP-rust-Parse-569329c
```

TestParseTraitDefaultMethod verifies that a trait method with a default
body (function_item, not function_signature_item) is still a method.

### IMP-rust-TestParseGenericImpl-b6bc2bb: TestParseGenericImpl
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/rust/rust_test.go
x-reqmd.source-line: 213
x-reqmd.symbol-kind: function
trace:
  - IMP-rust-Parse-569329c
```

TestParseGenericImpl verifies that the receiver of a method on a generic
impl (`impl Point<T>`) resolves to the base type identifier.

### IMP-rust-TestParseLocalItemsSkipped-b7af57b: TestParseLocalItemsSkipped
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/rust/rust_test.go
x-reqmd.source-line: 238
x-reqmd.symbol-kind: function
trace:
  - IMP-rust-Parse-569329c
```

TestParseLocalItemsSkipped verifies that items declared inside
function bodies (local consts, local structs) are not emitted as
top-level symbols.

### IMP-rust-TestParseEmpty-c20edd5: TestParseEmpty
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/rust/rust_test.go
x-reqmd.source-line: 272
x-reqmd.symbol-kind: function
trace:
  - IMP-rust-Parse-569329c
```

TestParseEmpty handles zero-byte input gracefully.

### IMP-rust-TestCleanComment-bbd700b: TestCleanComment
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/rust/rust_test.go
x-reqmd.source-line: 285
x-reqmd.symbol-kind: function
trace:
  - IMP-rust-cleanComment-9cb394b
```

TestCleanComment covers `//`, `///`, `//!`, `/* */`, `/** */` and
`/*! */` comment forms.

### IMP-rust-TestBindDocIncludesDoc-094039d: TestBindDocIncludesDoc
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/rust/rust_test.go
x-reqmd.source-line: 306
x-reqmd.symbol-kind: function
trace:
  - IMP-rust-BindDoc-c2956d0
```

TestBindDocIncludesDoc verifies that BindDoc attaches the full doc
comment block (including a reqmd:trace marker) to the following item.

### IMP-rust-TestParseMarksTestAttributes-4330cef: TestParseMarksTestAttributes
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/rust/rust_test.go
x-reqmd.source-line: 328
x-reqmd.symbol-kind: function
trace:
  - IMP-rust-Parse-569329c
```

TestParseMarksTestAttributes pins IsTest detection, including path-qualified
test attributes and inline test modules.

