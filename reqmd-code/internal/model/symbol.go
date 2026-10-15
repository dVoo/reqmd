package model

import "strings"

// SymbolKind enumerates the top-level declarations a language plugin can
// extract. Each language plugin maps its AST node kinds to one of these
// values; the writer treats them uniformly.
type SymbolKind string

const (
	// SymbolFunction is a free function (Go: `func Foo()`, Python: `def foo():`).
	SymbolFunction SymbolKind = "function"
	// SymbolMethod is a function bound to a receiver/owner (Go: `func (r *Foo) Bar()`,
	// Python: `def bar(self):` inside a class, Rust: `impl Foo { fn bar() }`).
	// The Receiver field on the Symbol carries the owner type name.
	SymbolMethod SymbolKind = "method"
	// SymbolType is a type declaration (Go: `type Foo struct`, Rust: `struct Foo`).
	// Methods declared inside a type are emitted as SymbolMethod children of the
	// parent type requirement, not as siblings.
	SymbolType SymbolKind = "type"
	// SymbolConst is a constant declaration (Go: top-level `const`, Rust: `const FOO: u32`).
	SymbolConst SymbolKind = "const"
	// SymbolVar is a variable declaration (Go: top-level `var`, Rust: `static`).
	SymbolVar SymbolKind = "var"
)

// Symbol describes one top-level declaration extracted from a source file.
// It is produced by language plugins and consumed by the extractor's
// normalizer/tracer, which turn it into a GeneratedReq for the writer.
//
// Coordinates are 1-based for human readability (matching what editors show).
type Symbol struct {
	// File is the absolute or repo-relative path of the source file the
	// symbol was extracted from. Carried through into the .md output's
	// x-reqmd.source-file attribute for provenance.
	File string
	// Kind classifies the declaration. See SymbolKind constants.
	Kind SymbolKind
	// Name is the identifier as written in the source, e.g. "Foo" for
	// `func Foo()` or `*Foo.Bar` for a method on a pointer receiver.
	Name string
	// Receiver is the receiver type name for methods (e.g. "*Foo" or "Foo"
	// for Go pointer/value receivers, "self" for Python instance methods,
	// "Foo" for Rust impl blocks). Empty for non-method symbols.
	Receiver string
	// Package is the language-level package or module the symbol belongs to
	// (Go package path, Python module path, Rust crate name, etc.). The
	// writer uses this to bucket symbols into per-package .md files.
	Package string
	// Doc is the cleaned comment text immediately preceding the symbol's
	// declaration. May be empty if no doc comment is present.
	Doc string
	// StartLine is the 1-based line number of the symbol's first line.
	StartLine int
	// EndLine is the 1-based line number of the symbol's last line.
	EndLine int
}

// StripReceiverName normalizes a receiver string to its parent type name:
// strips an optional leading '*' (pointer receiver) and any surrounding
// brackets the AST may emit. Used to look up a method's parent type in
// the per-package ID map. Uses TrimPrefix/TrimSuffix, not TrimLeft with a
// cutset, so interior characters are not stripped.
func StripReceiverName(s string) string {
	s = strings.TrimPrefix(s, "*")
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")
	return s
}
