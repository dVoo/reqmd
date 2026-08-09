package golang

import (
	"embed"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_go "github.com/tree-sitter/tree-sitter-go/bindings/go"

	"reqmd-import/internal/lang"
	"reqmd-import/internal/model"
)

//go:embed queries/*.scm
var queriesFS embed.FS

// GoLanguage is the Go language plugin. Implements lang.Language.
type GoLanguage struct{}

// Compile-time assertion: GoLanguage satisfies lang.Language.
var _ lang.Language = (*GoLanguage)(nil)

// Name returns the canonical language identifier.
func (l *GoLanguage) Name() string { return "go" }

// Extensions returns the file extensions this plugin handles.
func (l *GoLanguage) Extensions() []string { return []string{".go"} }

// Grammar returns the tree-sitter-go grammar language.
func (l *GoLanguage) Grammar() *sitter.Language { return sitter.NewLanguage(tree_sitter_go.Language()) }

// Query returns the embedded tree-sitter query for symbol extraction.
// If the query file is missing (a packaging error), it returns an empty
// query rather than panicking; Parse will then emit zero symbols.
func (l *GoLanguage) Query() string {
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
func (l *GoLanguage) BindDoc(node *sitter.Node, src []byte) string {
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
		if kind != "comment" && kind != "block_comment" {
			break
		}
		text := prev.Utf8Text(src)
		parts = append(parts, cleanComment(text, kind))
		prev = prev.PrevSibling()
	}

	// Reverse so the comments read top-to-bottom in source order.
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}
	return strings.Join(parts, "\n")
}

// cleanComment strips comment markers and trims per-line trailing
// whitespace. `///` godoc lines are reduced to `//` form so callers
// don't need to know about that distinction.
func cleanComment(text, kind string) string {
	lines := strings.Split(text, "\n")
	switch kind {
	case "comment":
		for i, ln := range lines {
			ln = strings.TrimPrefix(ln, "///")
			ln = strings.TrimPrefix(ln, "//")
			lines[i] = strings.TrimRight(ln, " \t")
		}
	case "block_comment":
		// /* ... */: drop the delimiters, keep interior lines.
		if len(lines) > 0 {
			lines[0] = strings.TrimPrefix(lines[0], "/*")
		}
		if len(lines) > 0 {
			last := lines[len(lines)-1]
			lines[len(lines)-1] = strings.TrimSuffix(last, "*/")
		}
		for i, ln := range lines {
			lines[i] = strings.TrimRight(ln, " \t")
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

// Parse parses the given Go source bytes and returns a flat list of
// top-level symbols, in source order. It uses the embedded tree-sitter
// query to find function, method, type, const, and var declarations,
// and the package_clause to populate Symbol.Package.
//
// Two passes over the query results: the first collects the package
// name (if any), the second builds the symbol list. Symbols with no
// detectable package get Package = "" — callers can fall back to the
// file's parent directory if needed.
func (l *GoLanguage) Parse(file string, src []byte) ([]model.Symbol, error) {
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

	pkgName := ""
	var out []model.Symbol

	qc := sitter.NewQueryCursor()
	defer qc.Close()
	matches := qc.Matches(q, root, src)
	captureNames := q.CaptureNames()
	for {
		m := matches.Next()
		if m == nil {
			break
		}
		// Build a name->node map for this match.
		captures := make(map[string]*sitter.Node, len(m.Captures))
		for i := range m.Captures {
			name := captureNames[m.Captures[i].Index]
			captures[name] = &m.Captures[i].Node
		}

		// Package clause: capture the package name once and remember it.
		if pkgNode := captures["pkg.name"]; pkgNode != nil && pkgName == "" {
			text := pkgNode.Utf8Text(src)
			pkgName = text
			continue
		}
		if captures["pkg.decl"] != nil {
			// The pkg.decl match was the package_clause itself;
			// the pkg.name capture inside it is what carries
			// the identifier. Skip the outer decl.
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

// symbolFromMatch builds a model.Symbol from the per-match capture map.
// The map keys are the @-names from queries/symbols.scm. Returns false
// when the match has no usable captures.
func symbolFromMatch(captures map[string]*sitter.Node, file, pkg string, src []byte, l *GoLanguage) (model.Symbol, bool) {
	switch {
	case captures["func.decl"] != nil:
		name := lang.CaptureContent(captures["func.name"], src)
		if name == "" {
			return model.Symbol{}, false
		}
		return lang.SymbolAt(captures["func.decl"], file, pkg, model.SymbolFunction, name, "", l, src), true
	case captures["method.decl"] != nil:
		name := lang.CaptureContent(captures["method.name"], src)
		if name == "" {
			return model.Symbol{}, false
		}
		return lang.SymbolAt(captures["method.decl"], file, pkg, model.SymbolMethod, name, receiverType(captures["method.decl"], src), l, src), true
	case captures["type.decl"] != nil:
		name := lang.CaptureContent(captures["type.name"], src)
		if name == "" {
			return model.Symbol{}, false
		}
		return lang.SymbolAt(captures["type.decl"], file, pkg, model.SymbolType, name, "", l, src), true
	case captures["const.decl"] != nil:
		// Skip const declarations nested inside function bodies
		// (e.g. `const src = ...` in a test helper) — only package-level
		// constants are top-level symbols. Also skip the blank identifier
		// `var _ T = ...` compile-time assertion idiom.
		if isInsideFunction(captures["const.decl"]) {
			return model.Symbol{}, false
		}
		name := lang.CaptureContent(captures["const.name"], src)
		if name == "" || name == "_" {
			return model.Symbol{}, false
		}
		return lang.SymbolAt(captures["const.decl"], file, pkg, model.SymbolConst, name, "", l, src), true
	case captures["var.decl"] != nil:
		// Same scope rule as const: only package-level variables are
		// top-level symbols.
		if isInsideFunction(captures["var.decl"]) {
			return model.Symbol{}, false
		}
		name := lang.CaptureContent(captures["var.name"], src)
		if name == "" || name == "_" {
			return model.Symbol{}, false
		}
		return lang.SymbolAt(captures["var.decl"], file, pkg, model.SymbolVar, name, "", l, src), true
	}
	return model.Symbol{}, false
}

// isInsideFunction reports whether node's ancestor chain contains a
// function_declaration or method_declaration. Used to reject const/var
// declarations nested inside function bodies, which are function-local
// and not top-level symbols.
func isInsideFunction(node *sitter.Node) bool {
	if node == nil {
		return false
	}
	for p := node.Parent(); p != nil; p = p.Parent() {
		switch p.Kind() {
		case "function_declaration", "method_declaration", "func_literal":
			return true
		}
	}
	return false
}

// receiverType extracts the receiver type name from a method_declaration
// node via the `receiver` field. The receiver is a `parameter_list`
// whose first `parameter_declaration` holds `<name> <type>`; we surface
// the type text (e.g. `*Foo` or `Foo`).
func receiverType(methodNode *sitter.Node, src []byte) string {
	if methodNode == nil {
		return ""
	}
	recvList := methodNode.ChildByFieldName("receiver")
	if recvList == nil {
		return ""
	}
	n := recvList.ChildCount()
	for i := uint(0); i < n; i++ {
		c := recvList.Child(i)
		if c.Kind() != "parameter_declaration" {
			continue
		}
		typeNode := c.ChildByFieldName("type")
		if typeNode != nil {
			text := typeNode.Utf8Text(src)
			return text
		}
	}
	return ""
}

func init() {
	lang.Register(&GoLanguage{})
}
