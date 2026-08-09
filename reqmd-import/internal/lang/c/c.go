// Package clang implements the C language plugin for reqmd-import:
//
// it parses a C source file via tree-sitter and emits a flat list of
// top-level symbols (functions, types, globals, macros).
//
// # Plugin shape
//
// CLanguage implements internal/lang.Language. Each call to Parse
// creates a fresh tree-sitter parser and query cursor (cheap); the
// grammar and query are loaded from package-level state.
//
// # Symbol mapping
//
// The C AST is queried via queries/symbols.scm and each match is
// mapped to a model.Symbol. Kind values used:
//
//	model.SymbolFunction - top-level `int foo(...) { ... }` definitions
//	model.SymbolType     - `struct Foo { }`, `enum Foo { }`, `union Foo { }`,
//	                       and `typedef ... Foo;` aliases
//	model.SymbolVar      - top-level variable declarations (`int x = 5;`)
//	model.SymbolConst    - object-like `#define FOO 100` macros
//
// C has no method concept, so model.SymbolMethod is never emitted.
// Function prototypes (`int foo(void);` without a body) are deliberately
// skipped: a prototype plus its definition would otherwise yield two
// requirements with the same ID in the same package.
//
// The Package field defaults to the file's basename without extension
// (C's translation unit is the closest analogue to a "module"); the
// extractor may rewrite it from the on-disk module path.
package clang

import (
	"embed"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_c "github.com/tree-sitter/tree-sitter-c/bindings/go"

	"reqmd-import/internal/lang"
	"reqmd-import/internal/model"
)

//go:embed queries/*.scm
var queriesFS embed.FS

// CLanguage is the C language plugin. Implements lang.Language.
type CLanguage struct{}

// Compile-time assertion: CLanguage satisfies lang.Language.
var _ lang.Language = (*CLanguage)(nil)

// Name returns the canonical language identifier.
func (l *CLanguage) Name() string { return "c" }

// Extensions returns the file extensions this plugin handles.
func (l *CLanguage) Extensions() []string { return []string{".c", ".h"} }

// Grammar returns the tree-sitter-c grammar language.
func (l *CLanguage) Grammar() *sitter.Language { return sitter.NewLanguage(tree_sitter_c.Language()) }

// Query returns the embedded tree-sitter query for symbol extraction.
// If the query file is missing (a packaging error), it returns an empty
// query rather than panicking; Parse will then emit zero symbols.
func (l *CLanguage) Query() string {
	data, err := queriesFS.ReadFile("queries/symbols.scm")
	if err != nil {
		return ""
	}
	return string(data)
}

// BindDoc returns the cleaned doc comment for the given declaration
// node. It walks backwards over any contiguous run of `//` line
// comments and `/* ... */` block comments, joins them with newlines,
// and strips comment markers. Returns the empty string when no
// preceding comment is present.
func (l *CLanguage) BindDoc(node *sitter.Node, src []byte) string {
	if node == nil {
		return ""
	}

	var parts []string
	prev := node.PrevSibling()
	// Walk backwards through comment nodes. Stop on the first
	// non-comment sibling, which is the boundary to the previous
	// declaration.
	for prev != nil {
		if prev.Kind() != "comment" {
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

// cleanComment strips C comment markers from a comment node's text and
// trims per-line whitespace. Both `//` line comments and `/* ... */`
// block comments arrive as a single tree-sitter `comment` node; the
// leading delimiter distinguishes the two styles. The conventional
// single space after the marker (`// foo` → `foo`) is dropped, as is
// per-line indentation inside block comments.
func cleanComment(text string) string {
	lines := strings.Split(text, "\n")
	if strings.HasPrefix(text, "/*") {
		// /* ... */ block: drop the delimiters, keep interior lines.
		if len(lines) > 0 {
			lines[0] = strings.TrimPrefix(lines[0], "/*")
		}
		if len(lines) > 0 {
			lines[len(lines)-1] = strings.TrimSuffix(lines[len(lines)-1], "*/")
		}
		for i, ln := range lines {
			lines[i] = strings.TrimSpace(ln)
		}
	} else {
		// `//` line comment (may be multi-line via trailing `\`).
		for i, ln := range lines {
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

// Parse parses the given C source bytes and returns a flat list of
// top-level symbols, in source order. It uses the embedded tree-sitter
// query to find function, type, variable, and macro declarations, and
// drops type specifiers that are merely the inner half of a typedef
// (those are emitted once, via the typedef itself).
func (l *CLanguage) Parse(file string, src []byte) ([]model.Symbol, error) {
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

	// C has no package clause; fall back to the file's basename without
	// extension as the module/package name.
	pkgName := moduleNameFromFile(file)
	captureNames := q.CaptureNames()

	var out []model.Symbol
	qc := sitter.NewQueryCursor()
	defer qc.Close()
	matches := qc.Matches(q, root, src)
	for {
		m := matches.Next()
		if m == nil {
			break
		}
		captures := capturesByName(m, captureNames)

		// A typedef `typedef struct Foo { ... } Foo;` matches both the
		// inner struct_specifier and the outer type_definition, yielding
		// two @type.decl captures for the same name. The specifier is
		// only meaningful as a standalone declaration, so drop any type
		// specifier whose parent chain runs through a type_definition.
		if cd := captures["type.decl"]; cd != nil && isInsideTypeDefinition(cd) {
			continue
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
// declaration.
func symbolFromMatch(captures map[string]*sitter.Node, file, pkg string, src []byte, l *CLanguage) (model.Symbol, bool) {
	switch {
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
	case captures["var.decl"] != nil:
		// Local variables inside function bodies share the same ID space
		// as package-level ones; drop them so two same-named locals in
		// different functions don't collide on one requirement ID.
		if isInsideFunction(captures["var.decl"]) {
			return model.Symbol{}, false
		}
		name := lang.CaptureContent(captures["var.name"], src)
		if name == "" {
			return model.Symbol{}, false
		}
		return lang.SymbolAt(captures["var.decl"], file, pkg, model.SymbolVar, name, "", l, src), true
	case captures["const.decl"] != nil:
		name := lang.CaptureContent(captures["const.name"], src)
		if name == "" {
			return model.Symbol{}, false
		}
		return lang.SymbolAt(captures["const.decl"], file, pkg, model.SymbolConst, name, "", l, src), true
	}
	return model.Symbol{}, false
}

// isInsideFunction reports whether node's ancestor chain contains a
// function_definition. Used to reject declarations nested inside
// function bodies (local structs, local variables), which are not
// top-level symbols.
func isInsideFunction(node *sitter.Node) bool {
	if node == nil {
		return false
	}
	for p := node.Parent(); p != nil; p = p.Parent() {
		if p.Kind() == "function_definition" {
			return true
		}
	}
	return false
}

// isInsideTypeDefinition reports whether node's ancestor chain contains
// a `type_definition` (i.e. the node is the inner struct/enum/union
// specifier of a typedef). Such specifiers are skipped in favor of the
// enclosing type_definition, which carries the typedef'd alias name.
func isInsideTypeDefinition(node *sitter.Node) bool {
	if node == nil {
		return false
	}
	for p := node.Parent(); p != nil; p = p.Parent() {
		if p.Kind() == "type_definition" {
			return true
		}
	}
	return false
}

// moduleNameFromFile returns the file's basename without its extension,
// used as C's default package/module name when no explicit package
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
	lang.Register(&CLanguage{})
}
