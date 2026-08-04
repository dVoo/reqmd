package python

import (
	"testing"

	sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"

	"reqmd-import/internal/lang"
	"reqmd-import/internal/model"
)

// TestPythonLanguageRegisters verifies that PythonLanguage self-registered
// itself with the global lang.Registry on package init, and that its
// Name/Extensions surface the expected values.
func TestPythonLanguageRegisters(t *testing.T) {
	l, ok := lang.Get("python")
	if !ok {
		t.Fatal("python language not registered")
	}
	if l.Name() != "python" {
		t.Errorf("expected name=python, got %q", l.Name())
	}
	if len(l.Extensions()) == 0 || l.Extensions()[0] != ".py" {
		t.Errorf("expected .py extension, got %v", l.Extensions())
	}
}

// TestPythonGrammarLoads verifies the tree-sitter-python binding is
// reachable and returns a non-nil grammar. This is a smoke test —
// the binding is cgo and a nil return would indicate a broken FFI
// link or a missing grammar dynamic library.
func TestPythonGrammarLoads(t *testing.T) {
	l := &PythonLanguage{}
	g := l.Grammar()
	if g == nil {
		t.Fatal("grammar is nil")
	}
}

// TestPythonQuery verifies the embedded .scm query file is present and
// readable. Returns true if the query text is non-empty (Fixer B's
// contribution); the skeleton ships a placeholder so this may pass
// with a one-line comment.
func TestPythonQuery(t *testing.T) {
	l := &PythonLanguage{}
	q := l.Query()
	if q == "" {
		t.Error("Query() returned empty string (missing embedded .scm)")
	}
}

// TestModuleNameFromFile covers the helper that derives a Python
// module/package name from a file path. It is used as the default
// value for Symbol.Package in the absence of explicit metadata.
func TestModuleNameFromFile(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"foo.py", "foo"},
		{"path/to/bar.py", "bar"},
		{"/abs/path/baz.py", "baz"},
		{"nested/dir.with.dot/qux.py", "qux"},
		{"", ""},
		{"noext", "noext"},
	}
	for _, c := range cases {
		if got := moduleNameFromFile(c.in); got != c.want {
			t.Errorf("moduleNameFromFile(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestCleanPythonComment verifies the per-line # prefix stripping
// used by BindDoc for module-level # comments.
func TestCleanPythonComment(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"# hello", "hello"},
		{"# foo\n# bar", "foo\nbar"},
		{"#", ""},
		{"# x   ", "x"},
	}
	for _, c := range cases {
		if got := cleanPythonComment(c.in); got != c.want {
			t.Errorf("cleanPythonComment(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestBindDocNilNode is a defensive guard: passing a nil node must
// return the empty string rather than panicking.
func TestBindDocNilNode(t *testing.T) {
	l := &PythonLanguage{}
	if got := l.BindDoc(nil, []byte("# x\ndef f(): pass")); got != "" {
		t.Errorf("BindDoc(nil) = %q, want empty", got)
	}
}

// TestParseSimple exercises the full Parse pipeline against a small
// Python snippet with one top-level function, one class, and one
// method. Verifies symbol count, kinds, and the method's Receiver
// (populated from the enclosing class_definition's name).
func TestParseSimple(t *testing.T) {
	src := []byte(`def add(a, b):
    """Add two integers."""
    return a + b


class Calculator:
    def multiply(self, a, b):
        return a * b
`)

	pl := &PythonLanguage{}
	syms, err := pl.Parse("arith.py", src)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(syms) != 3 {
		t.Fatalf("got %d symbols, want 3: %+v", len(syms), syms)
	}

	// Build a name-indexed view for stable assertions regardless of
	// match-iteration order.
	byName := make(map[string]model.Symbol, len(syms))
	for _, s := range syms {
		byName[s.Name] = s
	}

	if s, ok := byName["add"]; !ok {
		t.Errorf("missing top-level function 'add'")
	} else if s.Kind != model.SymbolFunction {
		t.Errorf("add.Kind = %q, want %q", s.Kind, model.SymbolFunction)
	} else if s.Receiver != "" {
		t.Errorf("add.Receiver = %q, want empty", s.Receiver)
	} else if s.Package != "arith" {
		t.Errorf("add.Package = %q, want %q", s.Package, "arith")
	} else if s.File != "arith.py" {
		t.Errorf("add.File = %q, want %q", s.File, "arith.py")
	}

	if s, ok := byName["Calculator"]; !ok {
		t.Errorf("missing class 'Calculator'")
	} else if s.Kind != model.SymbolType {
		t.Errorf("Calculator.Kind = %q, want %q", s.Kind, model.SymbolType)
	}

	if s, ok := byName["multiply"]; !ok {
		t.Errorf("missing method 'multiply'")
	} else if s.Kind != model.SymbolMethod {
		t.Errorf("multiply.Kind = %q, want %q", s.Kind, model.SymbolMethod)
	} else if s.Receiver != "Calculator" {
		t.Errorf("multiply.Receiver = %q, want %q", s.Receiver, "Calculator")
	}
}

// TestParseClassLevelConstRejected verifies that a class-level
// identifier assignment (`Cls.CONST = 99`) does NOT emit a
// module-level constant symbol. The query's @const.decl pattern
// matches the expression_statement regardless of lexical context; the
// Go-side parent-chain walk must drop class-body matches.
func TestParseClassLevelConstRejected(t *testing.T) {
	src := []byte(`class Calculator:
    VERSION = 99


MODULE_LEVEL = 1
`)

	pl := &PythonLanguage{}
	syms, err := pl.Parse("cfg.py", src)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(syms) != 2 {
		t.Fatalf("got %d symbols, want 2 (class + module const): %+v", len(syms), syms)
	}
	for _, s := range syms {
		if s.Name == "VERSION" {
			t.Errorf("class-level const VERSION must not be emitted")
		}
	}
	// MODULE_LEVEL should be present as a const.
	var foundConst bool
	for _, s := range syms {
		if s.Name == "MODULE_LEVEL" && s.Kind == model.SymbolConst {
			foundConst = true
		}
	}
	if !foundConst {
		t.Errorf("module-level const MODULE_LEVEL missing or wrong kind: %+v", syms)
	}
}

// TestParseDecoratedClassDedup verifies that a decorated class
// (`@decorator\nclass Foo: ...`) produces exactly one type symbol,
// not two. The query matches both the bare `class_definition` pattern
// and the `decorated_definition` unwrap; the node-identity dedup must
// collapse them.
func TestParseDecoratedClassDedup(t *testing.T) {
	src := []byte(`@decorator
class Foo:
    pass
`)

	pl := &PythonLanguage{}
	syms, err := pl.Parse("deco.py", src)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(syms) != 1 {
		t.Fatalf("got %d symbols, want 1 (decorated class deduplicated): %+v", len(syms), syms)
	}
	if syms[0].Kind != model.SymbolType || syms[0].Name != "Foo" {
		t.Errorf("got %+v, want type Foo", syms[0])
	}
}

// parseFirstFuncDef is a test helper that parses src and returns
// the first top-level function_definition node plus a cleanup func
// that must be called to release the underlying tree-sitter tree.
// Tests need direct AST access to call extractDocstring/BindDoc on
// a specific node; the public Parse path is exercised elsewhere.
func parseFirstFuncDef(t *testing.T, src []byte) (*sitter.Node, func()) {
	t.Helper()
	parser := sitter.NewParser()
	if err := parser.SetLanguage(sitter.NewLanguage(tree_sitter_python.Language())); err != nil {
		parser.Close()
		t.Fatalf("SetLanguage: %v", err)
	}
	tree := parser.Parse(src, nil)
	if tree == nil {
		parser.Close()
		t.Fatal("tree is nil")
	}
	root := tree.RootNode()
	// Walk top-level children looking for the first function_definition.
	n := root.ChildCount()
	for i := uint(0); i < n; i++ {
		c := root.Child(i)
		if c.Kind() == "function_definition" {
			return c, func() {
				tree.Close()
				parser.Close()
			}
		}
	}
	tree.Close()
	parser.Close()
	t.Fatal("no function_definition found in source")
	return nil, func() {}
}

// TestExtractDocstringTripleQuote verifies that a single-line
// `"""..."""` docstring is captured with delimiters stripped and
// per-line whitespace trimmed.
func TestExtractDocstringTripleQuote(t *testing.T) {
	src := []byte(`def f():
    """Hello world."""
    pass
`)
	fn, cleanup := parseFirstFuncDef(t, src)
	defer cleanup()
	got := extractDocstring(fn, src)
	want := "Hello world."
	if got != want {
		t.Errorf("extractDocstring = %q, want %q", got, want)
	}
}

// TestExtractDocstringMultiLine verifies that a multi-line
// `"""..."""` docstring is captured with delimiters stripped, the
// leading newline trimmed, and per-line whitespace trimmed.
func TestExtractDocstringMultiLine(t *testing.T) {
	src := []byte(`def f():
    """
    Multi-line
    docstring.
    """
    pass
`)
	fn, cleanup := parseFirstFuncDef(t, src)
	defer cleanup()
	got := extractDocstring(fn, src)
	want := "Multi-line\ndocstring."
	if got != want {
		t.Errorf("extractDocstring = %q, want %q", got, want)
	}
}

// TestExtractDocstringSingleTripleQuote verifies that `”'...”'`
// (single-quote triple) docstrings are also captured.
func TestExtractDocstringSingleTripleQuote(t *testing.T) {
	src := []byte(`def f():
    '''Single-quote docstring.'''
    pass
`)
	fn, cleanup := parseFirstFuncDef(t, src)
	defer cleanup()
	got := extractDocstring(fn, src)
	want := "Single-quote docstring."
	if got != want {
		t.Errorf("extractDocstring = %q, want %q", got, want)
	}
}

// TestExtractDocstringNoDocstring verifies that a function with no
// docstring returns the empty string.
func TestExtractDocstringNoDocstring(t *testing.T) {
	src := []byte(`def f():
    pass
`)
	fn, cleanup := parseFirstFuncDef(t, src)
	defer cleanup()
	if got := extractDocstring(fn, src); got != "" {
		t.Errorf("extractDocstring = %q, want empty", got)
	}
}

// TestExtractDocstringRawPrefix verifies that raw-string (`r"""..."""`)
// docstrings have the `r` prefix handled gracefully. Tree-sitter's
// source buffer includes the prefix, and our scanner must skip it
// before locating the triple-quote delimiters.
func TestExtractDocstringRawPrefix(t *testing.T) {
	src := []byte(`def f():
    r"""Raw docstring with \n unescaped."""
    pass
`)
	fn, cleanup := parseFirstFuncDef(t, src)
	defer cleanup()
	got := extractDocstring(fn, src)
	want := `Raw docstring with \n unescaped.`
	if got != want {
		t.Errorf("extractDocstring = %q, want %q", got, want)
	}
}

// TestBindDocIncludesDocstring verifies that BindDoc returns BOTH
// the leading `#` comments AND the docstring, joined with a blank
// line. The trace extractor regex is line-based, so explicit
// `reqmd:trace` markers in the leading comments and bare-ID
// references in the docstring are both picked up.
func TestBindDocIncludesDocstring(t *testing.T) {
	src := []byte(`# module comment
# reqmd:trace REQ-AUTH-001
def f():
    """A docstring referencing REQ-AUTH-002."""
    pass
`)
	fn, cleanup := parseFirstFuncDef(t, src)
	defer cleanup()
	pl := &PythonLanguage{}
	got := pl.BindDoc(fn, src)
	want := "module comment\nreqmd:trace REQ-AUTH-001\n\nA docstring referencing REQ-AUTH-002."
	if got != want {
		t.Errorf("BindDoc = %q, want %q", got, want)
	}
}

// TestBindDocDocstringOnly verifies that BindDoc returns the
// docstring when there are no leading `#` comments. This is the
// case that previously returned "" and caused the `multiply`
// heuristic-trace test to fail.
func TestBindDocDocstringOnly(t *testing.T) {
	src := []byte(`def f():
    """See REQ-ARITH-002 for details."""
    pass
`)
	fn, cleanup := parseFirstFuncDef(t, src)
	defer cleanup()
	pl := &PythonLanguage{}
	got := pl.BindDoc(fn, src)
	want := "See REQ-ARITH-002 for details."
	if got != want {
		t.Errorf("BindDoc = %q, want %q", got, want)
	}
}
