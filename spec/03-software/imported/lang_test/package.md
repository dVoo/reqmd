# Package: lang_test

<!-- reqmd: generated -->

## reqmd-code/internal/lang/helpers_test.go
### IMP-lang_test-TestModuleNameFromFile-ef68781: TestModuleNameFromFile
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/helpers_test.go
x-reqmd.source-line: 15
x-reqmd.symbol-kind: function
trace: []
```

TestModuleNameFromFile covers the helper every plugin uses to derive a
 default package name from the source path (one shared table instead of one
 per language).

### IMP-lang_test-TestHasAncestorKind-22d86e7: TestHasAncestorKind
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/helpers_test.go
x-reqmd.source-line: 34
x-reqmd.symbol-kind: function
trace: []
```

TestHasAncestorKind covers the shared ancestor walk: the nil guard, a
 matching ancestor, and one of several candidate kinds.

### IMP-lang_test-walk-9689cad: walk
```attr
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/helpers_test.go
x-reqmd.source-line: 72
x-reqmd.symbol-kind: function
trace: []
```

walk visits every node depth-first.

