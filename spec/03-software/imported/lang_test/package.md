# Package: lang\_test

<!-- reqmd: generated -->

## reqmd-code/internal/lang/helpers\_test.go
### IMP-lang\_test-TestModuleNameFromFile-ef68781: TestModuleNameFromFile
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/helpers_test.go
x-reqmd.source-line: 15
x-reqmd.symbol-kind: function
trace:
  - lang/IMP-lang-ModuleNameFromFile-8851a86
```

TestModuleNameFromFile covers the helper every plugin uses to derive a
default package name from the source path (one shared table instead of one
per language).

### IMP-lang\_test-TestHasAncestorKind-22d86e7: TestHasAncestorKind
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/helpers_test.go
x-reqmd.source-line: 38
x-reqmd.symbol-kind: function
trace:
  - lang/IMP-lang-HasAncestorKind-4f7ba1a
```

TestHasAncestorKind covers the shared ancestor walk: the nil guard, a
matching ancestor, and one of several candidate kinds.

### IMP-lang\_test-walk-9689cad: walk
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/helpers_test.go
x-reqmd.source-line: 76
x-reqmd.symbol-kind: function
trace: []
```

walk visits every node depth-first.

### IMP-lang\_test-TestIsTestFile-e0e6742: TestIsTestFile
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/helpers_test.go
x-reqmd.source-line: 84
x-reqmd.symbol-kind: function
trace:
  - lang/IMP-lang-IsTestFile-68ae539
```

TestIsTestFile pins the file-name conventions used by convention linking.

