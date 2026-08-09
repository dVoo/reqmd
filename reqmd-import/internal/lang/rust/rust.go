// Package rust implements the Rust language plugin for reqmd-import:
//
// it parses a Rust source file via tree-sitter and emits a flat list of
// top-level symbols (functions, methods, types, constants, statics).
//
// # Plugin shape
//
// RustLanguage implements internal/lang.Language. Each call to Parse
// creates a fresh tree-sitter parser and query cursor (cheap); the
// grammar and query are loaded from package-level state.
//
// # Symbol mapping
//
// The Rust AST is queried via queries/symbols.scm and each match is
// mapped to a model.Symbol. Kind values used:
//
//	model.SymbolFunction - free `fn foo(...)` items at module level
//	model.SymbolMethod   - `fn` items inside an `impl` block and trait
//	                       signatures/defaults inside a `trait`; the
//	                       Receiver field carries the impl'd type or
//	                       trait name
//	model.SymbolType     - `struct`, `enum`, `trait`, `union`, and
//	                       `type Foo = ...;` alias items
//	model.SymbolConst    - `const FOO: T = ...` items
//	model.SymbolVar      - `static FOO: T = ...` items
//
// The Package field defaults to the file's basename without extension
// (Rust's "package" == module == file); the extractor may rewrite it
// from the on-disk module path.
package rust

import (
	"embed"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_rust "github.com/tree-sitter/tree-sitter-rust/bindings/go"

	"reqmd-import/internal/lang"
	"reqmd-import/internal/model"
)

//go:embed queries/*.scm
var queriesFS embed.FS

// RustLanguage is the Rust language plugin. Implements lang.Language.
type RustLanguage struct{}

// Compile-time assertion: RustLanguage satisfies lang.Language.
var _ lang.Language = (*RustLanguage)(nil)

// Name returns the canonical language identifier.
func (l *RustLanguage) Name() string { return "rust" }

// Extensions returns the file extensions this plugin handles.
func (l *RustLanguage) Extensions() []string { return []string{".rs"} }

// Grammar returns the tree-sitter-rust grammar language.
func (l *RustLanguage) Grammar() *sitter.Language { return sitter.NewLanguage(tree_sitter_rust.Language()) }

// Query returns the embedded tree-sitter query for symbol extraction.
// If the query file is missing (a packaging error), it returns an empty
// query rather than panicking; Parse will then emit zero symbols.
func (l *RustLanguage) Query() string {
	data, err := queriesFS.ReadFile("queries/symbols.scm")
	if err != nil {
		return ""
	}
	return string(data)
}

// BindDoc returns the cleaned doc comment for the given declaration
// node. It walks backwards over any contiguous run of `//` line
// comments and `/* ... */` block comments, joins them with newlines,
// and strips comment markers (including `///` and `//!` doc-comment
// forms). Returns the empty string when no preceding comment is present.
func (l *RustLanguage) BindDoc(node *sitter.Node, src []byte) string {
	if node == nil {
		return ""
	}

	var parts []string
	prev := node.PrevSibling()
	// Walk backwards through comment nodes. Stop on the first
	// non-comment sibling, which is the boundary to the previous
	// declaration.
	for prev != nil {
		kind := prev.Kind()
		if kind != "line_comment" && kind != "block_comment" {
			break
		}
		parts = append(parts, cleanComment(prev.Utf8Text(src)))
		prev = prev.PrevSibling()
	}

	// Reverse so the comments read top-to-bottom in source order.
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}
	return strings.Join(parts, "\n")
}

// cleanComment strips Rust comment markers from a comment node's text
// and trims per-line whitespace. It accepts both `//`-style (including
// doc forms `///` and `//!`) and `/* ... */`-style (including `/**`
// and `/*!`) comments. The conventional single space after the marker
// (`// foo` → `foo`) is dropped.
func cleanComment(text string) string {
	lines := strings.Split(text, "\n")
	if strings.HasPrefix(text, "/*") {
		// /* ... */ block: drop the delimiters, keep interior lines.
		// Doc forms `/**` and `/*!` carry an extra marker character
		// that is stripped alongside the opening `/*`.
		if len(lines) > 0 {
			lines[0] = strings.TrimPrefix(lines[0], "/**")
			lines[0] = strings.TrimPrefix(lines[0], "/*!")
			lines[0] = strings.TrimPrefix(lines[0], "/*")
		}
		if len(lines) > 0 {
			lines[len(lines)-1] = strings.TrimSuffix(lines[len(lines)-1], "*/")
		}
		for i, ln := range lines {
			lines[i] = strings.TrimSpace(ln)
		}
	} else {
		// `//` line comment. Try doc-comment markers first (///, //!)
		// so their longer prefixes win over the plain `//` strip.
		for i, ln := range lines {
			for _, marker := range []string{"///", "//!"} {
				ln = strings.TrimPrefix(ln, marker)
			}
			ln = strings.TrimPrefix(ln, "//")
			lines[i] = strings.TrimSpace(ln)
		}
	}
	// Strip blank lines at the start and end for tidy output.
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	return strings.Join(lines, "\n")
}

// Parse parses the given Rust source bytes and returns a flat list of
// top-level symbols, in source order. It uses the embedded tree-sitter
// query to find function, method, type, const, and static declarations,
// and walks the parent chain to drop impl/trait-body function items from
// the free-function pass (those are methods) and to populate a method's
// Receiver from its enclosing impl/trait.
//
// # Dedup strategy
//
// The query produces intentional double-matches for the same source
// node: a method's `function_item` matches both the bare
// `function_item` pattern (@func.decl) and the impl-body-anchored
// pattern (@method.decl). To collapse these to a single emitted symbol
// we track emitted method decl nodes by `*sitter.Node` identity, then
// skip the bare @func.decl match that points at the same node.
func (l *RustLanguage) Parse(file string, src []byte) ([]model.Symbol, error) {
	parser := sitter.NewParser()
	defer parser.Close()
	if err := parser.SetLanguage(l.Grammar()); err != nil {
		return nil, err
	}

	tree := parser.Parse(src, nil)
	if tree == nil {
		return nil, nil
	}
	defer tree.Close()

	root := tree.RootNode()

	q, qerr := sitter.NewQuery(l.Grammar(), l.Query())
	if qerr != nil {
		return nil, qerr
	}
	defer q.Close()

	// Rust has no package clause; fall back to the file's basename
	// without extension as the module/package name.
	pkgName := moduleNameFromFile(file)
	captureNames := q.CaptureNames()

	// First pass: collect all method.decl nodes so we can later drop the
	// bare @func.decl match that points at the same function_item.
	methodNodes := make(map[uintptr]struct{})
	qc := sitter.NewQueryCursor()
	matches := qc.Matches(q, root, src)
	for {
		m := matches.Next()
		if m == nil {
			break
		}
		captures := capturesByName(m, captureNames)
		if n := captures["method.decl"]; n != nil {
			methodNodes[n.Id()] = struct{}{}
		}
	}
	qc.Close()

	// Reset the cursor and walk the matches again, this time
	// dispatching. Tree-sitter QueryCursor is single-pass, so we
	// re-Exec rather than try to seek.
	qc2 := sitter.NewQueryCursor()
	defer qc2.Close()
	matches2 := qc2.Matches(q, root, src)

	var out []model.Symbol
	for {
		m := matches2.Next()
		if m == nil {
			break
		}
		captures := capturesByName(m, captureNames)

		// Skip the bare @func.decl when the same node is also a method.
		// Methods win.
		if fn := captures["func.decl"]; fn != nil {
			if _, isMethod := methodNodes[fn.Id()]; isMethod {
				continue
			}
		}

		if sym, ok := symbolFromMatch(captures, file, pkgName, src, l); ok {
			out = append(out, sym)
		}
	}

	// Source-order sort: the query already visits nodes in pre-order,
	// but be defensive in case the binding reorders.
	lang.SortByStartLine(out)
	return out, nil
}

// capturesByName returns a map from capture name to node for a single
// query match. Capture names are resolved via the pre-fetched
// captureNames slice from q.CaptureNames().
func capturesByName(m *sitter.QueryMatch, captureNames []string) map[string]*sitter.Node {
	captures := make(map[string]*sitter.Node, len(m.Captures))
	for i := range m.Captures {
		name := captureNames[m.Captures[i].Index]
		captures[name] = &m.Captures[i].Node
	}
	return captures
}

// symbolFromMatch builds a model.Symbol from the per-match capture map.
// The map keys are the @-names from queries/symbols.scm. Returns false
// when the match has no usable captures or refers to a function-local
// item.
func symbolFromMatch(captures map[string]*sitter.Node, file, pkg string, src []byte, l *RustLanguage) (model.Symbol, bool) {
	switch {
	case captures["method.decl"] != nil:
		name := lang.CaptureContent(captures["method.name"], src)
		if name == "" {
			return model.Symbol{}, false
		}
		return lang.SymbolAt(captures["method.decl"], file, pkg, model.SymbolMethod, name, enclosingTypeName(captures["method.decl"], src), l, src), true
	case captures["func.decl"] != nil:
		name := lang.CaptureContent(captures["func.name"], src)
		if name == "" {
			return model.Symbol{}, false
		}
		return lang.SymbolAt(captures["func.decl"], file, pkg, model.SymbolFunction, name, "", l, src), true
	case captures["type.decl"] != nil:
		// Local types (`struct Foo { }` inside a function body) are
		// function-scoped, not top-level symbols.
		if isInsideFunction(captures["type.decl"]) {
			return model.Symbol{}, false
		}
		name := lang.CaptureContent(captures["type.name"], src)
		if name == "" {
			return model.Symbol{}, false
		}
		return lang.SymbolAt(captures["type.decl"], file, pkg, model.SymbolType, name, "", l, src), true
	case captures["const.decl"] != nil:
		// Rust allows `const` items inside function bodies; those are
		// function-scoped, not top-level symbols.
		if isInsideFunction(captures["const.decl"]) {
			return model.Symbol{}, false
		}
		name := lang.CaptureContent(captures["const.name"], src)
		if name == "" {
			return model.Symbol{}, false
		}
		return lang.SymbolAt(captures["const.decl"], file, pkg, model.SymbolConst, name, "", l, src), true
	case captures["var.decl"] != nil:
		if isInsideFunction(captures["var.decl"]) {
			return model.Symbol{}, false
		}
		name := lang.CaptureContent(captures["var.name"], src)
		if name == "" {
			return model.Symbol{}, false
		}
		return lang.SymbolAt(captures["var.decl"], file, pkg, model.SymbolVar, name, "", l, src), true
	}
	return model.Symbol{}, false
}

// isInsideFunction reports whether node's ancestor chain contains a
// function_item (which always nests its body in a `block`). Used to
// reject const/static/type items declared inside function bodies, which
// are function-scoped and not top-level symbols.
func isInsideFunction(node *sitter.Node) bool {
	if node == nil {
		return false
	}
	for p := node.Parent(); p != nil; p = p.Parent() {
		if p.Kind() == "function_item" {
			return true
		}
	}
	return false
}

// enclosingTypeName walks up the parent chain from a method node to the
// enclosing impl_item or trait_item and returns the owner's type name
// (e.g. "Point" for `impl Point { fn scale(&self) }` and "Shape" for
// `trait Shape { fn area(&self) }`). Generic and scoped impl types
// (`impl Point<T>`, `impl crate::Point`) are reduced to their base
// type identifier so the receiver matches the parent type's name.
// Returns "" when the method has no enclosing impl/trait.
func enclosingTypeName(methodNode *sitter.Node, src []byte) string {
	if methodNode == nil {
		return ""
	}
	for p := methodNode.Parent(); p != nil; p = p.Parent() {
		switch p.Kind() {
		case "trait_item":
			name := p.ChildByFieldName("name")
			if name == nil {
				return ""
			}
			return name.Utf8Text(src)
		case "impl_item":
			typ := p.ChildByFieldName("type")
			if typ == nil {
				return ""
			}
			// Walk down to the base type identifier, skipping generic
			// arguments and scope prefixes (`impl crate::Foo<T>` → Foo).
			for n := typ; n != nil; {
				if n.Kind() == "type_identifier" {
					return n.Utf8Text(src)
				}
				if n.Kind() == "scoped_type_identifier" {
					n = n.ChildByFieldName("name")
					continue
				}
				var next *sitter.Node
				for i := uint(0); i < n.ChildCount(); i++ {
					if n.Child(i).IsNamed() {
						next = n.Child(i)
						break
					}
				}
				if next == nil {
					break
				}
				n = next
			}
			return typ.Utf8Text(src)
		}
	}
	return ""
}

// moduleNameFromFile returns the file's basename without its extension,
// used as Rust's default package/module name when no explicit crate
// metadata is available. Empty string is returned for empty input.
func moduleNameFromFile(file string) string {
	if file == "" {
		return ""
	}
	base := file
	if i := strings.LastIndexAny(base, "/\\"); i >= 0 {
		base = base[i+1:]
	}
	if i := strings.LastIndex(base, "."); i >= 0 {
		base = base[:i]
	}
	return base
}

func init() {
	lang.Register(&RustLanguage{})
}
