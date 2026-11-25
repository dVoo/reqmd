# Package: diff

<!-- reqmd: generated -->

## internal/diff/diff.go
### IMP-diff-Category-246a6d8: Category
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/diff/diff.go
x-reqmd.source-line: 16
x-reqmd.symbol-kind: type
trace: []
```

Category describes what happened to a requirement between baselines.

### IMP-diff-Added-d5adc62: Added
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/diff/diff.go
x-reqmd.source-line: 19
x-reqmd.symbol-kind: const
trace: []
```

Requirement diff categories.

### IMP-diff-Modified-0e57aff: Modified
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/diff/diff.go
x-reqmd.source-line: 19
x-reqmd.symbol-kind: const
trace: []
```

Requirement diff categories.

### IMP-diff-Removed-f1267a4: Removed
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/diff/diff.go
x-reqmd.source-line: 19
x-reqmd.symbol-kind: const
trace: []
```

Requirement diff categories.

### IMP-diff-Unchanged-a14e729: Unchanged
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/diff/diff.go
x-reqmd.source-line: 19
x-reqmd.symbol-kind: const
trace: []
```

Requirement diff categories.

### IMP-diff-AttrChange-3bd6b27: AttrChange
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/diff/diff.go
x-reqmd.source-line: 27
x-reqmd.symbol-kind: type
trace: []
```

AttrChange records a single attribute change within a requirement.

### IMP-diff-ReqDiff-2a5fcbf: ReqDiff
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/diff/diff.go
x-reqmd.source-line: 34
x-reqmd.symbol-kind: type
trace: []
```

ReqDiff is the diff result for a single requirement.

### IMP-diff-DocDiff-91669bd: DocDiff
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/diff/diff.go
x-reqmd.source-line: 44
x-reqmd.symbol-kind: type
trace: []
```

DocDiff is the diff result for a single document directory.

### IMP-diff-SchemaChange-72ec63f: SchemaChange
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/diff/diff.go
x-reqmd.source-line: 54
x-reqmd.symbol-kind: type
trace: []
```

SchemaChange describes a single property change in a schema.

### IMP-diff-SchemaDiff-04372e8: SchemaDiff
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/diff/diff.go
x-reqmd.source-line: 61
x-reqmd.symbol-kind: type
trace: []
```

SchemaDiff is the diff result for a single document's schema.

### IMP-diff-SubmoduleChange-4b899ca: SubmoduleChange
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/diff/diff.go
x-reqmd.source-line: 69
x-reqmd.symbol-kind: type
trace: []
```

SubmoduleChange describes a submodule's pinned-commit change between two baselines.

### IMP-diff-Result-8016ed9: Result
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/diff/diff.go
x-reqmd.source-line: 77
x-reqmd.symbol-kind: type
trace: []
```

Result is the top-level diff result between two baselines.

### IMP-diff-Diff-3bcbf07: Diff
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/diff/diff.go
x-reqmd.source-line: 91
x-reqmd.symbol-kind: function
trace: []
```

Diff compares two sets of parsed documents and returns structured diff results.
docs1 is the older baseline (tag1), docs2 is the newer baseline (tag2).

### IMP-diff-Schemas-e4571f8: Schemas
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/diff/diff.go
x-reqmd.source-line: 197
x-reqmd.symbol-kind: function
trace: []
```

Schemas compares two sets of schemas and returns per-document schema diffs.

### IMP-diff-SchemasWithDiffer-a5b6a4c: SchemasWithDiffer
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/diff/diff.go
x-reqmd.source-line: 203
x-reqmd.symbol-kind: function
trace: []
```

SchemasWithDiffer compares two sets of schemas using the provided differ.

### IMP-diff-topLevelKeys-e882203: topLevelKeys
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/diff/diff.go
x-reqmd.source-line: 254
x-reqmd.symbol-kind: function
trace: []
```

### IMP-diff-buildReqMap-300da3e: buildReqMap
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/diff/diff.go
x-reqmd.source-line: 264
x-reqmd.symbol-kind: function
trace: []
```

buildReqMap groups requirements by document path.

### IMP-diff-buildSectionPaths-e1bf1c9: buildSectionPaths
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/diff/diff.go
x-reqmd.source-line: 280
x-reqmd.symbol-kind: function
trace: []
```

buildSectionPaths maps document path -> requirement ID -> the joined titles of
its ancestor sections. Sections are context, not content: a requirement moving
between chapters is a real review event even though no attribute changed, so
the diff reports it as a "parent" delta on the requirement.

### IMP-diff-diffAttrs-785aac9: diffAttrs
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/diff/diff.go
x-reqmd.source-line: 289
x-reqmd.symbol-kind: function
trace: []
```

diffAttrs compares two attribute maps using r3labs/diff and returns changes.

### IMP-diff-diffAttrsManual-d6fb6e7: diffAttrsManual
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/diff/diff.go
x-reqmd.source-line: 316
x-reqmd.symbol-kind: function
trace: []
```

diffAttrsManual is a fallback for when r3labs can't handle the types.

### IMP-diff-Submodules-15c6d17: Submodules
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/diff/diff.go
x-reqmd.source-line: 349
x-reqmd.symbol-kind: function
trace: []
```

Submodules compares two maps of submodule path → commit SHA and returns
a sorted list of SubmoduleChange. Submodules with the same SHA at both
tags are omitted (no noise). Result is sorted by path for deterministic output.

## internal/diff/diff\_test.go
### IMP-diff-TestDiff\_Empty-e01da55: TestDiff\_Empty
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/diff/diff_test.go
x-reqmd.source-line: 9
x-reqmd.symbol-kind: function
trace:
  - IMP-diff-Diff-3bcbf07
```

### IMP-diff-TestDiff\_Added-e11c4ce: TestDiff\_Added
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/diff/diff_test.go
x-reqmd.source-line: 27
x-reqmd.symbol-kind: function
trace:
  - IMP-diff-Diff-3bcbf07
```

### IMP-diff-TestDiff\_Removed-4b2a500: TestDiff\_Removed
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/diff/diff_test.go
x-reqmd.source-line: 56
x-reqmd.symbol-kind: function
trace:
  - IMP-diff-Diff-3bcbf07
```

### IMP-diff-TestDiff\_Modified-01c3983: TestDiff\_Modified
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/diff/diff_test.go
x-reqmd.source-line: 82
x-reqmd.symbol-kind: function
trace:
  - IMP-diff-Diff-3bcbf07
```

### IMP-diff-TestDiff\_Unchanged-af0c165: TestDiff\_Unchanged
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/diff/diff_test.go
x-reqmd.source-line: 128
x-reqmd.symbol-kind: function
trace:
  - IMP-diff-Diff-3bcbf07
```

### IMP-diff-TestDiff\_Mixed-32f4ddb: TestDiff\_Mixed
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/diff/diff_test.go
x-reqmd.source-line: 154
x-reqmd.symbol-kind: function
trace:
  - IMP-diff-Diff-3bcbf07
```

### IMP-diff-TestDiff\_SectionMove-801595a: TestDiff\_SectionMove
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/diff/diff_test.go
x-reqmd.source-line: 211
x-reqmd.symbol-kind: function
trace:
  - IMP-diff-Diff-3bcbf07
```

TestDiff_SectionMove verifies that moving a requirement to another chapter is
reported as a modification with a "parent" delta, even though no attribute
changed: sections are context, and a move changes what the requirement is
about.

### IMP-diff-TestSchemas\_AddedProperty-421be34: TestSchemas\_AddedProperty
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/diff/diff_test.go
x-reqmd.source-line: 254
x-reqmd.symbol-kind: function
trace:
  - IMP-diff-Schemas-e4571f8
```

### IMP-diff-TestSchemas\_RemovedProperty-b9174de: TestSchemas\_RemovedProperty
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/diff/diff_test.go
x-reqmd.source-line: 270
x-reqmd.symbol-kind: function
trace:
  - IMP-diff-Schemas-e4571f8
```

### IMP-diff-TestSchemas\_ChangedRequired-32e188a: TestSchemas\_ChangedRequired
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/diff/diff_test.go
x-reqmd.source-line: 283
x-reqmd.symbol-kind: function
trace:
  - IMP-diff-Schemas-e4571f8
```

### IMP-diff-TestSchemas\_NoChange-f0d8882: TestSchemas\_NoChange
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/diff/diff_test.go
x-reqmd.source-line: 296
x-reqmd.symbol-kind: function
trace:
  - IMP-diff-Schemas-e4571f8
```

### IMP-diff-TestSubmodules\_Empty-2f1b385: TestSubmodules\_Empty
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/diff/diff_test.go
x-reqmd.source-line: 309
x-reqmd.symbol-kind: function
trace:
  - IMP-diff-Submodules-15c6d17
```

### IMP-diff-TestSubmodules\_NilMaps-92fcc28: TestSubmodules\_NilMaps
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/diff/diff_test.go
x-reqmd.source-line: 316
x-reqmd.symbol-kind: function
trace:
  - IMP-diff-Submodules-15c6d17
```

### IMP-diff-TestSubmodules\_NoChange-e14c5e7: TestSubmodules\_NoChange
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/diff/diff_test.go
x-reqmd.source-line: 323
x-reqmd.symbol-kind: function
trace:
  - IMP-diff-Submodules-15c6d17
```

### IMP-diff-TestSubmodules\_Added-d05731e: TestSubmodules\_Added
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/diff/diff_test.go
x-reqmd.source-line: 334
x-reqmd.symbol-kind: function
trace:
  - IMP-diff-Submodules-15c6d17
```

### IMP-diff-TestSubmodules\_Removed-78d4a10: TestSubmodules\_Removed
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/diff/diff_test.go
x-reqmd.source-line: 360
x-reqmd.symbol-kind: function
trace:
  - IMP-diff-Submodules-15c6d17
```

### IMP-diff-TestSubmodules\_Updated-c244b93: TestSubmodules\_Updated
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/diff/diff_test.go
x-reqmd.source-line: 386
x-reqmd.symbol-kind: function
trace:
  - IMP-diff-Submodules-15c6d17
```

### IMP-diff-TestSubmodules\_Mixed-d9985cb: TestSubmodules\_Mixed
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/diff/diff_test.go
x-reqmd.source-line: 411
x-reqmd.symbol-kind: function
trace:
  - IMP-diff-Submodules-15c6d17
```

### IMP-diff-TestSubmodules\_SortedOutput-1e3d7e8: TestSubmodules\_SortedOutput
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: internal/diff/diff_test.go
x-reqmd.source-line: 439
x-reqmd.symbol-kind: function
trace:
  - IMP-diff-Submodules-15c6d17
```

