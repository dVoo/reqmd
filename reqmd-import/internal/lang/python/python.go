// Package python implements the Python language plugin for reqmd-import:
//
// it parses a Python source file via tree-sitter and emits a flat list of
// top-level symbols (functions, classes, methods, constants).
//
// # Plugin shape
//
// PythonLanguage implements internal/lang.Language. Each call to Parse
// creates a fresh tree-sitter parser and query cursor (cheap); the
// grammar and query are loaded from package-level state.
//
// # Symbol mapping
//
// The Python AST is queried via queries/symbols.scm and each match is
// mapped to a model.Symbol. Kind values used:
//
//	model.SymbolFunction - `def foo():` at module top level
//	model.SymbolMethod   - `def bar(self):` inside a class
//	model.SymbolType     - `class Foo:`
//	model.SymbolConst    - module-level `NAME = value` UPPER_CASE binding
//
// The Receiver field is populated for methods from the enclosing
// class_definition's `name` field. The Package field defaults to the
// file's basename without extension (Python's "package" == module == file);
// the extractor may rewrite it from the on-disk module path.
package python

import (
	"embed"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	pytree "github.com/smacker/go-tree-sitter/python"

	"reqmd-import/internal/lang"
	"reqmd-import/internal/model"
)

//go:embed queries/*.scm
var queriesFS embed.FS

// PythonLanguage is the Python language plugin. Implements lang.Language.
type PythonLanguage struct{}

// Compile-time assertion: PythonLanguage satisfies lang.Language.
var _ lang.Language = (*PythonLanguage)(nil)

// Name returns the canonical language identifier.
func (l *PythonLanguage) Name() string { return "python" }

// Extensions returns the file extensions this plugin handles.
func (l *PythonLanguage) Extensions() []string { return []string{".py"} }

// Grammar returns the tree-sitter-python grammar language.
func (l *PythonLanguage) Grammar() *sitter.Language { return pytree.GetLanguage() }

// Query returns the embedded tree-sitter query for symbol extraction.
// If the query file is missing (a packaging error), it returns an empty
// query rather than panicking; Parse will then emit zero symbols.
func (l *PythonLanguage) Query() string {
	data, err := queriesFS.ReadFile("queries/symbols.scm")
	if err != nil {
		return ""
	}
	return string(data)
}

// BindDoc returns the cleaned doc for the given declaration node. It
// joins the leading `#` line comments with the function/class
// docstring (if any) using a blank-line separator. The trace
// extractor regex is line-based, so the order of the two halves is
// immaterial. Returns the empty string when neither is present.
func (l *PythonLanguage) BindDoc(node *sitter.Node, src []byte) string {
	if node == nil {
		return ""
	}

	var parts []string
	prev := node.PrevSibling()
	// Walk backwards through comment nodes. Stop on the first
	// non-comment sibling, which is the boundary to the previous
	// declaration.
	for prev != nil {
		kind := prev.Type()
		if kind != "comment" {
			break
		}
		text := prev.Content(src)
		parts = append(parts, cleanPythonComment(text))
		prev = prev.PrevSibling()
	}

	// Reverse so the comments read top-to-bottom in source order.
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}
	comments := strings.Join(parts, "\n")

	doc := extractDocstring(node, src)

	if comments == "" {
		return doc
	}
	if doc == "" {
		return comments
	}
	return comments + "\n\n" + doc
}

// cleanPythonComment strips the leading `#` and trims trailing
// whitespace per line. Python's comment grammar produces a single
// node spanning one line (`# foo`), so this is a thin per-line pass.
func cleanPythonComment(text string) string {
	lines := strings.Split(text, "\n")
	for i, ln := range lines {
		ln = strings.TrimPrefix(ln, "#")
		// Drop the conventional single space after `#` (e.g. `# foo` → `foo`).
		ln = strings.TrimLeft(ln, " \t")
		lines[i] = strings.TrimRight(ln, " \t")
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

// extractDocstring returns the cleaned docstring of a
// function_definition or class_definition node, or "" when the node
// has no docstring. The tree-sitter-python shape is:
//
//	function_definition
//	  body: (block
//	    (expression_statement
//	      (string)))
//
// The string node's text includes the surrounding `"""`/`”'`
// delimiters (and an optional `r`/`b`/`u`/`rb`/... prefix). We strip
// the prefix and delimiters, then trim each line. Any unexpected AST
// shape returns "" rather than panicking.
func extractDocstring(node *sitter.Node, src []byte) string {
	if node == nil {
		return ""
	}
	body := node.ChildByFieldName("body")
	if body == nil || body.Type() != "block" {
		return ""
	}
	first := body.Child(0)
	if first == nil || first.Type() != "expression_statement" {
		return ""
	}
	// Walk expression_statement's children to find the string node.
	// The string is a direct child of the expression_statement.
	var stringNode *sitter.Node
	for i := 0; i < int(first.ChildCount()); i++ {
		c := first.Child(i)
		if c.Type() == "string" {
			stringNode = c
			break
		}
	}
	if stringNode == nil {
		return ""
	}
	raw := stringNode.Content(src)
	return stripDocstringDelimiters(raw)
}

// stripDocstringDelimiters removes the leading prefix letters (r, b,
// u, rb, br, f, rf, fr, in any case-insensitive combination) and
// the surrounding triple-quote delimiters (`"""` or `”'`) from a
// docstring literal, then trims each line. Returns "" if the input
// is too short to contain the delimiters.
func stripDocstringDelimiters(raw string) string {
	// Find the first quote character; everything before it is the
	// prefix (r, b, u, rb, br, f, rf, fr). Python 3.6+ also allows
	// the prefix to be any combination of r/R/b/B/u/U/f/F, in any
	// order, but for our purposes a simple "scan non-quote chars"
	// is sufficient.
	i := 0
	for i < len(raw) {
		c := raw[i]
		if c == '"' || c == '\'' {
			break
		}
		i++
	}
	prefix := raw[:i]
	body := raw[i:]

	// The body must start with a triple quote and end with the same
	// triple quote. Identify the quote char, drop the 3-char open
	// and 3-char close.
	if len(body) < 6 {
		return ""
	}
	quote := body[0]
	if body[0] != body[1] || body[1] != body[2] {
		return ""
	}
	if body[len(body)-1] != quote || body[len(body)-2] != quote || body[len(body)-3] != quote {
		return ""
	}
	inner := body[3 : len(body)-3]

	// `prefix` is kept only for documentation: the inner string has
	// already been extracted; we don't try to interpret escape
	// sequences (raw strings are already unescaped in the source
	// buffer tree-sitter gives us, so this is harmless for the
	// common case).
	_ = prefix

	// Trim each line: drop leading/trailing whitespace per row to
	// strip the indentation that Python's PEP 257 docstring
	// convention adds inside the triple quotes.
	lines := strings.Split(inner, "\n")
	for k, ln := range lines {
		lines[k] = strings.TrimSpace(ln)
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

// Parse parses the given Python source bytes and returns a flat list
// of top-level symbols, in source order. It uses the embedded
// tree-sitter query to find function, method, class, and module-level
// constant declarations, and walks the parent chain to drop class-body
// constants and to populate a method's Receiver from its enclosing
// class_definition's `name` field.
//
// # Dedup strategy
//
// The query produces intentional double-matches for the same source
// node: a method's `function_definition` matches both the bare
// `function_definition` pattern (@func.decl) and the class-body-anchored
// pattern (@method.decl); a decorated class matches both the bare
// `class_definition` pattern and the `decorated_definition` unwrap.
// To collapse these to a single emitted symbol we track emitted decl
// nodes by `*sitter.Node` identity. The pass is split into two phases
// so that @method.decl wins over @func.decl when both point at the
// same node (methods take precedence over plain functions).
func (l *PythonLanguage) Parse(file string, src []byte) ([]model.Symbol, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(l.Grammar())
	defer parser.Close()

	tree := parser.Parse(nil, src)
	if tree == nil {
		return nil, nil
	}
	defer tree.Close()

	root := tree.RootNode()

	q, err := sitter.NewQuery([]byte(l.Query()), l.Grammar())
	if err != nil {
		return nil, err
	}
	defer q.Close()

	qc := sitter.NewQueryCursor()
	defer qc.Close()
	qc.Exec(q, root)

	// Python has no package clause; fall back to the file's basename
	// without extension as the module/package name.
	pkgName := moduleNameFromFile(file)

	// First pass: collect all method.decl nodes so we can later drop
	// the bare @func.decl match that points at the same function_definition.
	methodNodes := make(map[*sitter.Node]struct{})
	for {
		m, ok := qc.NextMatch()
		if !ok {
			break
		}
		captures := capturesByName(m, q)
		if n := captures["method.decl"]; n != nil {
			methodNodes[n] = struct{}{}
		}
	}

	// Reset the cursor and walk the matches again, this time
	// dispatching. Tree-sitter QueryCursor is single-pass, so we
	// re-Exec rather than try to seek.
	qc2 := sitter.NewQueryCursor()
	defer qc2.Close()
	qc2.Exec(q, root)

	var out []model.Symbol
	seen := make(map[*sitter.Node]struct{})
	for {
		m, ok := qc2.NextMatch()
		if !ok {
			break
		}
		captures := capturesByName(m, q)

		// Skip the bare @func.decl when the same node is also a
		// method. Methods win.
		if fn := captures["func.decl"]; fn != nil {
			if _, isMethod := methodNodes[fn]; isMethod {
				continue
			}
		}

		// Class-level constants (`Cls.CONST = ...`) match the
		// @const.decl pattern at the expression_statement node. The
		// captured node is the expression_statement; walk parents
		// and reject if any ancestor is a class_definition.
		if cd := captures["const.decl"]; cd != nil && isInsideClassBody(cd) {
			continue
		}

		if sym, ok := symbolFromMatch(captures, file, pkgName, src, l); ok {
			// Final dedup: drop any symbol whose primary decl
			// node has already been emitted. Handles decorated
			// classes (same inner class_definition matched by
			// two patterns).
			primary := primaryDeclNode(captures)
			if primary == nil {
				continue
			}
			if _, dup := seen[primary]; dup {
				continue
			}
			seen[primary] = struct{}{}
			out = append(out, sym)
		}
	}

	// Source-order sort: the query already visits nodes in pre-order,
	// but be defensive in case the binding reorders.
	lang.SortByStartLine(out)
	return out, nil
}

// capturesByName returns a map from capture name to node for a single
// query match. Capture names are resolved via q.CaptureNameForId.
func capturesByName(m *sitter.QueryMatch, q *sitter.Query) map[string]*sitter.Node {
	captures := make(map[string]*sitter.Node, len(m.Captures))
	for i := range m.Captures {
		name := q.CaptureNameForId(m.Captures[i].Index)
		captures[name] = m.Captures[i].Node
	}
	return captures
}

// symbolFromMatch builds a model.Symbol from the per-match capture map.
// The map keys are the @-names from queries/symbols.scm. Returns false
// when the match has no usable captures.
func symbolFromMatch(captures map[string]*sitter.Node, file, pkg string, src []byte, l *PythonLanguage) (model.Symbol, bool) {
	switch {
	case captures["method.decl"] != nil:
		name := lang.CaptureContent(captures["method.name"], src)
		if name == "" {
			return model.Symbol{}, false
		}
		return lang.SymbolAt(captures["method.decl"], file, pkg, model.SymbolMethod, name, enclosingClassName(captures["method.decl"], src), l, src), true
	case captures["func.decl"] != nil:
		name := lang.CaptureContent(captures["func.name"], src)
		if name == "" {
			return model.Symbol{}, false
		}
		return lang.SymbolAt(captures["func.decl"], file, pkg, model.SymbolFunction, name, "", l, src), true
	case captures["type.decl"] != nil:
		name := lang.CaptureContent(captures["type.name"], src)
		if name == "" {
			return model.Symbol{}, false
		}
		return lang.SymbolAt(captures["type.decl"], file, pkg, model.SymbolType, name, "", l, src), true
	case captures["const.decl"] != nil:
		name := lang.CaptureContent(captures["const.name"], src)
		if name == "" {
			return model.Symbol{}, false
		}
		return lang.SymbolAt(captures["const.decl"], file, pkg, model.SymbolConst, name, "", l, src), true
	}
	return model.Symbol{}, false
}

// primaryDeclNode returns the captured "decl" node for the match's
// dispatch category, or nil when the match carries no primary decl.
func primaryDeclNode(captures map[string]*sitter.Node) *sitter.Node {
	switch {
	case captures["method.decl"] != nil:
		return captures["method.decl"]
	case captures["func.decl"] != nil:
		return captures["func.decl"]
	case captures["type.decl"] != nil:
		return captures["type.decl"]
	case captures["const.decl"] != nil:
		return captures["const.decl"]
	}
	return nil
}

// isInsideClassBody walks the parent chain of node and reports whether
// any ancestor is a `class_definition`. Used to reject class-level
// identifier assignments (`Cls.CONST = ...`) that otherwise match the
// @const.decl pattern. The captured @const.decl node is the
// `expression_statement`; climbing through `block` and `class_definition`
// is the only place where it appears (module-level constants sit under
// `module`/`block`).
func isInsideClassBody(node *sitter.Node) bool {
	if node == nil {
		return false
	}
	for p := node.Parent(); p != nil; p = p.Parent() {
		if p.Type() == "class_definition" {
			return true
		}
	}
	return false
}

// enclosingClassName walks up the parent chain from a method's
// function_definition node to the enclosing class_definition and
// returns the class's name text (e.g. "Calculator" for
// `class Calculator: def multiply(self, ...)`). Returns "" when the
// method has no enclosing class (defensive — the query should not
// produce such a match) or when the class has no name field (also
// defensive).
func enclosingClassName(methodNode *sitter.Node, src []byte) string {
	if methodNode == nil {
		return ""
	}
	for p := methodNode.Parent(); p != nil; p = p.Parent() {
		if p.Type() != "class_definition" {
			continue
		}
		name := p.ChildByFieldName("name")
		if name == nil {
			return ""
		}
		return name.Content(src)
	}
	return ""
}

// moduleNameFromFile returns the file's basename without its extension,
// used as Python's default package/module name when no explicit package
// metadata is available. Empty string is returned for empty input.
func moduleNameFromFile(file string) string {
	if file == "" {
		return ""
	}
	// Trim to the last path element.
	base := file
	if i := strings.LastIndexAny(base, "/\\"); i >= 0 {
		base = base[i+1:]
	}
	// Strip the extension; Python files always have one.
	if i := strings.LastIndex(base, "."); i >= 0 {
		base = base[:i]
	}
	return base
}

func init() {
	lang.Register(&PythonLanguage{})
}
