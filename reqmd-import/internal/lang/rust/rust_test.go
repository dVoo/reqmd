package rust

import (
	"testing"

	"reqmd-import/internal/lang"
	"reqmd-import/internal/model"
)

// TestRustLanguageRegisters verifies that RustLanguage self-registered
// itself with the global lang.Registry on package init.
func TestRustLanguageRegisters(t *testing.T) {
	l, ok := lang.Get("rust")
	if !ok {
		t.Fatal("rust language not registered")
	}
	if l.Name() != "rust" {
		t.Errorf("expected name=rust, got %q", l.Name())
	}
}

// TestParseSample verifies that Parse correctly identifies and classifies
// all top-level symbols in a representative Rust file: a const, a struct,
// an enum, a free function, a trait with a method signature, a static,
// and an impl block method. The impl method must not be double-counted as
// a free function, and its Receiver must be the impl'd type.
func TestParseSample(t *testing.T) {
	const src = `//! crate docs

/// Maximum buffer size.
const MAX: u32 = 100;

/// A 2d point.
struct Point {
    x: i32,
    y: i32,
}

/// Color palette.
enum Color {
    Red,
    Green,
}

/// Adds two numbers.
fn add(a: i32, b: i32) -> i32 {
    a + b
}

impl Point {
    /// Scales the point.
    fn scale(&self, f: i32) -> i32 {
        self.x * f
    }
}

/// Shape abstraction.
trait Shape {
    /// Computes the area.
    fn area(&self) -> i32;
}

/// Version string.
static VERSION: &str = "1.0";
`

	l := &RustLanguage{}
	syms, err := l.Parse("geom.rs", []byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(syms) != 8 {
		t.Fatalf("expected 8 symbols, got %d: %+v", len(syms), syms)
	}

	want := []struct {
		Kind     string
		Name     string
		Receiver string
		HasDoc   bool
	}{
		{"const", "MAX", "", true},
		{"type", "Point", "", true},
		{"type", "Color", "", true},
		{"function", "add", "", true},
		{"method", "scale", "Point", true},
		{"type", "Shape", "", true},
		{"method", "area", "Shape", false},
		{"var", "VERSION", "", true},
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
			t.Errorf("sym[%d] %q receiver = %q, want %q", i, got.Name, got.Receiver, w.Receiver)
		}
		if w.HasDoc && got.Doc == "" {
			t.Errorf("sym[%d] %q: expected doc, got empty", i, w.Name)
		}
		if got.Package != "geom" {
			t.Errorf("sym[%d] package = %q, want geom", i, got.Package)
		}
		if got.StartLine < 1 || got.EndLine < got.StartLine {
			t.Errorf("sym[%d] bad span: %d..%d", i, got.StartLine, got.EndLine)
		}
	}
}

// TestParseTraitDefaultMethod verifies that a trait method with a default
// body (function_item, not function_signature_item) is still a method.
func TestParseTraitDefaultMethod(t *testing.T) {
	const src = `trait Greeter {
    fn hello(&self) -> &str {
        "hi"
    }
}
`
	l := &RustLanguage{}
	syms, err := l.Parse("greet.rs", []byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(syms) != 2 {
		t.Fatalf("expected 2 symbols, got %d: %+v", len(syms), syms)
	}
	if syms[0].Kind != model.SymbolType || syms[0].Name != "Greeter" {
		t.Errorf("sym[0] = %+v, want type Greeter", syms[0])
	}
	if syms[1].Kind != model.SymbolMethod || syms[1].Name != "hello" || syms[1].Receiver != "Greeter" {
		t.Errorf("sym[1] = %+v, want method hello on Greeter", syms[1])
	}
}

// TestParseGenericImpl verifies that the receiver of a method on a generic
// impl (`impl Point<T>`) resolves to the base type identifier.
func TestParseGenericImpl(t *testing.T) {
	const src = `struct Cache<T> { v: T }

impl<T> Cache<T> {
    fn get(&self) -> &T {
        &self.v
    }
}
`
	l := &RustLanguage{}
	syms, err := l.Parse("cache.rs", []byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(syms) != 2 {
		t.Fatalf("expected 2 symbols, got %d: %+v", len(syms), syms)
	}
	if syms[1].Kind != model.SymbolMethod || syms[1].Receiver != "Cache" {
		t.Errorf("sym[1] = %+v, want method get on Cache", syms[1])
	}
}

// TestParseLocalItemsSkipped verifies that items declared inside
// function bodies (local consts, local structs) are not emitted as
// top-level symbols.
func TestParseLocalItemsSkipped(t *testing.T) {
	const src = `const GLOBAL: u32 = 1;

fn a() {
    const LOCAL: u32 = 5;
    struct Inner { v: u32 }
    let x = 1;
}

fn b() {
    const LOCAL: u32 = 6;
}
`
	l := &RustLanguage{}
	syms, err := l.Parse("local.rs", []byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	names := make(map[string]bool)
	for _, s := range syms {
		names[s.Name] = true
	}
	if names["LOCAL"] {
		t.Errorf("local const LOCAL must not be emitted")
	}
	if names["Inner"] {
		t.Errorf("local struct Inner must not be emitted")
	}
	if !names["GLOBAL"] || !names["a"] || !names["b"] {
		t.Errorf("missing top-level symbols; got %v (len=%d)", names, len(syms))
	}
}

// TestParseEmpty handles zero-byte input gracefully.
func TestParseEmpty(t *testing.T) {
	l := &RustLanguage{}
	syms, err := l.Parse("empty.rs", []byte(""))
	if err != nil {
		t.Fatalf("Parse empty: %v", err)
	}
	if len(syms) != 0 {
		t.Errorf("expected 0 symbols from empty input, got %d", len(syms))
	}
}

// TestCleanComment covers `//`, `///`, `//!`, `/* */`, `/** */` and
// `/*! */` comment forms.
func TestCleanComment(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"// hello", "hello"},
		{"/// doc line", "doc line"},
		{"//! inner doc", "inner doc"},
		{"// foo\n// bar", "foo\nbar"},
		{"/* hello */", "hello"},
		{"/** block doc */", "block doc"},
		{"/*! inner block */", "inner block"},
	}
	for _, c := range cases {
		if got := cleanComment(c.in); got != c.want {
			t.Errorf("cleanComment(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestBindDocIncludesDoc verifies that BindDoc attaches the full doc
// comment block (including a reqmd:trace marker) to the following item.
func TestBindDocIncludesDoc(t *testing.T) {
	const src = `/// Computes the result.
/// reqmd:trace REQ-MATH-001
fn compute() -> i32 {
    0
}
`
	l := &RustLanguage{}
	syms, err := l.Parse("doc.rs", []byte(src))
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

// TestModuleNameFromFile covers the helper that derives a Rust
// module/package name from a file path.
func TestModuleNameFromFile(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"foo.rs", "foo"},
		{"path/to/bar.rs", "bar"},
		{"nested/dir.with.dot/qux.rs", "qux"},
		{"", ""},
		{"noext", "noext"},
	}
	for _, c := range cases {
		if got := moduleNameFromFile(c.in); got != c.want {
			t.Errorf("moduleNameFromFile(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
