# Package: lang

<!-- reqmd: generated -->

## reqmd-code/internal/lang/helpers.go
### IMP-lang-IsTestFile-68ae539: IsTestFile
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/helpers.go
x-reqmd.source-line: 24
x-reqmd.symbol-kind: function
trace: []
```

IsTestFile reports whether path follows one of the supported languages'
test-file conventions. It is language-agnostic on purpose: a resolver that
sees only a symbol should not need to know which plugin produced it.

	Go         *_test.go
	Python     test_*.py, *_test.py
	Rust       tests/*.rs, *_test.rs
	C/C++      test_*.c/.cc/.cpp/.cxx, *_test.c/.cc/.cpp/.cxx

A new language adds its convention here, and can still mark inline tests
through model.Symbol.IsTest.

### IMP-lang-HasAncestorKind-4f7ba1a: HasAncestorKind
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/helpers.go
x-reqmd.source-line: 47
x-reqmd.symbol-kind: function
trace: []
```

HasAncestorKind reports whether any ancestor of node has one of the given
tree-sitter node kinds. Every plugin uses it to reject declarations local
to a function/class body rather than top-level symbols; variadic because
some languages spell the same scope several ways (Go's function, method
and func-literal).

### IMP-lang-CapturesByName-4d94612: CapturesByName
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/helpers.go
x-reqmd.source-line: 64
x-reqmd.symbol-kind: function
trace: []
```

CapturesByName maps capture name to node for a single query match.
Capture names are resolved through the query's CaptureNames() slice, so
the caller fetches that once per query and passes it in.

### IMP-lang-ModuleNameFromFile-8851a86: ModuleNameFromFile
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/helpers.go
x-reqmd.source-line: 74
x-reqmd.symbol-kind: function
trace: []
```

ModuleNameFromFile returns the file's basename without its extension,
the default package/module name for languages whose package is the file.

### IMP-lang-CaptureContent-447b610: CaptureContent
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/helpers.go
x-reqmd.source-line: 104
x-reqmd.symbol-kind: function
trace: []
```

CaptureContent returns the textual content of a captured node, or the
empty string when the node is nil. It exists as a shared helper
because every language plugin needs the same nil-guarded access to
tree-sitter node text; previously each plugin reimplemented this
3-liner.

### IMP-lang-SortByStartLine-a1bbfc8: SortByStartLine
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/helpers.go
x-reqmd.source-line: 114
x-reqmd.symbol-kind: function
trace: []
```

SortByStartLine sorts symbols in place by StartLine ascending.
Hoisted because every plugin's Parse returns symbols in source
order but wants a defensive re-sort as the last step.

### IMP-lang-SymbolAt-2492232: SymbolAt
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/helpers.go
x-reqmd.source-line: 123
x-reqmd.symbol-kind: function
trace: []
```

SymbolAt fills in the span and doc fields for a single symbol. The
language parameter is the interface (not the concrete plugin type)
so this helper is reusable across plugins: each plugin's only
dependency on its own type was the call to BindDoc, which is part
of the Language contract.

## reqmd-code/internal/lang/registry.go
### IMP-lang-Language-a0504b1: Language
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/registry.go
x-reqmd.source-line: 26
x-reqmd.symbol-kind: type
trace: []
```

Language is the contract a source-language plugin must satisfy.

Implementations are expected to be safe to call concurrently: the
registry caches a single instance per language and the extractor
dispatches jobs to language plugins in parallel.

### IMP-lang-registry-70620c8: registry
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/registry.go
x-reqmd.source-line: 62
x-reqmd.symbol-kind: type
trace: []
```

Registry is the global, append-only set of registered Language plugins.
It is safe for concurrent reads after the init() phase; concurrent
Register calls during init() are protected by a mutex.

### IMP-lang-globalRegistry-011cf43: globalRegistry
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/registry.go
x-reqmd.source-line: 67
x-reqmd.symbol-kind: var
trace: []
```

### IMP-lang-Register-dd7e002: Register
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/registry.go
x-reqmd.source-line: 74
x-reqmd.symbol-kind: function
trace: []
```

Register adds a Language to the global registry. Typically called from
a plugin's init() function. Duplicate registrations panic — they
indicate a programming error (two packages claiming the same language).

### IMP-lang-Get-795ebc9: Get
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/registry.go
x-reqmd.source-line: 88
x-reqmd.symbol-kind: function
trace: []
```

Get fetches a registered Language by name. Returns (nil, false) if
the language is not registered.

### IMP-lang-All-37b991a: All
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/registry.go
x-reqmd.source-line: 98
x-reqmd.symbol-kind: function
trace: []
```

All returns every registered language, sorted by name for deterministic
output. Used by the --lang flag validation and by the
extractor's language auto-detection pass.

### IMP-lang-ForExtension-6733a2b: ForExtension
```yaml reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/registry.go
x-reqmd.source-line: 112
x-reqmd.symbol-kind: function
trace: []
```

ForExtension returns the first Language that claims the given file
extension. Returns (nil, false) if no plugin handles the extension.
Comparison is exact-match on the extensions list (no normalization).

