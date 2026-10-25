package lang

import (
	"sort"
	"strings"

	"github.com/dVoo/reqmd/reqmd-code/internal/model"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

// HasAncestorKind reports whether any ancestor of node has one of the given
// tree-sitter node kinds. Every plugin uses it to reject declarations local
// to a function/class body rather than top-level symbols; variadic because
// some languages spell the same scope several ways (Go's function, method
// and func-literal).
func HasAncestorKind(node *sitter.Node, kinds ...string) bool {
	if node == nil {
		return false
	}
	for p := node.Parent(); p != nil; p = p.Parent() {
		for _, k := range kinds {
			if p.Kind() == k {
				return true
			}
		}
	}
	return false
}

// CapturesByName maps capture name to node for a single query match.
// Capture names are resolved through the query's CaptureNames() slice, so
// the caller fetches that once per query and passes it in.
func CapturesByName(m *sitter.QueryMatch, captureNames []string) map[string]*sitter.Node {
	captures := make(map[string]*sitter.Node, len(m.Captures))
	for i := range m.Captures {
		captures[captureNames[m.Captures[i].Index]] = &m.Captures[i].Node
	}
	return captures
}

// ModuleNameFromFile returns the file's basename without its extension,
// the default package/module name for languages whose package is the file.
func ModuleNameFromFile(file string) string {
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

// CaptureContent returns the textual content of a captured node, or the
// empty string when the node is nil. It exists as a shared helper
// because every language plugin needs the same nil-guarded access to
// tree-sitter node text; previously each plugin reimplemented this
// 3-liner.
func CaptureContent(node *sitter.Node, src []byte) string {
	if node == nil {
		return ""
	}
	return node.Utf8Text(src)
}

// SortByStartLine sorts symbols in place by StartLine ascending.
// Hoisted because every plugin's Parse returns symbols in source
// order but wants a defensive re-sort as the last step.
func SortByStartLine(s []model.Symbol) {
	sort.Slice(s, func(i, j int) bool { return s[i].StartLine < s[j].StartLine })
}

// SymbolAt fills in the span and doc fields for a single symbol. The
// language parameter is the interface (not the concrete plugin type)
// so this helper is reusable across plugins: each plugin's only
// dependency on its own type was the call to BindDoc, which is part
// of the Language contract.
func SymbolAt(node *sitter.Node, file, pkg string, kind model.SymbolKind, name, receiver string, l Language, src []byte) model.Symbol {
	return model.Symbol{
		File:      file,
		Package:   pkg,
		Kind:      kind,
		Name:      name,
		Receiver:  receiver,
		Doc:       l.BindDoc(node, src),
		StartLine: int(node.StartPosition().Row) + 1,
		EndLine:   int(node.EndPosition().Row) + 1,
	}
}
