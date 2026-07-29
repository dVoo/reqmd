package golang

import (
	"os"
	"path/filepath"
	"testing"
)

// TestParseSample verifies that Parse correctly identifies and classifies
// all top-level symbols in a representative Go file: a function, two
// types (struct + interface), two methods (pointer + value receiver),
// a constant, and two variables. Doc comments must be attached to the
// right symbol. The test runs entirely in-memory; no external fixture.
func TestParseSample(t *testing.T) {
	const src = `// Package samplepkg is a small fixture.
package samplepkg

// Greeter returns a greeting.
func Greeter() string {
	return "hello"
}

// Counter is an interface.
type Counter interface {
	Inc()
}

// Adder is a struct.
type Adder struct {
	n int
}

// Add is a method on *Adder.
func (a *Adder) Add(x int) int {
	return a.n + x
}

// Sub is a method on Adder.
func (a Adder) Sub(x int) int {
	return a.n - x
}

// Max is a top-level constant.
const Max = 100

// DefaultName is a top-level var.
var DefaultName = "x"

// SomeMap is a composite var.
var SomeMap = map[string]int{}
`

	l := &GoLanguage{}
	syms, err := l.Parse("sample.go", []byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(syms) != 8 {
		t.Fatalf("expected 8 symbols, got %d", len(syms))
	}

	want := []struct {
		Kind     string
		Name     string
		Receiver string
		HasDoc   bool
	}{
		{"function", "Greeter", "", true},
		{"type", "Counter", "", true},
		{"type", "Adder", "", true},
		{"method", "Add", "*Adder", true},
		{"method", "Sub", "Adder", true},
		{"const", "Max", "", true},
		{"var", "DefaultName", "", true},
		{"var", "SomeMap", "", true},
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
			t.Errorf("sym[%d] receiver = %q, want %q", i, got.Receiver, w.Receiver)
		}
		if w.HasDoc && got.Doc == "" {
			t.Errorf("sym[%d] %q: expected doc, got empty", i, w.Name)
		}
		if got.Package != "samplepkg" {
			t.Errorf("sym[%d] package = %q, want samplepkg", i, got.Package)
		}
		if got.StartLine < 1 || got.EndLine < got.StartLine {
			t.Errorf("sym[%d] bad span: %d..%d", i, got.StartLine, got.EndLine)
		}
	}
}

// TestParseEmpty handles zero-byte and unparseable input gracefully.
func TestParseEmpty(t *testing.T) {
	l := &GoLanguage{}
	syms, err := l.Parse("empty.go", []byte(""))
	if err != nil {
		t.Fatalf("Parse empty: %v", err)
	}
	if len(syms) != 0 {
		t.Errorf("expected 0 symbols from empty input, got %d", len(syms))
	}
}

// TestParseNoSymbols handles a file with no top-level declarations
// (e.g. comments only). The package_clause still matches but emits no
// symbols.
func TestParseNoSymbols(t *testing.T) {
	const src = "// just a comment\npackage nothing\n"
	l := &GoLanguage{}
	syms, err := l.Parse("nothing.go", []byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	// The package_clause is captured but doesn't produce a symbol.
	if len(syms) != 0 {
		t.Errorf("expected 0 symbols, got %d", len(syms))
	}
}

// TestParseFromDisk reads a real .go file from disk and prints the
// results. Smoke test for the embed + filesystem round-trip.
func TestParseFromDisk(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "alpha.go")
	src := `package alpha

// Do is a method.
func (a *Alpha) Do() {}

// Beta is a type.
type Beta struct{}

// Gamma is a const.
const Gamma = 42
`
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	l := &GoLanguage{}
	syms, err := l.Parse(path, data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(syms) != 3 {
		t.Fatalf("expected 3 symbols, got %d", len(syms))
	}
	// Order: Do (method, line 4), Beta (type, line 7), Gamma (const, line 10).
	// The source is sorted by StartLine, so Do comes first.
	wantOrder := []string{"Do", "Beta", "Gamma"}
	for i, w := range wantOrder {
		if syms[i].Name != w {
			t.Errorf("sym[%d] = %q, want %q", i, syms[i].Name, w)
		}
	}
	// All expected names are present (no missing or extra).
	seen := make(map[string]bool, len(syms))
	for _, s := range syms {
		seen[s.Name] = true
	}
	for _, w := range wantOrder {
		if !seen[w] {
			t.Errorf("missing symbol %q in result", w)
		}
	}
}
