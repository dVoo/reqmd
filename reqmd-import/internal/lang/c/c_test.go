package clang

import (
	"os"
	"path/filepath"
	"testing"

	"reqmd-import/internal/lang"
	"reqmd-import/internal/model"
)

// TestCLanguageRegisters verifies that CLanguage self-registered itself
// with the global lang.Registry on package init.
func TestCLanguageRegisters(t *testing.T) {
	l, ok := lang.Get("c")
	if !ok {
		t.Fatal("c language not registered")
	}
	if l.Name() != "c" {
		t.Errorf("expected name=c, got %q", l.Name())
	}
}

// TestParseSample verifies that Parse correctly identifies and classifies
// all top-level symbols in a representative C file: a macro, a typedef'd
// struct, a standalone struct, an enum, a global variable, and two
// functions. Doc comments must be attached to the right symbol.
func TestParseSample(t *testing.T) {
	const src = `// header comment
#define MAX_SIZE 100

typedef struct Point {
    int x;
    int y;
} Point;

struct Counter {
    int n;
};

enum Color { RED, GREEN };

int global_counter = 5;

// add returns the sum.
int add(int a, int b) {
    return a + b;
}

// helper is internal.
static void helper(void) {}
`

	l := &CLanguage{}
	syms, err := l.Parse("sample.c", []byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(syms) != 7 {
		t.Fatalf("expected 7 symbols, got %d: %+v", len(syms), syms)
	}

	want := []struct {
		Kind   string
		Name   string
		HasDoc bool
	}{
		{"const", "MAX_SIZE", false},
		{"type", "Point", false},
		{"type", "Counter", false},
		{"type", "Color", false},
		{"var", "global_counter", false},
		{"function", "add", true},
		{"function", "helper", true},
	}
	for i, w := range want {
		got := syms[i]
		if string(got.Kind) != w.Kind {
			t.Errorf("sym[%d] kind = %q, want %q", i, got.Kind, w.Kind)
		}
		if got.Name != w.Name {
			t.Errorf("sym[%d] name = %q, want %q", i, got.Name, w.Name)
		}
		if w.HasDoc && got.Doc == "" {
			t.Errorf("sym[%d] %q: expected doc, got empty", i, w.Name)
		}
		if got.Package != "sample" {
			t.Errorf("sym[%d] package = %q, want sample", i, got.Package)
		}
		if got.StartLine < 1 || got.EndLine < got.StartLine {
			t.Errorf("sym[%d] bad span: %d..%d", i, got.StartLine, got.EndLine)
		}
	}
}

// TestParseTypedefDedup verifies that a `typedef struct Foo { ... } Foo;`
// emits exactly one type symbol (the typedef alias), not two: both the
// inner struct_specifier and the outer type_definition match the query.
func TestParseTypedefDedup(t *testing.T) {
	const src = `typedef struct Wrapper {
    int v;
} Wrapper;
`
	l := &CLanguage{}
	syms, err := l.Parse("wrap.c", []byte(src))
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

// TestParseMultiDeclarator verifies that `int a, b;` emits one symbol per
// declarator, not one per declaration.
func TestParseMultiDeclarator(t *testing.T) {
	const src = "int a, b;\n"
	l := &CLanguage{}
	syms, err := l.Parse("multi.c", []byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(syms) != 2 {
		t.Fatalf("expected 2 symbols, got %d: %+v", len(syms), syms)
	}
	if syms[0].Name != "a" || syms[1].Name != "b" {
		t.Errorf("got names %q and %q, want a and b", syms[0].Name, syms[1].Name)
	}
}

// TestParsePrototypeSkipped verifies that a function prototype without a
// body (`int foo(int);`) is not emitted as a symbol. The definition
// (with a body) is emitted exactly once.
func TestParsePrototypeSkipped(t *testing.T) {
	const src = `int foo(int x);
int foo(int x) {
    return x;
}
`
	l := &CLanguage{}
	syms, err := l.Parse("proto.c", []byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(syms) != 1 {
		t.Fatalf("expected 1 symbol (definition only), got %d: %+v", len(syms), syms)
	}
	if syms[0].Kind != model.SymbolFunction || syms[0].Name != "foo" {
		t.Errorf("got %+v, want function foo", syms[0])
	}
}

// TestParseLocalDeclarationsSkipped verifies that declarations inside
// function bodies (local structs, local variables) are not emitted as
// top-level symbols. Two same-named locals in different functions would
// otherwise collide on the same requirement ID.
func TestParseLocalDeclarationsSkipped(t *testing.T) {
	const src = `int global_var = 1;

static int first(void) {
    struct Inner { int v; };
    int local_var = 1;
    return local_var;
}

static int second(void) {
    struct Inner { int v; };
    int local_var = 2;
    return local_var;
}
`
	l := &CLanguage{}
	syms, err := l.Parse("local.c", []byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	names := make(map[string]bool)
	for _, s := range syms {
		names[s.Name] = true
	}
	if len(names) != 3 {
		t.Errorf("expected only top-level symbols (global_var, first, second), got %d: %+v", len(syms), syms)
	}
	if names["local_var"] {
		t.Errorf("local var local_var must not be emitted")
	}
	if names["Inner"] {
		t.Errorf("local struct Inner must not be emitted")
	}
	if !names["global_var"] || !names["first"] || !names["second"] {
		t.Errorf("missing top-level symbols; got %v", names)
	}
}

// TestParseEmpty handles zero-byte input gracefully.
func TestParseEmpty(t *testing.T) {
	l := &CLanguage{}
	syms, err := l.Parse("empty.c", []byte(""))
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
		{"// x   ", "x"},
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
	l := &CLanguage{}
	syms, err := l.Parse("doc.c", []byte(src))
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

// TestParseFromDisk reads a real .c file from disk and verifies the
// embed + filesystem round-trip.
func TestParseFromDisk(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "alpha.c")
	src := `// Do adds.
int do_add(int a) { return a; }

// Beta is a type.
struct Beta { int v; };

// Gamma is a macro.
#define Gamma 42
`
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	l := &CLanguage{}
	syms, err := l.Parse(path, data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(syms) != 3 {
		t.Fatalf("expected 3 symbols, got %d: %+v", len(syms), syms)
	}
}

// TestModuleNameFromFile covers the helper that derives a C
// module/package name from a file path.
func TestModuleNameFromFile(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"foo.c", "foo"},
		{"path/to/bar.c", "bar"},
		{"nested/dir.with.dot/qux.c", "qux"},
		{"", ""},
		{"noext", "noext"},
	}
	for _, c := range cases {
		if got := moduleNameFromFile(c.in); got != c.want {
			t.Errorf("moduleNameFromFile(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
