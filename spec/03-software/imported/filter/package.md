# Package: filter

<!-- reqmd: generated -->

## internal/filter/filter.go
### IMP-filter-builtInVars-7de697e: builtInVars
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter.go
x-reqmd.source-line: 36
x-reqmd.symbol-kind: var
trace: []
```

builtInVars are always available in the filter scope regardless of
schema declarations. `id` and `title` live on model.Node, not in
Attrs; the rest are the RFC §2.3 built-in attributes (status,
disposition, trace, version). They are always accepted by the typo
check even when no schema in the tree declares them; at evaluation time
a requirement lacking the attribute sees nil (see AllowUndefinedVariables).

### IMP-filter-exprBuiltins-d2b10a6: exprBuiltins
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter.go
x-reqmd.source-line: 48
x-reqmd.symbol-kind: var
trace: []
```

exprBuiltins is the set of expr built-in function names that are NOT
variable references. We exclude them from attribute validation so that
`id startsWith "SYS-"` or `len(variant) > 0` don't trip the typo check.

### IMP-filter-Filter-fe786ff: Filter
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter.go
x-reqmd.source-line: 67
x-reqmd.symbol-kind: type
trace: []
```

Filter is a compiled attribute filter expression. It is safe for
concurrent use after construction — the underlying vm.Program is
read-only.

#### IMP-filter-Expr-2023150: Expr
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter.go
x-reqmd.source-line: 175
x-reqmd.symbol-kind: method
trace: []
```

Expr returns the original expression string.

#### IMP-filter-Refs-477c375: Refs
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter.go
x-reqmd.source-line: 179
x-reqmd.symbol-kind: method
trace: []
```

Refs returns the sorted, deduplicated list of variable names
referenced in the expression (excluding expr built-in functions).

#### IMP-filter-Match-57b6e5c: Match
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter.go
x-reqmd.source-line: 184
x-reqmd.symbol-kind: method
trace: []
```

Match evaluates the compiled expression against a single requirement.
Returns (true, nil) if the requirement matches, (false, nil) if not,
and (false, err) if evaluation fails (e.g. type mismatch).

#### IMP-filter-FilterDocs-2f82f12: FilterDocs
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter.go
x-reqmd.source-line: 203
x-reqmd.symbol-kind: method
trace: []
```

FilterDocs returns a new []model.Document containing only matching
requirements per document. Non-requirement nodes (containers and info
items) are always kept, so document structure and content survive
filtering. Documents with all requirements filtered out are kept (with
only items) so document-level structure is preserved for list/stats/
export output.

#### IMP-filter-FilterDocsAndIDs-97e9fb7: FilterDocsAndIDs
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter.go
x-reqmd.source-line: 212
x-reqmd.symbol-kind: method
trace: []
```

FilterDocsAndIDs is the single-pass form of FilterDocs + MatchingIDs: the
same pruned documents and the same matching-ID set from one evaluation per
requirement. Used by the exports and serve, which need both views of the
same filtered tree.

#### IMP-filter-MatchingIDs-7e11246: MatchingIDs
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter.go
x-reqmd.source-line: 230
x-reqmd.symbol-kind: method
trace: []
```

MatchingIDs returns a set of requirement IDs that match the filter.
Used by check for filter-aware graph checks — the graph is
built from the full tree, but only matching reqs are checked/reported.

#### IMP-filter-pruneNodes-5c73a8e: pruneNodes
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter.go
x-reqmd.source-line: 264
x-reqmd.symbol-kind: method
trace: []
```

pruneNodes returns the subset of a node list to keep: requirements are kept
only when they match the filter; children of a pruned requirement are
re-attached to the pruned node's position so no matching requirement is lost.

Sections are kept only when they still lead somewhere: a section with
surviving children stays (with their prose), a section that had children but
lost them all is dropped — its prose was an introduction to content the
filter removed — and a bare heading with no prose at all is dropped. A
section that never had children and carries prose (a leaf information item,
or the headingless document preamble) is content and is kept regardless of
the filter. This rule is presentational: it never affects MatchingIDs, so
`check` stays requirement-scoped.

Kept nodes are shallow copies — the input tree is never mutated, so multiple
filtered views of the same documents stay independent.

Every matching requirement's ID is added to ids, so one walk yields both the
pruned tree and the MatchingIDs set.

### IMP-filter-Compile-a260df6: Compile
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter.go
x-reqmd.source-line: 86
x-reqmd.symbol-kind: function
trace:
  - SW-FIL-001
```

reqmd:trace SW-FIL-001
Compile parses, validates, and compiles a filter expression (strict
mode).

validAttrs is the union of all schema property names (plus built-in
variable names) across every document in the tree. Referencing a name
not in this set is a compile-time ERROR — this catches typos and
references to attributes that no document declares.

The expression is compiled with AllowUndefinedVariables so that a
requirement whose document does not declare a referenced attribute
evaluates that variable to nil (the req is excluded unless the
expression handles absence, e.g. `variant == null`).

### IMP-filter-CompileRelaxed-1eadd2d: CompileRelaxed
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter.go
x-reqmd.source-line: 99
x-reqmd.symbol-kind: function
trace: []
```

CompileRelaxed compiles a filter expression without hard-failing on
attribute names that are not declared in any schema.yaml. validAttrs is
the expanded allowed set (schema-declared names plus every attribute
actually found on a requirement); declaredAttrs is the schema-declared
set plus built-ins and drives which references produce a warning. Every
referenced name outside declaredAttrs yields a warning instead of an
error, so undeclared-but-present attributes become filterable while true
typos no longer abort the run.

### IMP-filter-compile-d149656: compile
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter.go
x-reqmd.source-line: 107
x-reqmd.symbol-kind: function
trace: []
```

compile is the shared implementation for strict and relaxed modes.
strict=true returns an error for any referenced name outside validAttrs
(the global typo check). strict=false never errors on unknown names;
instead every name outside declaredAttrs is collected as a warning.

### IMP-filter-buildEnv-76a4b5e: buildEnv
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter.go
x-reqmd.source-line: 304
x-reqmd.symbol-kind: function
trace: []
```

buildEnv constructs the expr evaluation environment for a requirement:
the req's Attrs map plus `id` and `title` injected as top-level vars.
We copy into a new map to avoid mutating the original Attrs.

### IMP-filter-BuildValidAttrs-4f190c9: BuildValidAttrs
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter.go
x-reqmd.source-line: 315
x-reqmd.symbol-kind: function
trace: []
```

BuildValidAttrs computes the union of all schema property names across
every document plus the built-in variable names. The result is used by
Compile for strict attribute validation.

### IMP-filter-BuildFoundAttrs-157302b: BuildFoundAttrs
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter.go
x-reqmd.source-line: 332
x-reqmd.symbol-kind: function
trace: []
```

BuildFoundAttrs computes the union of every attribute key actually
present on some requirement in the tree, whether or not it is declared
in a schema.yaml. Relaxed filtering unions this with BuildValidAttrs so
undeclared-but-present attributes can be filtered.

### IMP-filter-WriteWarnings-d60a791: WriteWarnings
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter.go
x-reqmd.source-line: 347
x-reqmd.symbol-kind: function
trace: []
```

WriteWarnings writes relaxed-mode filter warnings to w, one per line,
prefixed with "reqmd: warning: ". It is a no-op when warnings is empty
so callers can invoke it unconditionally.

### IMP-filter-CompileForDocs-787afca: CompileForDocs
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter.go
x-reqmd.source-line: 362
x-reqmd.symbol-kind: function
trace: []
```

CompileForDocs compiles an expression against the attribute union of the
given documents. It is the shared entry point for every CLI command that
accepts a --filter flag: Compile + BuildValidAttrs in one call.

When allowUndeclared is false the strict typo check applies and any
reference to an attribute not declared in a schema.yaml is an error.
When true, references to attributes found on any requirement are allowed
(undeclared-but-present) and every out-of-schema reference is returned as
a warning instead of an error. Returns (nil, nil, nil) when expr is empty.

### IMP-filter-refCollector-b8d7be1: refCollector
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter.go
x-reqmd.source-line: 381
x-reqmd.symbol-kind: type
trace: []
```

refCollector implements ast.Visitor and collects all IdentifierNode
values from the expression AST. These are cross-referenced against
validAttrs to catch typos.

#### IMP-filter-Visit-a17ddac: Visit
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter.go
x-reqmd.source-line: 385
x-reqmd.symbol-kind: method
trace: []
```

## internal/filter/filter\_test.go
### IMP-filter-testReq-328520b: testReq
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 13
x-reqmd.symbol-kind: function
trace: []
```

### IMP-filter-testDocs-202d8ea: testDocs
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 22
x-reqmd.symbol-kind: function
trace: []
```

### IMP-filter-validAttrsFor-d5cca1a: validAttrsFor
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 59
x-reqmd.symbol-kind: function
trace: []
```

### IMP-filter-TestCompileAndMatch\_Equality-8fffca1: TestCompileAndMatch\_Equality
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 63
x-reqmd.symbol-kind: function
trace:
  - IMP-filter-Compile-a260df6
```

### IMP-filter-TestCompileAndMatch\_ArrayIn-1c93a59: TestCompileAndMatch\_ArrayIn
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 87
x-reqmd.symbol-kind: function
trace:
  - IMP-filter-Compile-a260df6
```

### IMP-filter-TestCompileAndMatch\_NilAttribute-6d118bc: TestCompileAndMatch\_NilAttribute
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 111
x-reqmd.symbol-kind: function
trace:
  - IMP-filter-Compile-a260df6
```

### IMP-filter-TestCompileAndMatch\_NilCoalescing-1f25f20: TestCompileAndMatch\_NilCoalescing
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 128
x-reqmd.symbol-kind: function
trace:
  - IMP-filter-Compile-a260df6
```

### IMP-filter-TestCompileAndMatch\_NumericComparison-12893e9: TestCompileAndMatch\_NumericComparison
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 164
x-reqmd.symbol-kind: function
trace:
  - IMP-filter-Compile-a260df6
```

### IMP-filter-TestCompileAndMatch\_BuiltInIDAndTitle-d27bc68: TestCompileAndMatch\_BuiltInIDAndTitle
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 188
x-reqmd.symbol-kind: function
trace:
  - IMP-filter-Compile-a260df6
```

### IMP-filter-TestCompileAndMatch\_BooleanComposition-3fe2287: TestCompileAndMatch\_BooleanComposition
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 212
x-reqmd.symbol-kind: function
trace:
  - IMP-filter-Compile-a260df6
```

### IMP-filter-TestCompile\_SyntaxError-8019356: TestCompile\_SyntaxError
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 236
x-reqmd.symbol-kind: function
trace:
  - IMP-filter-Compile-a260df6
```

### IMP-filter-TestCompile\_UnknownAttribute-407e2d1: TestCompile\_UnknownAttribute
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 244
x-reqmd.symbol-kind: function
trace:
  - IMP-filter-Compile-a260df6
```

### IMP-filter-TestCompile\_EmptyExpression-1ae7b0a: TestCompile\_EmptyExpression
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 252
x-reqmd.symbol-kind: function
trace:
  - IMP-filter-Compile-a260df6
```

### IMP-filter-TestCompile\_AllowsExprBuiltins-71f36c6: TestCompile\_AllowsExprBuiltins
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 259
x-reqmd.symbol-kind: function
trace:
  - IMP-filter-Compile-a260df6
```

### IMP-filter-TestCompile\_AllowsCrossSchemaAttribute-185a5b4: TestCompile\_AllowsCrossSchemaAttribute
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 269
x-reqmd.symbol-kind: function
trace:
  - IMP-filter-Compile-a260df6
```

### IMP-filter-TestFilterDocs-d2134a8: TestFilterDocs
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 279
x-reqmd.symbol-kind: function
trace:
  - IMP-filter-FilterDocs-2f82f12
```

### IMP-filter-TestFilterDocs\_PreservesDocStructure-0485295: TestFilterDocs\_PreservesDocStructure
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 309
x-reqmd.symbol-kind: function
trace:
  - IMP-filter-FilterDocs-2f82f12
```

### IMP-filter-TestFilterDocs\_DoesNotMutateOriginal-6641d5c: TestFilterDocs\_DoesNotMutateOriginal
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 335
x-reqmd.symbol-kind: function
trace:
  - IMP-filter-FilterDocs-2f82f12
```

### IMP-filter-TestFilterDocs\_KeepsSectionsThatStillLeadSomewhere-295da3b: TestFilterDocs\_KeepsSectionsThatStillLeadSomewhere
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 354
x-reqmd.symbol-kind: function
trace:
  - IMP-filter-FilterDocs-2f82f12
```

### IMP-filter-TestFilterDocs\_DropsSectionThatLostAllChildren-5424ce1: TestFilterDocs\_DropsSectionThatLostAllChildren
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 395
x-reqmd.symbol-kind: function
trace:
  - IMP-filter-FilterDocs-2f82f12
```

A section that had children but lost them all to the filter is context for
content that is gone: it is dropped even when it carries prose.

### IMP-filter-TestFilterDocs\_KeepsLeafSectionsAndPreamble-8746b1e: TestFilterDocs\_KeepsLeafSectionsAndPreamble
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 421
x-reqmd.symbol-kind: function
trace:
  - IMP-filter-FilterDocs-2f82f12
```

A leaf section is content: it survives a filter even with no requirements in
sight. A rationale-only section counts as content too.

### IMP-filter-titles-4cf1316: titles
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 451
x-reqmd.symbol-kind: function
trace: []
```

titles is a debug helper for prune assertions.

### IMP-filter-TestFilterDocs\_ReattachesChildrenOfPrunedReq-c822721: TestFilterDocs\_ReattachesChildrenOfPrunedReq
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 459
x-reqmd.symbol-kind: function
trace:
  - IMP-filter-FilterDocs-2f82f12
```

### IMP-filter-TestFilterDocsAndIDs\_MatchesSeparateCalls-5a8d517: TestFilterDocsAndIDs\_MatchesSeparateCalls
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 485
x-reqmd.symbol-kind: function
trace:
  - IMP-filter-FilterDocsAndIDs-97e9fb7
```

FilterDocsAndIDs must return exactly what FilterDocs and MatchingIDs return
separately: the same pruned tree (nested sections, a container that only
partly matches, a pruned requirement whose child survives) and the same ID
set, from a single evaluation pass.

### IMP-filter-TestMatchingIDs-1de58f4: TestMatchingIDs
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 551
x-reqmd.symbol-kind: function
trace:
  - IMP-filter-MatchingIDs-7e11246
```

### IMP-filter-TestExprAndRefs-3707609: TestExprAndRefs
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 578
x-reqmd.symbol-kind: function
trace:
  - IMP-filter-Expr-2023150
```

### IMP-filter-TestBuildValidAttrs-d411a09: TestBuildValidAttrs
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 599
x-reqmd.symbol-kind: function
trace:
  - IMP-filter-BuildValidAttrs-4f190c9
```

### IMP-filter-undeclaredDocs-e3bdc27: undeclaredDocs
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 619
x-reqmd.symbol-kind: function
trace: []
```

undeclaredDocs is testDocs plus a requirement carrying an attribute that
no schema declares (`owner`), the relaxed-filtering scenario.

### IMP-filter-TestBuildFoundAttrs-57b7a6b: TestBuildFoundAttrs
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 628
x-reqmd.symbol-kind: function
trace:
  - IMP-filter-BuildFoundAttrs-157302b
```

### IMP-filter-TestCompileForDocs\_StrictRejectsUndeclared-ae86a23: TestCompileForDocs\_StrictRejectsUndeclared
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 640
x-reqmd.symbol-kind: function
trace:
  - IMP-filter-CompileForDocs-787afca
```

TestCompileForDocs_StrictRejectsUndeclared pins the default behavior: the
global typo check rejects attributes that no schema declares.

### IMP-filter-TestCompileForDocs\_RelaxedAllowsFoundAttr-6e25897: TestCompileForDocs\_RelaxedAllowsFoundAttr
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 650
x-reqmd.symbol-kind: function
trace:
  - IMP-filter-CompileForDocs-787afca
```

TestCompileForDocs_RelaxedAllowsFoundAttr is the relaxed-filter contract:
an attribute present on a requirement becomes filterable even though no
schema.yaml declares it, and the deviation is reported as a warning.

### IMP-filter-TestCompileForDocs\_RelaxedWarnsOnTrulyUnknown-d916b6a: TestCompileForDocs\_RelaxedWarnsOnTrulyUnknown
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 677
x-reqmd.symbol-kind: function
trace:
  - IMP-filter-CompileForDocs-787afca
```

TestCompileForDocs_RelaxedWarnsOnTrulyUnknown verifies relaxed mode turns
a genuine typo into a warning instead of aborting the run.

### IMP-filter-TestCompileForDocs\_RelaxedDeclaredAttrNoWarning-256c7a0: TestCompileForDocs\_RelaxedDeclaredAttrNoWarning
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 698
x-reqmd.symbol-kind: function
trace:
  - IMP-filter-CompileForDocs-787afca
```

TestCompileForDocs_RelaxedDeclaredAttrNoWarning keeps relaxed mode quiet
for expressions that only use declared attributes.

### IMP-filter-TestCompileForDocs\_EmptyExpression-61baade: TestCompileForDocs\_EmptyExpression
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 709
x-reqmd.symbol-kind: function
trace:
  - IMP-filter-CompileForDocs-787afca
```

TestCompileForDocs_EmptyExpression asserts the no-filter fast path.

