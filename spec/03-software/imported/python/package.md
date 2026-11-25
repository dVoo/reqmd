# Package: python

<!-- reqmd: generated -->

## reqmd-code/internal/lang/python/python.go
### IMP-python-queriesFS-c8ef16c: queriesFS
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/python/python.go
x-reqmd.source-line: 40
x-reqmd.symbol-kind: var
trace: []
```

go:embed queries/*.scm

### IMP-python-PythonLanguage-78bcb41: PythonLanguage
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/python/python.go
x-reqmd.source-line: 43
x-reqmd.symbol-kind: type
trace: []
```

PythonLanguage is the Python language plugin. Implements lang.Language.

#### IMP-python-Name-c97ec76: Name
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/python/python.go
x-reqmd.source-line: 49
x-reqmd.symbol-kind: method
trace: []
```

Name returns the canonical language identifier.

#### IMP-python-Extensions-0f80cdb: Extensions
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/python/python.go
x-reqmd.source-line: 52
x-reqmd.symbol-kind: method
trace: []
```

Extensions returns the file extensions this plugin handles.

#### IMP-python-Grammar-9f8e02e: Grammar
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/python/python.go
x-reqmd.source-line: 55
x-reqmd.symbol-kind: method
trace: []
```

Grammar returns the tree-sitter-python grammar language.

#### IMP-python-Query-cba0153: Query
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/python/python.go
x-reqmd.source-line: 62
x-reqmd.symbol-kind: method
trace: []
```

Query returns the embedded tree-sitter query for symbol extraction.
If the query file is missing (a packaging error), it returns an empty
query rather than panicking; Parse will then emit zero symbols.

#### IMP-python-BindDoc-d920b70: BindDoc
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/python/python.go
x-reqmd.source-line: 75
x-reqmd.symbol-kind: method
trace: []
```

BindDoc returns the cleaned doc for the given declaration node. It
joins the leading `#` line comments with the function/class
docstring (if any) using a blank-line separator. The trace
extractor regex is line-based, so the order of the two halves is
immaterial. Returns the empty string when neither is present.

#### IMP-python-Parse-3dc8efc: Parse
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/python/python.go
x-reqmd.source-line: 260
x-reqmd.symbol-kind: method
trace: []
```

Parse parses the given Python source bytes and returns a flat list
of top-level symbols, in source order. It uses the embedded
tree-sitter query to find function, method, class, and module-level
constant declarations, and walks the parent chain to drop class-body
constants and to populate a method's Receiver from its enclosing
class_definition's `name` field.

##### Dedup strategy
The query produces intentional double-matches for the same source
node: a method's `function_definition` matches both the bare
`function_definition` pattern (@func.decl) and the class-body-anchored
pattern (@method.decl); a decorated class matches both the bare
`class_definition` pattern and the `decorated_definition` unwrap.
To collapse these to a single emitted symbol we track emitted decl
nodes by `*sitter.Node` identity. The pass is split into two phases
so that @method.decl wins over @func.decl when both point at the
same node (methods take precedence over plain functions).

### IMP-python-cleanPythonComment-49cc265: cleanPythonComment
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/python/python.go
x-reqmd.source-line: 121
x-reqmd.symbol-kind: function
trace: []
```

cleanPythonComment strips the leading `#` and trims trailing
whitespace per line. Python's comment grammar produces a single
node spanning one line (`# foo`), so this is a thin per-line pass.

### IMP-python-extractDocstring-598a161: extractDocstring
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/python/python.go
x-reqmd.source-line: 152
x-reqmd.symbol-kind: function
trace: []
```

extractDocstring returns the cleaned docstring of a
function_definition or class_definition node, or "" when the node
has no docstring. The tree-sitter-python shape is:

	function_definition
	  body: (block
	    (expression_statement
	      (string)))

The string node's text includes the surrounding `"""`/`”'`
delimiters (and an optional `r`/`b`/`u`/`rb`/... prefix). We strip
the prefix and delimiters, then trim each line. Any unexpected AST
shape returns "" rather than panicking.

### IMP-python-stripDocstringDelimiters-335b922: stripDocstringDelimiters
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/python/python.go
x-reqmd.source-line: 186
x-reqmd.symbol-kind: function
trace: []
```

stripDocstringDelimiters removes the leading prefix letters (r, b,
u, rb, br, f, rf, fr, in any case-insensitive combination) and
the surrounding triple-quote delimiters (`"""` or `”'`) from a
docstring literal, then trims each line. Returns "" if the input
is too short to contain the delimiters.

### IMP-python-symbolFromMatch-36ab32a: symbolFromMatch
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/python/python.go
x-reqmd.source-line: 361
x-reqmd.symbol-kind: function
trace: []
```

symbolFromMatch builds a model.Symbol from the per-match capture map.
The map keys are the @-names from queries/symbols.scm. Returns false
when the match has no usable captures.

### IMP-python-primaryDeclNode-17bcd09: primaryDeclNode
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/python/python.go
x-reqmd.source-line: 393
x-reqmd.symbol-kind: function
trace: []
```

primaryDeclNode returns the captured "decl" node for the match's
dispatch category, or nil when the match carries no primary decl.

### IMP-python-enclosingClassName-d0cc098: enclosingClassName
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/python/python.go
x-reqmd.source-line: 414
x-reqmd.symbol-kind: function
trace: []
```

enclosingClassName walks up the parent chain from a method's
function_definition node to the enclosing class_definition and
returns the class's name text (e.g. "Calculator" for
`class Calculator: def multiply(self, ...)`). Returns "" when the
method has no enclosing class (defensive — the query should not
produce such a match) or when the class has no name field (also
defensive).

### IMP-python-init-c358feb: init
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/python/python.go
x-reqmd.source-line: 432
x-reqmd.symbol-kind: function
trace: []
```

## reqmd-code/internal/lang/python/python\_test.go
### IMP-python-TestPythonLanguageRegisters-a4c10ad: TestPythonLanguageRegisters
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/python/python_test.go
x-reqmd.source-line: 17
x-reqmd.symbol-kind: function
trace:
  - IMP-python-PythonLanguage-78bcb41
```

TestPythonLanguageRegisters verifies that PythonLanguage self-registered
itself with the global lang.Registry on package init, and that its
Name/Extensions surface the expected values.

### IMP-python-TestPythonGrammarLoads-b58999d: TestPythonGrammarLoads
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/python/python_test.go
x-reqmd.source-line: 34
x-reqmd.symbol-kind: function
trace: []
```

TestPythonGrammarLoads verifies the tree-sitter-python binding is
reachable and returns a non-nil grammar. This is a smoke test —
the binding is cgo and a nil return would indicate a broken FFI
link or a missing grammar dynamic library.

### IMP-python-TestPythonQuery-f454545: TestPythonQuery
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/python/python_test.go
x-reqmd.source-line: 46
x-reqmd.symbol-kind: function
trace: []
```

TestPythonQuery verifies the embedded .scm query file is present and
readable. Returns true if the query text is non-empty (Fixer B's
contribution); the skeleton ships a placeholder so this may pass
with a one-line comment.

### IMP-python-TestCleanPythonComment-15afc45: TestCleanPythonComment
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/python/python_test.go
x-reqmd.source-line: 56
x-reqmd.symbol-kind: function
trace:
  - IMP-python-cleanPythonComment-49cc265
```

TestCleanPythonComment verifies the per-line # prefix stripping
used by BindDoc for module-level # comments.

### IMP-python-TestBindDocNilNode-8a79417: TestBindDocNilNode
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/python/python_test.go
x-reqmd.source-line: 74
x-reqmd.symbol-kind: function
trace:
  - IMP-python-BindDoc-d920b70
```

TestBindDocNilNode is a defensive guard: passing a nil node must
return the empty string rather than panicking.

### IMP-python-TestParseSimple-9fec2e6: TestParseSimple
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/python/python_test.go
x-reqmd.source-line: 85
x-reqmd.symbol-kind: function
trace:
  - IMP-python-Parse-3dc8efc
```

TestParseSimple exercises the full Parse pipeline against a small
Python snippet with one top-level function, one class, and one
method. Verifies symbol count, kinds, and the method's Receiver
(populated from the enclosing class_definition's name).

### IMP-python-TestParseClassLevelConstRejected-b350f83: TestParseClassLevelConstRejected
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/python/python_test.go
x-reqmd.source-line: 144
x-reqmd.symbol-kind: function
trace:
  - IMP-python-Parse-3dc8efc
```

TestParseClassLevelConstRejected verifies that a class-level
identifier assignment (`Cls.CONST = 99`) does NOT emit a
module-level constant symbol. The query's @const.decl pattern
matches the expression_statement regardless of lexical context; the
Go-side parent-chain walk must drop class-body matches.

### IMP-python-TestBindDocAboveDecorator-165622a: TestBindDocAboveDecorator
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/python/python_test.go
x-reqmd.source-line: 180
x-reqmd.symbol-kind: function
trace:
  - IMP-python-BindDoc-d920b70
```

TestBindDocAboveDecorator verifies that a `#` comment above a decorator
binds to the decorated function (the query captures the inner node, so
the walk must start at the decorated_definition wrapper).

### IMP-python-TestParseDecoratedClassDedup-5790775: TestParseDecoratedClassDedup
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/python/python_test.go
x-reqmd.source-line: 206
x-reqmd.symbol-kind: function
trace:
  - IMP-python-Parse-3dc8efc
```

TestParseDecoratedClassDedup verifies that a decorated class
(`@decorator\nclass Foo: ...`) produces exactly one type symbol,
not two. The query matches both the bare `class_definition` pattern
and the `decorated_definition` unwrap; the node-identity dedup must
collapse them.

### IMP-python-parseFirstFuncDef-50667f3: parseFirstFuncDef
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/python/python_test.go
x-reqmd.source-line: 230
x-reqmd.symbol-kind: function
trace: []
```

parseFirstFuncDef is a test helper that parses src and returns
the first top-level function_definition node plus a cleanup func
that must be called to release the underlying tree-sitter tree.
Tests need direct AST access to call extractDocstring/BindDoc on
a specific node; the public Parse path is exercised elsewhere.

### IMP-python-TestExtractDocstringTripleQuote-a0d35e3: TestExtractDocstringTripleQuote
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/python/python_test.go
x-reqmd.source-line: 262
x-reqmd.symbol-kind: function
trace:
  - IMP-python-extractDocstring-598a161
```

TestExtractDocstringTripleQuote verifies that a single-line
`"""..."""` docstring is captured with delimiters stripped and
per-line whitespace trimmed.

### IMP-python-TestExtractDocstringMultiLine-26cf0ac: TestExtractDocstringMultiLine
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/python/python_test.go
x-reqmd.source-line: 279
x-reqmd.symbol-kind: function
trace:
  - IMP-python-extractDocstring-598a161
```

TestExtractDocstringMultiLine verifies that a multi-line
`"""..."""` docstring is captured with delimiters stripped, the
leading newline trimmed, and per-line whitespace trimmed.

### IMP-python-TestExtractDocstringSingleTripleQuote-ceec747: TestExtractDocstringSingleTripleQuote
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/python/python_test.go
x-reqmd.source-line: 298
x-reqmd.symbol-kind: function
trace:
  - IMP-python-extractDocstring-598a161
```

TestExtractDocstringSingleTripleQuote verifies that `”'...”'`
(single-quote triple) docstrings are also captured.

### IMP-python-TestExtractDocstringNoDocstring-ffb7811: TestExtractDocstringNoDocstring
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/python/python_test.go
x-reqmd.source-line: 314
x-reqmd.symbol-kind: function
trace:
  - IMP-python-extractDocstring-598a161
```

TestExtractDocstringNoDocstring verifies that a function with no
docstring returns the empty string.

### IMP-python-TestExtractDocstringRawPrefix-cfae201: TestExtractDocstringRawPrefix
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/python/python_test.go
x-reqmd.source-line: 329
x-reqmd.symbol-kind: function
trace:
  - IMP-python-extractDocstring-598a161
```

TestExtractDocstringRawPrefix verifies that raw-string (`r"""..."""`)
docstrings have the `r` prefix handled gracefully. Tree-sitter's
source buffer includes the prefix, and our scanner must skip it
before locating the triple-quote delimiters.

### IMP-python-TestBindDocIncludesDocstring-7d92880: TestBindDocIncludesDocstring
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/python/python_test.go
x-reqmd.source-line: 348
x-reqmd.symbol-kind: function
trace:
  - IMP-python-BindDoc-d920b70
```

TestBindDocIncludesDocstring verifies that BindDoc returns BOTH
the leading `#` comments AND the docstring, joined with a blank
line. The trace extractor regex is line-based, so explicit
`reqmd:trace` markers in the leading comments and bare-ID
references in the docstring are both picked up.

### IMP-python-TestBindDocDocstringOnly-2afde69: TestBindDocDocstringOnly
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/python/python_test.go
x-reqmd.source-line: 369
x-reqmd.symbol-kind: function
trace:
  - IMP-python-BindDoc-d920b70
```

TestBindDocDocstringOnly verifies that BindDoc returns the
docstring when there are no leading `#` comments. This is the
case that previously returned "" and caused the `multiply`
heuristic-trace test to fail.

### IMP-python-TestParseFunctionLocalConstRejected-3d5a3d0: TestParseFunctionLocalConstRejected
```reqmd
status: approved
x-reqmd.imported: true
x-reqmd.source-file: reqmd-code/internal/lang/python/python_test.go
x-reqmd.source-line: 386
x-reqmd.symbol-kind: function
trace:
  - IMP-python-Parse-3dc8efc
```

TestParseFunctionLocalConstRejected pins that only module-level bindings are
constants: an assignment inside a function body is a local, not a symbol.

