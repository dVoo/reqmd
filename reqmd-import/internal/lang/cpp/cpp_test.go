package cpp

import (
	"testing"

	"reqmd-import/internal/lang"
	"reqmd-import/internal/model"
)

// TestCppLanguageRegisters verifies that CppLanguage self-registered
// itself with the global lang.Registry on package init.
func TestCppLanguageRegisters(t *testing.T) {
	l, ok := lang.Get("cpp")
	if !ok {
		t.Fatal("cpp language not registered")
	}
	if l.Name() != "cpp" {
		t.Errorf("expected name=cpp, got %q", l.Name())
	}
}

// TestParseSample verifies that Parse correctly identifies and classifies
// symbols in a representative C++ file: a class with an inline method and
// a declared method, a struct with a method, and a free function. The
// inline method must not be double-counted as a free function, and the
// method Receiver must carry the enclosing class name.
func TestParseSample(t *testing.T) {
	const src = `namespace calc {
// Calculator adds and subtracts.
class Calculator {
public:
    // add sums two ints.
    int add(int a, int b) {
        return a + b;
    }
    // sub declares a subtraction.
    int sub(int a, int b);
};
// Vec is a 2d vector.
struct Vec {
    int dot(Vec other);
};
// multiply multiplies two ints.
int multiply(int a, int b) {
    return a * b;
}
}
`

	l := &CppLanguage{}
	syms, err := l.Parse("calc.cpp", []byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(syms) != 6 {
		t.Fatalf("expected 6 symbols, got %d: %+v", len(syms), syms)
	}

	want := []struct {
		Kind     string
		Name     string
		Receiver string
		HasDoc   bool
	}{
		{"type", "Calculator", "", true},
		{"method", "add", "Calculator", true},
		{"method", "sub", "Calculator", true},
		{"type", "Vec", "", true},
		{"method", "dot", "Vec", false},
		{"function", "multiply", "", true},
	}
	for i, w := range want {
		got := syms[i]
		if string(got.Kind) != w.Kind {
			t.Errorf("sym[%d] kind = %q, want %q", i, got.Kind, w.Kind)
		}
		if got.Name != w.Name {
			t.Errorf("sym[%d] name = %q, want %q", i, got.Name, w.Name)
		}
		if got.Receiver != w.Receiver {
			t.Errorf("sym[%d] receiver = %q, want %q", i, got.Name, w.Receiver)
		}
		if w.HasDoc && got.Doc == "" {
			t.Errorf("sym[%d] %q: expected doc, got empty", i, w.Name)
		}
		if got.Package != "calc" {
			t.Errorf("sym[%d] package = %q, want calc", i, got.Package)
		}
		if got.StartLine < 1 || got.EndLine < got.StartLine {
			t.Errorf("sym[%d] bad span: %d..%d", i, got.StartLine, got.EndLine)
		}
	}
}

// TestParseTypedefDedup verifies that a `typedef struct Foo { ... } Foo;`
// emits exactly one type symbol (the typedef alias), not two.
func TestParseTypedefDedup(t *testing.T) {
	const src = `typedef struct Wrapper {
    int v;
} Wrapper;
`
	l := &CppLanguage{}
	syms, err := l.Parse("wrap.cpp", []byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(syms) != 1 {
		t.Fatalf("expected 1 symbol (deduped typedef), got %d: %+v", len(syms), syms)
	}
	if syms[0].Kind != model.SymbolType || syms[0].Name != "Wrapper" {
		t.Errorf("got %+v, want type Wrapper", syms[0])
	}
}

// TestParseAliasAndMacro verifies `using` aliases, enums, and macros are
// classified as type/const symbols.
func TestParseAliasAndMacro(t *testing.T) {
	const src = `#define LIMIT 100
using Count = int;
enum Status { OK, FAIL };
`
	l := &CppLanguage{}
	syms, err := l.Parse("alias.cpp", []byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(syms) != 3 {
		t.Fatalf("expected 3 symbols, got %d: %+v", len(syms), syms)
	}
	if syms[0].Kind != model.SymbolConst || syms[0].Name != "LIMIT" {
		t.Errorf("sym[0] = %+v, want const LIMIT", syms[0])
	}
	if syms[1].Kind != model.SymbolType || syms[1].Name != "Count" {
		t.Errorf("sym[1] = %+v, want type Count", syms[1])
	}
	if syms[2].Kind != model.SymbolType || syms[2].Name != "Status" {
		t.Errorf("sym[2] = %+v, want type Status", syms[2])
	}
}

// TestParseLocalDeclarationsSkipped verifies that declarations inside
// function bodies (local structs, local variables) are not emitted as
// top-level symbols.
func TestParseLocalDeclarationsSkipped(t *testing.T) {
	const src = `int global_var = 1;

static int first() {
    struct Inner { int v; };
    int local_var = 1;
    return local_var;
}

static int second() {
    int local_var = 2;
    return local_var;
}
`
	l := &CppLanguage{}
	syms, err := l.Parse("local.cpp", []byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	names := make(map[string]bool)
	for _, s := range syms {
		names[s.Name] = true
	}
	if names["local_var"] {
		t.Errorf("local var local_var must not be emitted")
	}
	if names["Inner"] {
		t.Errorf("local struct Inner must not be emitted")
	}
	if !names["global_var"] || !names["first"] || !names["second"] {
		t.Errorf("missing top-level symbols; got %v (len=%d)", names, len(syms))
	}
}

// TestParseEmpty handles zero-byte input gracefully.
func TestParseEmpty(t *testing.T) {
	l := &CppLanguage{}
	syms, err := l.Parse("empty.cpp", []byte(""))
	if err != nil {
		t.Fatalf("Parse empty: %v", err)
	}
	if len(syms) != 0 {
		t.Errorf("expected 0 symbols from empty input, got %d", len(syms))
	}
}

// TestCleanComment covers both `//` line comments and `/* ... */` block
// comments, including multi-line blocks.
func TestCleanComment(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"// hello", "hello"},
		{"// foo\n// bar", "foo\nbar"},
		{"/* hello */", "hello"},
		{"/* line1\n   line2 */", "line1\nline2"},
	}
	for _, c := range cases {
		if got := cleanComment(c.in); got != c.want {
			t.Errorf("cleanComment(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestBindDocIncludesDoc verifies that BindDoc attaches the full comment
// block (including a reqmd:trace marker) to the following declaration.
func TestBindDocIncludesDoc(t *testing.T) {
	const src = `// Computes the result.
// reqmd:trace REQ-MATH-001
int compute(void) {
    return 0;
}
`
	l := &CppLanguage{}
	syms, err := l.Parse("doc.cpp", []byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(syms) != 1 {
		t.Fatalf("expected 1 symbol, got %d", len(syms))
	}
	if got, want := syms[0].Doc, "Computes the result.\nreqmd:trace REQ-MATH-001"; got != want {
		t.Errorf("BindDoc = %q, want %q", got, want)
	}
}

// TestModuleNameFromFile covers the helper that derives a C++
// module/package name from a file path.
func TestModuleNameFromFile(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"foo.cpp", "foo"},
		{"path/to/bar.cc", "bar"},
		{"nested/dir.with.dot/qux.hpp", "qux"},
		{"", ""},
		{"noext", "noext"},
	}
	for _, c := range cases {
		if got := moduleNameFromFile(c.in); got != c.want {
			t.Errorf("moduleNameFromFile(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
