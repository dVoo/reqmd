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
x-reqmd.source-line: 174
x-reqmd.symbol-kind: method
trace: []
```

Expr returns the original expression string.

#### IMP-filter-Refs-477c375: Refs
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter.go
x-reqmd.source-line: 178
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
x-reqmd.source-line: 183
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
x-reqmd.source-line: 202
x-reqmd.symbol-kind: method
trace: []
```

FilterDocs returns a new []model.Document containing only matching
 requirements per document. Non-requirement nodes (containers and info
 items) are always kept, so document structure and content survive
 filtering. Documents with all requirements filtered out are kept (with
 only items) so document-level structure is preserved for list/stats/
 export output.

#### IMP-filter-MatchingIDs-7e11246: MatchingIDs
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter.go
x-reqmd.source-line: 219
x-reqmd.symbol-kind: method
trace: []
```

MatchingIDs returns a set of requirement IDs that match the filter.
 Used by check/serve for filter-aware graph checks — the graph is
 built from the full tree, but only matching reqs are checked/reported.

#### IMP-filter-pruneNodes-5c73a8e: pruneNodes
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter.go
x-reqmd.source-line: 250
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

### IMP-filter-Compile-a260df6: Compile
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter.go
x-reqmd.source-line: 85
x-reqmd.symbol-kind: function
trace: []
```

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
x-reqmd.source-line: 98
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
x-reqmd.source-line: 106
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
x-reqmd.source-line: 289
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
x-reqmd.source-line: 300
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
x-reqmd.source-line: 317
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
x-reqmd.source-line: 332
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
x-reqmd.source-line: 347
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
x-reqmd.source-line: 366
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
x-reqmd.source-line: 370
x-reqmd.symbol-kind: method
trace: []
```

## internal/filter/filter_test.go
### IMP-filter-testReq-328520b: testReq
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 11
x-reqmd.symbol-kind: function
trace: []
```

### IMP-filter-testDocs-202d8ea: testDocs
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 20
x-reqmd.symbol-kind: function
trace: []
```

### IMP-filter-validAttrsFor-d5cca1a: validAttrsFor
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 57
x-reqmd.symbol-kind: function
trace: []
```

### IMP-filter-TestCompileAndMatch_Equality-8fffca1: TestCompileAndMatch_Equality
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 61
x-reqmd.symbol-kind: function
trace: []
```

### IMP-filter-TestCompileAndMatch_ArrayIn-1c93a59: TestCompileAndMatch_ArrayIn
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 85
x-reqmd.symbol-kind: function
trace: []
```

### IMP-filter-TestCompileAndMatch_NilAttribute-6d118bc: TestCompileAndMatch_NilAttribute
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 109
x-reqmd.symbol-kind: function
trace: []
```

### IMP-filter-TestCompileAndMatch_NilCoalescing-1f25f20: TestCompileAndMatch_NilCoalescing
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 126
x-reqmd.symbol-kind: function
trace: []
```

### IMP-filter-TestCompileAndMatch_NumericComparison-12893e9: TestCompileAndMatch_NumericComparison
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 162
x-reqmd.symbol-kind: function
trace: []
```

### IMP-filter-TestCompileAndMatch_BuiltInIDAndTitle-d27bc68: TestCompileAndMatch_BuiltInIDAndTitle
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 186
x-reqmd.symbol-kind: function
trace: []
```

### IMP-filter-TestCompileAndMatch_BooleanComposition-3fe2287: TestCompileAndMatch_BooleanComposition
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 210
x-reqmd.symbol-kind: function
trace: []
```

### IMP-filter-TestCompile_SyntaxError-8019356: TestCompile_SyntaxError
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 234
x-reqmd.symbol-kind: function
trace: []
```

### IMP-filter-TestCompile_UnknownAttribute-407e2d1: TestCompile_UnknownAttribute
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 242
x-reqmd.symbol-kind: function
trace: []
```

### IMP-filter-TestCompile_EmptyExpression-1ae7b0a: TestCompile_EmptyExpression
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 250
x-reqmd.symbol-kind: function
trace: []
```

### IMP-filter-TestCompile_AllowsExprBuiltins-71f36c6: TestCompile_AllowsExprBuiltins
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 257
x-reqmd.symbol-kind: function
trace: []
```

### IMP-filter-TestCompile_AllowsCrossSchemaAttribute-185a5b4: TestCompile_AllowsCrossSchemaAttribute
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 267
x-reqmd.symbol-kind: function
trace: []
```

### IMP-filter-TestFilterDocs-d2134a8: TestFilterDocs
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 277
x-reqmd.symbol-kind: function
trace: []
```

### IMP-filter-TestFilterDocs_PreservesDocStructure-0485295: TestFilterDocs_PreservesDocStructure
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 307
x-reqmd.symbol-kind: function
trace: []
```

### IMP-filter-TestFilterDocs_DoesNotMutateOriginal-6641d5c: TestFilterDocs_DoesNotMutateOriginal
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 333
x-reqmd.symbol-kind: function
trace: []
```

### IMP-filter-TestFilterDocs_KeepsSectionsThatStillLeadSomewhere-295da3b: TestFilterDocs_KeepsSectionsThatStillLeadSomewhere
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 352
x-reqmd.symbol-kind: function
trace: []
```

### IMP-filter-TestFilterDocs_DropsSectionThatLostAllChildren-5424ce1: TestFilterDocs_DropsSectionThatLostAllChildren
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 393
x-reqmd.symbol-kind: function
trace: []
```

A section that had children but lost them all to the filter is context for
 content that is gone: it is dropped even when it carries prose.

### IMP-filter-TestFilterDocs_KeepsLeafSectionsAndPreamble-8746b1e: TestFilterDocs_KeepsLeafSectionsAndPreamble
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 419
x-reqmd.symbol-kind: function
trace: []
```

A leaf section is content: it survives a filter even with no requirements in
 sight. A rationale-only section counts as content too.

### IMP-filter-titles-4cf1316: titles
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 449
x-reqmd.symbol-kind: function
trace: []
```

titles is a debug helper for prune assertions.

### IMP-filter-TestFilterDocs_ReattachesChildrenOfPrunedReq-c822721: TestFilterDocs_ReattachesChildrenOfPrunedReq
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 457
x-reqmd.symbol-kind: function
trace: []
```

### IMP-filter-TestMatchingIDs-1de58f4: TestMatchingIDs
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 479
x-reqmd.symbol-kind: function
trace: []
```

### IMP-filter-TestExprAndRefs-3707609: TestExprAndRefs
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 506
x-reqmd.symbol-kind: function
trace: []
```

### IMP-filter-TestBuildValidAttrs-d411a09: TestBuildValidAttrs
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 527
x-reqmd.symbol-kind: function
trace: []
```

### IMP-filter-undeclaredDocs-e3bdc27: undeclaredDocs
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 547
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
x-reqmd.source-line: 556
x-reqmd.symbol-kind: function
trace: []
```

### IMP-filter-TestCompileForDocs_StrictRejectsUndeclared-ae86a23: TestCompileForDocs_StrictRejectsUndeclared
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 568
x-reqmd.symbol-kind: function
trace: []
```

TestCompileForDocs_StrictRejectsUndeclared pins the default behavior: the
 global typo check rejects attributes that no schema declares.

### IMP-filter-TestCompileForDocs_RelaxedAllowsFoundAttr-6e25897: TestCompileForDocs_RelaxedAllowsFoundAttr
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 578
x-reqmd.symbol-kind: function
trace: []
```

TestCompileForDocs_RelaxedAllowsFoundAttr is the relaxed-filter contract:
 an attribute present on a requirement becomes filterable even though no
 schema.yaml declares it, and the deviation is reported as a warning.

### IMP-filter-TestCompileForDocs_RelaxedWarnsOnTrulyUnknown-d916b6a: TestCompileForDocs_RelaxedWarnsOnTrulyUnknown
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 605
x-reqmd.symbol-kind: function
trace: []
```

TestCompileForDocs_RelaxedWarnsOnTrulyUnknown verifies relaxed mode turns
 a genuine typo into a warning instead of aborting the run.

### IMP-filter-TestCompileForDocs_RelaxedDeclaredAttrNoWarning-256c7a0: TestCompileForDocs_RelaxedDeclaredAttrNoWarning
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 626
x-reqmd.symbol-kind: function
trace: []
```

TestCompileForDocs_RelaxedDeclaredAttrNoWarning keeps relaxed mode quiet
 for expressions that only use declared attributes.

### IMP-filter-TestCompileForDocs_EmptyExpression-61baade: TestCompileForDocs_EmptyExpression
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/filter/filter_test.go
x-reqmd.source-line: 637
x-reqmd.symbol-kind: function
trace: []
```

TestCompileForDocs_EmptyExpression asserts the no-filter fast path.

